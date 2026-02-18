// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/boyter/cs/pkg/common"
	"github.com/boyter/cs/pkg/search"
	"github.com/boyter/cs/pkg/snippet"
	"github.com/boyter/gocodewalker"
)

// SearchStats holds counters readable after the search channel drains.
type SearchStats struct {
	FileCount     atomic.Int64
	TextFileCount atomic.Int64
}

// DoSearch runs the search pipeline and returns a channel of matched FileJob results
// plus stats that are populated as the search runs.
// If cache is non-nil, it will attempt to use cached file locations from a previous
// prefix query instead of walking the filesystem, and will store results for future use.
func DoSearch(ctx context.Context, cfg *Config, query string, cache *SearchCache) (<-chan *common.FileJob, *SearchStats) {
	out := make(chan *common.FileJob, runtime.NumCPU())
	stats := &SearchStats{}

	// Parse query into AST
	lexer := search.NewLexer(strings.NewReader(query))
	parser := search.NewParser(lexer)
	ast, _ := parser.ParseQuery()
	if ast == nil {
		close(out)
		return out, stats
	}
	transformer := &search.Transformer{}
	ast, _ = transformer.TransformAST(ast)
	ast = search.PlanAST(ast)

	// Resolve language types to extensions
	if len(cfg.LanguageTypes) > 0 {
		langExts := languageExtensions(cfg.LanguageTypes)
		cfg.AllowListExtensions = append(cfg.AllowListExtensions, langExts...)
	}

	// Determine walk directory
	dir := "."
	if strings.TrimSpace(cfg.Directory) != "" {
		dir = cfg.Directory
	}
	if cfg.FindRoot {
		dir = gocodewalker.FindRepositoryRoot(dir)
	}

	fileQueue := make(chan *gocodewalker.File, 1000)

	// Try cache hit path: feed cached file locations instead of walking
	var walkerToTerminate *gocodewalker.FileWalker
	if cache != nil {
		if cachedFiles, ok := cache.FindPrefixFiles(cfg.AllowListExtensions, query); ok {
			go func() {
				defer close(fileQueue)
				for _, loc := range cachedFiles {
					select {
					case <-ctx.Done():
						return
					case fileQueue <- &gocodewalker.File{
						Location: loc,
						Filename: filepath.Base(loc),
					}:
					}
				}
			}()
			goto startWorkers
		}
	}

	// Set up file walker (cache miss or no cache)
	{
		walker := gocodewalker.NewFileWalker(dir, fileQueue)
		walker.AllowListExtensions = cfg.AllowListExtensions
		walker.IgnoreIgnoreFile = cfg.IgnoreIgnoreFile
		walker.IgnoreGitIgnore = cfg.IgnoreGitIgnore
		walker.LocationExcludePattern = cfg.LocationExcludePattern
		walker.IncludeHidden = cfg.IncludeHidden
		walker.ExcludeDirectory = cfg.PathDenylist
		walkerToTerminate = walker

		go func() { _ = walker.Start() }()
	}

startWorkers:
	// Ensure walker is terminated on context cancellation
	searchDone := make(chan struct{})
	if walkerToTerminate != nil {
		walker := walkerToTerminate
		go func() {
			select {
			case <-ctx.Done():
				walker.Terminate()
			case <-searchDone:
			}
		}()
	}

	// Track matched file locations for cache population
	var matchedMu sync.Mutex
	var matchedLocations []string

	// Fan out workers to read and search files in parallel
	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range fileQueue {
				select {
				case <-ctx.Done():
					return
				default:
				}

				stats.FileCount.Add(1)

				// Read file content with max size limit
				content, err := readFileContent(f.Location, cfg.MaxReadSizeBytes)
				if err != nil || len(content) == 0 {
					continue
				}

				// Binary check: look for NUL byte in first 10KB
				if !cfg.IncludeBinaryFiles {
					check := content
					if len(check) > 10_000 {
						check = content[:10_000]
					}
					if bytes.IndexByte(check, 0) != -1 {
						continue
					}
				}

				// Minified check
				if !cfg.IncludeMinified {
					lines := bytes.Split(content, []byte("\n"))
					sumLineLength := 0
					for _, s := range lines {
						sumLineLength += len(s)
					}
					avgLineLength := sumLineLength / len(lines)
					if avgLineLength > cfg.MinifiedLineByteLength {
						continue
					}
				}

				stats.TextFileCount.Add(1)

				// Evaluate query AST against file content
				matched, matchLocations := search.EvaluateFile(ast, content, f.Filename, cfg.CaseSensitive)
				if !matched {
					continue
				}

				// Track matched file location for cache
				if cache != nil {
					matchedMu.Lock()
					matchedLocations = append(matchedLocations, f.Location)
					matchedMu.Unlock()
				}

				snippet.AddPhraseMatchLocations(content, strings.Trim(query, "\""), matchLocations)

				lang, sccLines, sccCode, sccComment, sccBlank, sccComplexity, contentByteType := fileCodeStats(f.Filename, content)

				fj := &common.FileJob{
					Filename:        f.Filename,
					Extension:       gocodewalker.GetExtension(f.Filename),
					Location:        f.Location,
					Content:         content,
					ContentByteType: contentByteType,
					Bytes:           len(content),
					MatchLocations:  matchLocations,
					Language:        lang,
					Lines:           sccLines,
					Code:            sccCode,
					Comment:         sccComment,
					Blank:           sccBlank,
					Complexity:      sccComplexity,
				}

				select {
				case out <- fj:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
		close(searchDone)

		// Populate cache with matched file locations
		if cache != nil && len(matchedLocations) > 0 {
			cache.Store(cfg.AllowListExtensions, query, matchedLocations)
		}
	}()

	return out, stats
}

// readFileContent reads a file, limiting to maxBytes if the file is larger.
func readFileContent(location string, maxBytes int64) ([]byte, error) {
	fi, err := os.Lstat(location)
	if err != nil {
		return nil, err
	}

	if fi.Size() < maxBytes {
		return os.ReadFile(location)
	}

	fil, err := os.Open(location)
	if err != nil {
		return nil, err
	}
	defer fil.Close()

	buf := make([]byte, maxBytes)
	n, err := fil.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
