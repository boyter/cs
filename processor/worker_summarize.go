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

		//// TODO flip the order of these so we extract the snippet first then highlight
		//coloredContent := str.HighlightString(string(res.Content), l, fmtBegin, fmtEnd)
		//snippets := extractSnippets(coloredContent, l, int(SnippetLength), "…")
		//
		//for _, s := range snippets {
		//	fmt.Println(s.Content)
		//	fmt.Println("----------------------------------------------------")
		//}


		v3 := extractRelevantV3(res, documentFrequency, int(SnippetLength), "…")

		// we have the snippet now we need to highlight it
		// get all the locations that fall in the snippet length
		// remove the size of the cut, and highlight
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

		coloredContent := str.HighlightString(string(v3.Content), l, fmtBegin, fmtEnd)

		fmt.Println(coloredContent)
	}
}
