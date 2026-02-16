// SPDX-License-Identifier: MIT OR Unlicense

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
	ch, stats := DoSearch(ctx, cfg, query)

	// Collect all results
	var results []*common.FileJob
	for fj := range ch {
		results = append(results, fj)
	}

	// Apply result limit
	if cfg.ResultLimit > 0 && len(results) > cfg.ResultLimit {
		results = results[:cfg.ResultLimit]
	}

	// Rank results
	textFileCount := int(stats.TextFileCount.Load())
	results = ranker.RankResults(cfg.Ranker, textFileCount, results)

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
		snippets := snippet.ExtractRelevant(res, documentFrequency, cfg.SnippetLength)
		if len(snippets) > cfg.SnippetCount {
			snippets = snippets[:cfg.SnippetCount]
		}

		lines := ""
		for i := 0; i < len(snippets); i++ {
			lines += fmt.Sprintf("%d-%d ", snippets[i].LineStart, snippets[i].LineEnd)
		}

		color.Magenta(fmt.Sprintf("%s Lines %s(%.3f)", res.Location, lines, res.Score))

		for i := 0; i < len(snippets); i++ {
			// Get all match locations that fall within this snippet
			var l [][]int
			for _, value := range res.MatchLocations {
				for _, s := range value {
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
				displayContent = str.HighlightString(snippets[i].Content, l, fmtBegin, fmtEnd)
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

type jsonResult struct {
	Filename       string  `json:"filename"`
	Location       string  `json:"location"`
	Content        string  `json:"content"`
	Score          float64 `json:"score"`
	MatchLocations [][]int `json:"matchlocations"`
}

func formatJSON(cfg *Config, results []*common.FileJob) {
	var jsonResults []jsonResult

	documentFrequency := ranker.CalculateDocumentTermFrequency(results)

	for _, res := range results {
		snippets := snippet.ExtractRelevant(res, documentFrequency, cfg.SnippetLength)
		if len(snippets) == 0 {
			continue
		}
		v3 := snippets[0]

		var l [][]int
		for _, value := range res.MatchLocations {
			for _, s := range value {
				if s[0] >= v3.StartPos && s[1] <= v3.EndPos {
					l = append(l, []int{
						s[0] - v3.StartPos,
						s[1] - v3.StartPos,
					})
				}
			}
		}

		jsonResults = append(jsonResults, jsonResult{
			Filename:       res.Filename,
			Location:       res.Location,
			Content:        v3.Content,
			Score:          res.Score,
			MatchLocations: l,
		})
	}

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

	printable := strings.Join(vimGrepOutput, "\n")
	fmt.Println(printable)
}
