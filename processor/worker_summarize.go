package processor

import (
	"fmt"
	"github.com/boyter/cs/processor/snippet"
	"github.com/fatih/color"
	str "github.com/boyter/cs/string"
)

type ResultSummarizer struct {
	input chan *fileJob
	ResultLimit int64
}

func NewResultSummarizer(input chan *fileJob) ResultSummarizer {
	return ResultSummarizer{
		input: input,
		ResultLimit: -1,
	}
}

func (f *ResultSummarizer) Start() string {
	// First step is to collect results so we can rank them
	results := []*fileJob{}
	for res := range f.input {
		results = append(results, res)
	}

	// TODO this should probably be done inside the processor to save on CPU there
	if f.ResultLimit != -1 {
		if int64(len(results)) > f.ResultLimit {
			results = results[:f.ResultLimit]
		}
	}

	// Rank results then sort for display
	rankResults(SearchBytes, results)
	sortResults(results)

	for _, res := range results {
		fmtBegin := "\033[1;31m"
		fmtEnd := "\033[0m"
		color.Magenta("%s (%.3f)", res.Location, res.Score)

		// Combine all the locations such that we can highlight
		locs := [][]int{}
		for _, value := range res.MatchLocations {
			locs = append(locs, value...)
		}

		coloredContent := str.HighlightString(string(res.Content), locs, fmtBegin, fmtEnd)
		relevant, _, _ := str.ExtractRelevant(coloredContent, locs, int(SnippetLength), snippet.CalculatePrevCount(int(SnippetLength), 6), "…")

		fmt.Println(relevant)
		fmt.Println("")
	}

	return ""
}
