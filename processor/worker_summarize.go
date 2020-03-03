package processor

import (
	"fmt"
	str "github.com/boyter/cs/string"
	"github.com/fatih/color"
)

type ResultSummarizer struct {
	input            chan *fileJob
	ResultLimit      int64
	FileReaderWorker *FileReaderWorker
	SnippetCount     int
}

func NewResultSummarizer(input chan *fileJob) ResultSummarizer {
	return ResultSummarizer{
		input:       input,
		ResultLimit: -1,
		SnippetCount: 1,
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

		v3 := extractRelevantV3(res, documentFrequency, int(SnippetLength), "…")[0]

		// We have the snippet so now we need to highlight it
		// we get all the locations that fall in the snippet length
		// and then remove the length of the snippet cut which
		// makes out location line up with the snippet size
		l := [][]int{}
		for _, value := range res.MatchLocations {
			for _, s := range value {
				if s[0] >= v3.StartPos && s[1] <= v3.EndPos {
					s[0] = s[0] - v3.StartPos
					s[1] = s[1] - v3.StartPos
					l = append(l, s)
				}
			}
		}

		coloredContent := str.HighlightString(v3.Content, l, fmtBegin, fmtEnd)

		fmt.Println(coloredContent)
		fmt.Println("")
	}
}
