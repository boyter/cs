// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	str "github.com/boyter/go-string"
	"github.com/boyter/gocodewalker"

	"github.com/boyter/cs/pkg/common"
	"github.com/boyter/cs/pkg/ranker"
	"github.com/boyter/cs/pkg/snippet"
)

// HTTP template types — prefixed with "http" to avoid collision with TUI's searchResult.
type httpSearch struct {
	SearchTerm          string             `json:"searchTerm"`
	SnippetSize         int                `json:"snippetSize"`
	Results             []httpSearchResult `json:"results"`
	ResultsCount        int                `json:"resultsCount"`
	RuntimeMilliseconds int64              `json:"runtimeMilliseconds"`
	ProcessedFileCount  int64              `json:"processedFileCount"`
	ExtensionFacet      []httpFacetResult  `json:"extensionFacet,omitempty"`
	Pages               []httpPageResult   `json:"pages,omitempty"`
	Ext                 string             `json:"ext,omitempty"`
	Ranker              string             `json:"ranker"`
	CodeFilter          string             `json:"codeFilter"`
	Gravity             string             `json:"gravity"`
	Noise               string             `json:"noise"`
}

type httpLineResult struct {
	LineNumber int           `json:"lineNumber"`
	Content    template.HTML `json:"content"`
	Gap        bool          `json:"gap,omitempty"`
}

type httpSearchResult struct {
	Title              string           `json:"title"`
	Location           string           `json:"location"`
	Content            []template.HTML  `json:"content,omitempty"`
	StartPos           int              `json:"startPos"`
	EndPos             int              `json:"endPos"`
	Score              float64          `json:"score"`
	IsLineMode         bool             `json:"isLineMode,omitempty"`
	LineResults        []httpLineResult `json:"lineResults,omitempty"`
	Language           string           `json:"language,omitempty"`
	Lines              int64            `json:"lines,omitempty"`
	Code               int64            `json:"code,omitempty"`
	Comment            int64            `json:"comment,omitempty"`
	Blank              int64            `json:"blank,omitempty"`
	Complexity         int64            `json:"complexity,omitempty"`
	DuplicateCount     int              `json:"duplicateCount,omitempty"`
	DuplicateLocations []string         `json:"duplicateLocations,omitempty"`
}

type httpFileDisplay struct {
	Location            string
	Content             template.HTML
	RuntimeMilliseconds int64
	Language            string
	Lines               int64
	Code                int64
	Comment             int64
	Blank               int64
	Complexity          int64
}

type httpFacetResult struct {
	Title       string `json:"title"`
	Count       int    `json:"count"`
	SearchTerm  string `json:"searchTerm"`
	SnippetSize int    `json:"snippetSize"`
	Ranker      string `json:"ranker"`
	CodeFilter  string `json:"codeFilter"`
	Gravity     string `json:"gravity"`
	Noise       string `json:"noise"`
}

type httpPageResult struct {
	SearchTerm  string `json:"searchTerm"`
	SnippetSize int    `json:"snippetSize"`
	Value       int    `json:"value"`
	Name        string `json:"name"`
	Ext         string `json:"ext,omitempty"`
	Ranker      string `json:"ranker"`
	CodeFilter  string `json:"codeFilter"`
	Gravity     string `json:"gravity"`
	Noise       string `json:"noise"`
}

func StartHttpServer(cfg *Config) {
	cache := NewSearchCache()
	searchTmpl, err := resolveSearchTemplate(cfg)
	if err != nil {
		log.Fatalf("failed to load search template: %v", err)
	}
	displayTmpl, err := resolveDisplayTemplate(cfg)
	if err != nil {
		log.Fatalf("failed to load display template: %v", err)
	}

	http.HandleFunc("/file/raw/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Replace(r.URL.Path, "/file/raw/", "", 1)
		w.Header().Set("Content-Type", "text/plain")
		http.ServeFile(w, r, path)
	})

	http.HandleFunc("/file/", func(w http.ResponseWriter, r *http.Request) {
		startTime := makeTimestampMilli()
		startPos := tryParseInt(r.URL.Query().Get("sp"), 0)
		endPos := tryParseInt(r.URL.Query().Get("ep"), 0)

		path := strings.Replace(r.URL.Path, "/file/", "", 1)

		if strings.TrimSpace(cfg.Directory) != "" {
			path = "/" + path
		}

		content, err := os.ReadFile(path)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		// Create a random str to define where the start and end of
		// our highlight should be which we swap out later after we have
		// HTML escaped everything
		md5Digest := md5.New()
		fmtBegin := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("begin_%d", makeTimestampNano()))))
		fmtEnd := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("end_%d", makeTimestampNano()))))

		coloredContent := str.HighlightString(string(content), [][]int{{startPos, endPos}}, fmtBegin, fmtEnd)
		coloredContent = html.EscapeString(coloredContent)
		coloredContent = strings.Replace(coloredContent, fmtBegin, fmt.Sprintf(`<strong id="%d">`, startPos), -1)
		coloredContent = strings.Replace(coloredContent, fmtEnd, "</strong>", -1)

		lang, sccLines, sccCode, sccComment, sccBlank, sccComplexity, _ := fileCodeStats(filepath.Base(path), content)

		err = displayTmpl.Execute(w, httpFileDisplay{
			Location:            path,
			Content:             template.HTML(coloredContent),
			RuntimeMilliseconds: makeTimestampMilli() - startTime,
			Language:            lang,
			Lines:               sccLines,
			Code:                sccCode,
			Comment:             sccComment,
			Blank:               sccBlank,
			Complexity:          sccComplexity,
		})
		if err != nil {
			log.Printf("template execute error: %v", err)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := makeTimestampMilli()
		query := r.URL.Query().Get("q")
		snippetLength := tryParseInt(r.URL.Query().Get("ss"), 300)
		ext := r.URL.Query().Get("ext")
		page := tryParseInt(r.URL.Query().Get("p"), 0)
		pageSize := 20

		// Parse ranking/filter params with defaults from server config
		rankerParam := r.URL.Query().Get("rk")
		if rankerParam == "" {
			rankerParam = cfg.Ranker
		}
		codeFilter := r.URL.Query().Get("cf")
		if codeFilter == "" {
			codeFilter = "default"
		}
		gravityParam := r.URL.Query().Get("gv")
		if gravityParam == "" {
			gravityParam = cfg.GravityIntent
		}
		noiseParam := r.URL.Query().Get("ns")
		if noiseParam == "" {
			noiseParam = cfg.NoiseIntent
		}

		testPenalty := cfg.TestPenalty
		if tp := r.URL.Query().Get("tp"); tp != "" {
			if v, err := strconv.ParseFloat(tp, 64); err == nil {
				testPenalty = v
			}
		}

		var results []*common.FileJob
		var processedFileCount int64

		if query != "" {
			// Make a copy of config so we can adjust per request
			searchCfg := *cfg
			if len(ext) != 0 {
				searchCfg.AllowListExtensions = []string{ext}
			} else {
				searchCfg.AllowListExtensions = []string{}
			}

			// Apply ranking/filter params
			searchCfg.Ranker = rankerParam
			searchCfg.GravityIntent = gravityParam
			searchCfg.NoiseIntent = noiseParam
			switch codeFilter {
			case "only-code":
				searchCfg.OnlyCode = true
				searchCfg.OnlyComments = false
				searchCfg.OnlyStrings = false
			case "only-comments":
				searchCfg.OnlyCode = false
				searchCfg.OnlyComments = true
				searchCfg.OnlyStrings = false
			case "only-strings":
				searchCfg.OnlyCode = false
				searchCfg.OnlyComments = false
				searchCfg.OnlyStrings = true
			case "only-declarations":
				searchCfg.OnlyDeclarations = true
			case "only-usages":
				searchCfg.OnlyUsages = true
			default:
				searchCfg.OnlyCode = false
				searchCfg.OnlyComments = false
				searchCfg.OnlyStrings = false
			}
			// Auto-switch ranker to structural when code filter is active (matches TUI behavior)
			if searchCfg.HasContentFilter() {
				searchCfg.Ranker = "structural"
				rankerParam = "structural"
			}

			ctx := context.Background()
			ch, stats := DoSearch(ctx, &searchCfg, query, cache)

			for fj := range ch {
				results = append(results, fj)
			}

			processedFileCount = stats.TextFileCount.Load()
			testIntent := ranker.HasTestIntent(strings.Fields(query))
			results = ranker.RankResults(searchCfg.Ranker, int(processedFileCount), results, searchCfg.StructuralRankerConfig(), searchCfg.ResolveGravityStrength(), searchCfg.ResolveNoiseSensitivity(), testPenalty, testIntent)
		}

		// Dedup (before pagination, so freed slots get backfilled)
		if r.URL.Query().Get("dedup") == "true" || r.URL.Query().Get("dedup") == "1" {
			results = ranker.DeduplicateResults(results)
		}

		// Create a random str to define where the start and end of
		// our highlight should be which we swap out later after we have
		// HTML escaped everything
		md5Digest := md5.New()
		fmtBegin := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("begin_%d", makeTimestampNano()))))
		fmtEnd := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("end_%d", makeTimestampNano()))))

		documentTermFrequency := ranker.CalculateDocumentTermFrequency(results)

		var searchResults []httpSearchResult
		extensionFacets := map[string]int{}

		// if we have more than the page size of results, lets just show the first page
		displayResults := results
		pages := httpCalculatePages(results, pageSize, query, snippetLength, ext, rankerParam, codeFilter, gravityParam, noiseParam)

		if displayResults != nil && len(displayResults) > pageSize {
			displayResults = displayResults[:pageSize]
		}
		if page != 0 && page <= len(pages) {
			end := page*pageSize + pageSize
			if end > len(results) {
				end = len(results)
			}
			displayResults = results[page*pageSize : end]
		}

		// loop over all results so we can get the facets
		for _, res := range results {
			extensionFacets[gocodewalker.GetExtension(res.Filename)] = extensionFacets[gocodewalker.GetExtension(res.Filename)] + 1
		}

		for _, res := range displayResults {
			fileMode := resolveSnippetMode(cfg.SnippetMode, res.Filename)

			if fileMode == "lines" {
				lineResults := snippet.FindMatchingLines(res, 2)
				if len(lineResults) == 0 {
					continue
				}
				var httpLines []httpLineResult
				prevLine := 0
				for _, lr := range lineResults {
					coloredLine := str.HighlightString(lr.Content, lr.Locs, fmtBegin, fmtEnd)
					coloredLine = html.EscapeString(coloredLine)
					coloredLine = strings.Replace(coloredLine, fmtBegin, "<strong>", -1)
					coloredLine = strings.Replace(coloredLine, fmtEnd, "</strong>", -1)
					gap := prevLine > 0 && lr.LineNumber > prevLine+1
					prevLine = lr.LineNumber
					httpLines = append(httpLines, httpLineResult{
						LineNumber: lr.LineNumber,
						Content:    template.HTML(coloredLine),
						Gap:        gap,
					})
				}
				// Compute byte range covering the displayed lines for click-through highlighting
				var startPos, endPos int
				if len(lineResults) > 0 {
					firstLine := lineResults[0].LineNumber
					lastLine := lineResults[len(lineResults)-1].LineNumber
					line := 1
					for i := 0; i < len(res.Content); i++ {
						if line == firstLine && (i == 0 || res.Content[i-1] == '\n') {
							startPos = i
						}
						if res.Content[i] == '\n' && line == lastLine {
							endPos = i
							break
						}
						if res.Content[i] == '\n' {
							line++
						}
					}
					if endPos == 0 {
						endPos = len(res.Content)
					}
				}

				searchResults = append(searchResults, httpSearchResult{
					Title:              res.Location,
					Location:           res.Location,
					Score:              res.Score,
					StartPos:           startPos,
					EndPos:             endPos,
					IsLineMode:         true,
					LineResults:        httpLines,
					Language:           res.Language,
					Lines:              res.Lines,
					Code:               res.Code,
					Comment:            res.Comment,
					Blank:              res.Blank,
					Complexity:         res.Complexity,
					DuplicateCount:     res.DuplicateCount,
					DuplicateLocations: res.DuplicateLocations,
				})
			} else {
				snippets := snippet.ExtractRelevant(res, documentTermFrequency, snippetLength)
				if len(snippets) == 0 {
					continue
				}
				v3 := snippets[0]

				// We have the snippet so now we need to highlight it
				// we get all the locations that fall in the snippet length
				// and then remove the length of the snippet cut which
				// makes our location line up with the snippet size
				var l [][]int
				for _, value := range res.MatchLocations {
					for _, s := range value {
						if s[0] >= v3.StartPos && s[1] <= v3.EndPos {
							s[0] = s[0] - v3.StartPos
							s[1] = s[1] - v3.StartPos
							l = append(l, s)
						}
					}
				}

				// We want to escape the output, so we highlight, then escape then replace
				// our special start and end strings with actual HTML
				coloredContent := v3.Content
				// If endpos = 0 don't highlight anything because it means its a filename match
				if v3.EndPos != 0 {
					coloredContent = str.HighlightString(v3.Content, l, fmtBegin, fmtEnd)
					coloredContent = html.EscapeString(coloredContent)
					coloredContent = strings.Replace(coloredContent, fmtBegin, "<strong>", -1)
					coloredContent = strings.Replace(coloredContent, fmtEnd, "</strong>", -1)
				}

				searchResults = append(searchResults, httpSearchResult{
					Title:              res.Location,
					Location:           res.Location,
					Content:            []template.HTML{template.HTML(coloredContent)},
					StartPos:           v3.StartPos,
					EndPos:             v3.EndPos,
					Score:              res.Score,
					Language:           res.Language,
					Lines:              res.Lines,
					Code:               res.Code,
					Comment:            res.Comment,
					Blank:              res.Blank,
					Complexity:         res.Complexity,
					DuplicateCount:     res.DuplicateCount,
					DuplicateLocations: res.DuplicateLocations,
				})
			}
		}

		searchData := httpSearch{
			SearchTerm:          query,
			SnippetSize:         snippetLength,
			Results:             searchResults,
			ResultsCount:        len(results),
			RuntimeMilliseconds: makeTimestampMilli() - startTime,
			ProcessedFileCount:  processedFileCount,
			ExtensionFacet:      httpCalculateExtensionFacet(extensionFacets, query, snippetLength, rankerParam, codeFilter, gravityParam, noiseParam),
			Pages:               pages,
			Ext:                 ext,
			Ranker:              rankerParam,
			CodeFilter:          codeFilter,
			Gravity:             gravityParam,
			Noise:               noiseParam,
		}

		if r.URL.Query().Get("format") == "json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(searchData)
			return
		}

		err := searchTmpl.Execute(w, searchData)
		if err != nil {
			log.Printf("template execute error: %v", err)
		}
	})

	fmt.Printf("starting HTTP server on %s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, nil))
}

func httpCalculateExtensionFacet(extensionFacets map[string]int, query string, snippetLength int, rankerParam, codeFilter, gravityParam, noiseParam string) []httpFacetResult {
	var ef []httpFacetResult

	for k, v := range extensionFacets {
		ef = append(ef, httpFacetResult{
			Title:       k,
			Count:       v,
			SearchTerm:  query,
			SnippetSize: snippetLength,
			Ranker:      rankerParam,
			CodeFilter:  codeFilter,
			Gravity:     gravityParam,
			Noise:       noiseParam,
		})
	}

	sort.Slice(ef, func(i, j int) bool {
		if ef[i].Count == ef[j].Count {
			return strings.Compare(ef[i].Title, ef[j].Title) < 0
		}
		return ef[i].Count > ef[j].Count
	})

	return ef
}

func httpCalculatePages(results []*common.FileJob, pageSize int, query string, snippetLength int, ext string, rankerParam, codeFilter, gravityParam, noiseParam string) []httpPageResult {
	var pages []httpPageResult

	if len(results) == 0 {
		return pages
	}

	if len(results) <= pageSize {
		pages = append(pages, httpPageResult{
			SearchTerm:  query,
			SnippetSize: snippetLength,
			Value:       0,
			Name:        "1",
			Ranker:      rankerParam,
			CodeFilter:  codeFilter,
			Gravity:     gravityParam,
			Noise:       noiseParam,
		})
		return pages
	}

	a := 1
	if len(results)%pageSize == 0 {
		a = 0
	}

	for i := 0; i < len(results)/pageSize+a; i++ {
		pages = append(pages, httpPageResult{
			SearchTerm:  query,
			SnippetSize: snippetLength,
			Value:       i,
			Name:        strconv.Itoa(i + 1),
			Ext:         ext,
			Ranker:      rankerParam,
			CodeFilter:  codeFilter,
			Gravity:     gravityParam,
			Noise:       noiseParam,
		})
	}
	return pages
}

func tryParseInt(x string, def int) int {
	t, err := strconv.Atoi(x)
	if err != nil {
		return def
	}
	return t
}

func makeTimestampMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func makeTimestampNano() int64 {
	return time.Now().UnixNano()
}
