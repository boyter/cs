// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	str "github.com/boyter/go-string"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"

	"github.com/boyter/cs/pkg/common"
	"github.com/boyter/cs/pkg/ranker"
	"github.com/boyter/cs/pkg/snippet"
)

// ConsoleSearch runs a non-interactive search and prints results to stdout.
func ConsoleSearch(cfg *Config) {
	query := strings.Join(cfg.SearchString, " ")

	ctx := context.Background()
	ch, stats := DoSearch(ctx, cfg, query, nil)

	// Collect all results
	var results []*common.FileJob
	for fj := range ch {
		results = append(results, fj)
	}

	// Rank results
	textFileCount := int(stats.TextFileCount.Load())
	testIntent := ranker.HasTestIntent(strings.Fields(query))
	results = ranker.RankResults(cfg.Ranker, textFileCount, results, cfg.StructuralRankerConfig(), cfg.ResolveGravityStrength(), cfg.ResolveNoiseSensitivity(), cfg.TestPenalty, testIntent)

	// Dedup (before limit, so freed slots get backfilled)
	if cfg.Dedup {
		results = ranker.DeduplicateResults(results)
	}

	// Apply result limit (after dedup)
	if cfg.ResultLimit > 0 && len(results) > cfg.ResultLimit {
		results = results[:cfg.ResultLimit]
	}

	// Route to formatter
	switch cfg.Format {
	case "json":
		formatJSON(cfg, results)
	case "vimgrep":
		formatVimGrep(cfg, results)
	default:
		formatDefault(cfg, results)
	}
}

func formatDefault(cfg *Config, results []*common.FileJob) {
	noColor := os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()))

	fmtBegin := "\033[1;31m"
	fmtEnd := "\033[0m"
	if noColor {
		fmtBegin = ""
		fmtEnd = ""
		color.NoColor = true
	}

	documentFrequency := ranker.CalculateDocumentTermFrequency(results)

	for _, res := range results {
		fileMode := resolveSnippetMode(cfg.SnippetMode, res.Filename)

		if fileMode == "lines" {
			lineResults := snippet.FindMatchingLines(res, 2)
			if len(lineResults) == 0 {
				continue
			}
			lines := fmt.Sprintf("%d-%d", lineResults[0].LineNumber, lineResults[len(lineResults)-1].LineNumber)
			codeStats := formatCodeStats(res)
			if res.Language != "" {
				color.Magenta(fmt.Sprintf("%s (%s) Lines %s (%.3f)%s", res.Location, res.Language, lines, res.Score, codeStats))
			} else {
				color.Magenta(fmt.Sprintf("%s Lines %s (%.3f)%s", res.Location, lines, res.Score, codeStats))
			}
			if res.DuplicateCount > 0 {
				color.Cyan(fmt.Sprintf("  +%d duplicate(s) in: %s", res.DuplicateCount, strings.Join(res.DuplicateLocations, ", ")))
			}
			prevLine := 0
			for _, lr := range lineResults {
				if prevLine > 0 && lr.LineNumber > prevLine+1 {
					fmt.Println("")
				}
				prevLine = lr.LineNumber
				var displayContent string
				if !noColor && !cfg.NoSyntax {
					displayContent = RenderANSILine(lr.Content, lr.Locs)
				} else {
					displayContent = str.HighlightString(lr.Content, lr.Locs, fmtBegin, fmtEnd)
				}
				fmt.Printf("%4d %s\n", lr.LineNumber, displayContent)
			}
			fmt.Println("")
		} else {
			snippets := snippet.ExtractRelevant(res, documentFrequency, cfg.SnippetLength)
			if len(snippets) > cfg.SnippetCount {
				snippets = snippets[:cfg.SnippetCount]
			}

			lines := ""
			for i := 0; i < len(snippets); i++ {
				lines += fmt.Sprintf("%d-%d ", snippets[i].LineStart, snippets[i].LineEnd)
			}

			codeStats := formatCodeStats(res)
			if res.Language != "" {
				color.Magenta(fmt.Sprintf("%s (%s) Lines %s(%.3f)%s", res.Location, res.Language, lines, res.Score, codeStats))
			} else {
				color.Magenta(fmt.Sprintf("%s Lines %s(%.3f)%s", res.Location, lines, res.Score, codeStats))
			}
			if res.DuplicateCount > 0 {
				color.Cyan(fmt.Sprintf("  +%d duplicate(s) in: %s", res.DuplicateCount, strings.Join(res.DuplicateLocations, ", ")))
			}

			for i := 0; i < len(snippets); i++ {
				// Get all match locations that fall within this snippet
				var l [][]int
				for _, value := range res.MatchLocations {
					for _, s := range value {
						if len(s) < 2 {
							continue
						}
						if s[0] >= snippets[i].StartPos && s[1] <= snippets[i].EndPos {
							l = append(l, []int{
								s[0] - snippets[i].StartPos,
								s[1] - snippets[i].StartPos,
							})
						}
					}
				}

				displayContent := snippets[i].Content

				// Highlight if we have positions to highlight
				if !(snippets[i].StartPos == 0 && snippets[i].EndPos == 0) {
					if !noColor && !cfg.NoSyntax {
						displayContent = RenderANSILine(snippets[i].Content, l)
					} else {
						displayContent = str.HighlightString(snippets[i].Content, l, fmtBegin, fmtEnd)
					}
				}

				fmt.Println(displayContent)
				if i == len(snippets)-1 {
					fmt.Println("")
				} else {
					fmt.Println("")
					fmt.Println("\u001B[1;37m……………snip……………\u001B[0m")
					fmt.Println("")
				}
			}
		}
	}
}

// formatCodeStats returns a formatted string of code counting stats for a file result.
// Returns empty string if no stats are available.
func formatCodeStats(res *common.FileJob) string {
	if res.Lines == 0 {
		return ""
	}
	return fmt.Sprintf(" Lines:%d (Code:%d Comment:%d Blank:%d Complexity:%d)", res.Lines, res.Code, res.Comment, res.Blank, res.Complexity)
}

type jsonLineResult struct {
	LineNumber int     `json:"line_number"`
	Content    string  `json:"content"`
	Locs       [][]int `json:"match_positions,omitempty"`
}

type jsonResult struct {
	Filename           string           `json:"filename"`
	Location           string           `json:"location"`
	Content            string           `json:"content,omitempty"`
	Score              float64          `json:"score"`
	MatchLocations     [][]int          `json:"matchlocations,omitempty"`
	Lines              []jsonLineResult `json:"lines,omitempty"`
	Language           string           `json:"language,omitempty"`
	TotalLines         int64            `json:"total_lines"`
	Code               int64            `json:"code"`
	Comment            int64            `json:"comment"`
	Blank              int64            `json:"blank"`
	Complexity         int64            `json:"complexity"`
	DuplicateCount     int              `json:"duplicate_count,omitempty"`
	DuplicateLocations []string         `json:"duplicate_locations,omitempty"`
}

// buildJSONResults converts ranked FileJob results into a slice of jsonResult
// suitable for JSON serialization. Used by both formatJSON and the MCP server.
func buildJSONResults(cfg *Config, results []*common.FileJob) []jsonResult {
	var jsonResults []jsonResult

	documentFrequency := ranker.CalculateDocumentTermFrequency(results)

	for _, res := range results {
		fileMode := resolveSnippetMode(cfg.SnippetMode, res.Filename)

		if fileMode == "lines" {
			lineResults := snippet.FindMatchingLines(res, 2)
			if len(lineResults) == 0 {
				continue
			}
			var jLines []jsonLineResult
			for _, lr := range lineResults {
				jLines = append(jLines, jsonLineResult{
					LineNumber: lr.LineNumber,
					Content:    lr.Content,
					Locs:       lr.Locs,
				})
			}
			jsonResults = append(jsonResults, jsonResult{
				Filename:           res.Filename,
				Location:           res.Location,
				Score:              res.Score,
				Lines:              jLines,
				Language:           res.Language,
				TotalLines:         res.Lines,
				Code:               res.Code,
				Comment:            res.Comment,
				Blank:              res.Blank,
				Complexity:         res.Complexity,
				DuplicateCount:     res.DuplicateCount,
				DuplicateLocations: res.DuplicateLocations,
			})
		} else {
			snippets := snippet.ExtractRelevant(res, documentFrequency, cfg.SnippetLength)
			if len(snippets) == 0 {
				continue
			}
			v3 := snippets[0]

			var l [][]int
			for _, value := range res.MatchLocations {
				for _, s := range value {
					if len(s) < 2 {
						continue
					}
					if s[0] >= v3.StartPos && s[1] <= v3.EndPos {
						l = append(l, []int{
							s[0] - v3.StartPos,
							s[1] - v3.StartPos,
						})
					}
				}
			}

			jsonResults = append(jsonResults, jsonResult{
				Filename:           res.Filename,
				Location:           res.Location,
				Content:            v3.Content,
				Score:              res.Score,
				MatchLocations:     l,
				Language:           res.Language,
				TotalLines:         res.Lines,
				Code:               res.Code,
				Comment:            res.Comment,
				Blank:              res.Blank,
				Complexity:         res.Complexity,
				DuplicateCount:     res.DuplicateCount,
				DuplicateLocations: res.DuplicateLocations,
			})
		}
	}

	return jsonResults
}

func formatJSON(cfg *Config, results []*common.FileJob) {
	jsonResults := buildJSONResults(cfg, results)
	jsonString, _ := json.Marshal(jsonResults)
	if cfg.FileOutput == "" {
		fmt.Println(string(jsonString))
	} else {
		_ = os.WriteFile(cfg.FileOutput, jsonString, 0600)
		fmt.Println("results written to " + cfg.FileOutput)
	}
}

func formatVimGrep(cfg *Config, results []*common.FileJob) {
	snippetLen := 50 // vim quickfix puts each hit on its own line
	documentFrequency := ranker.CalculateDocumentTermFrequency(results)

	var vimGrepOutput []string
	for _, res := range results {
		fileMode := resolveSnippetMode(cfg.SnippetMode, res.Filename)

		if fileMode == "lines" {
			lineResults := snippet.FindMatchingLines(res, 0)
			for _, lr := range lineResults {
				col := 1
				if len(lr.Locs) > 0 {
					col = lr.Locs[0][0] + 1
				}
				hint := strings.ReplaceAll(lr.Content, "\n", "\\n")
				line := fmt.Sprintf("%v:%v:%v:%v", res.Location, lr.LineNumber, col, hint)
				vimGrepOutput = append(vimGrepOutput, line)
			}
		} else {
			snippets := snippet.ExtractRelevant(res, documentFrequency, snippetLen)
			if len(snippets) > cfg.SnippetCount {
				snippets = snippets[:cfg.SnippetCount]
			}

			for _, snip := range snippets {
				hint := strings.ReplaceAll(snip.Content, "\n", "\\n")
				line := fmt.Sprintf("%v:%v:%v:%v", res.Location, snip.LineStart, snip.StartPos, hint)
				vimGrepOutput = append(vimGrepOutput, line)
			}
		}
	}

	printable := strings.Join(vimGrepOutput, "\n")
	fmt.Println(printable)
}
