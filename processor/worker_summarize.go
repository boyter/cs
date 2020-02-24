package processor

import (
	"fmt"
	str "github.com/boyter/cs/string"
	"github.com/fatih/color"
)

type ResultSummarizer struct {
	input            chan *fileJob
	ResultLimit      int64
	FileReaderWorker *FileReaderWorker2
}

func NewResultSummarizer(input chan *fileJob) ResultSummarizer {
	return ResultSummarizer{
		input:       input,
		ResultLimit: -1,
	}
}

func (f *ResultSummarizer) Start() {
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

	rankResults(int(f.FileReaderWorker.GetFileCount()), results)

	fmtBegin := "\033[1;31m"
	fmtEnd := "\033[0m"

	documentFrequency := calculateDocumentFrequency(results)

	for _, res := range results {
		color.Magenta("%s (%.3f)", res.Location, res.Score)

		// Combine all the locations such that we can highlight correctly
		l := [][]int{}
		for _, value := range res.MatchLocations {
			l = append(l, value...)
		}

		// TODO flip the order of these so we extract the snippet first then highlight
		coloredContent := str.HighlightString(string(res.Content), l, fmtBegin, fmtEnd)
		snippets := extractSnippets(coloredContent, l, int(SnippetLength), "…")

		for _, s := range snippets {
			fmt.Println(s.Content)
			fmt.Println("----------------------------------------------------")
		}

		extractRelevantV3(res, documentFrequency, int(SnippetLength), "…")
	}
}
