package processor

import (
	"fmt"
	str "github.com/boyter/cs/string"
	"github.com/fatih/color"
)

type ResultSummarizer struct {
	input chan *fileJob
	ResultLimit int64
	FileReaderWorker *FileReaderWorker2
}

func NewResultSummarizer(input chan *fileJob) ResultSummarizer {
	return ResultSummarizer{
		input: input,
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

	// Add one here so even if we get the same value it produces some sort of result
	rankResults(int(f.FileReaderWorker.GetFileCount()), results)

	fmtBegin := "\033[1;31m"
	fmtEnd := "\033[0m"

	for _, res := range results {
		color.Magenta("%s (%.3f)", res.Location, res.Score)

		// Combine all the locations such that we can highlight correctly
		l := [][]int{}
		for _, value := range res.MatchLocations {
			l = append(l, value...)
		}

		coloredContent := str.HighlightString(string(res.Content), l, fmtBegin, fmtEnd)
		// TODO account for cutting off the highlight portions
		relevant, _, _ := str.ExtractRelevant(coloredContent, l, int(SnippetLength), str.CalculatePrevCount(int(SnippetLength), 6), "…")

		fmt.Println(relevant)
		fmt.Println("\n\n")
	}
}

var (
	Black   = Color("\033[1;30m%s\033[0m")
	Red     = Color("\033[1;31m%s\033[0m")
	Green   = Color("\033[1;32m%s\033[0m")
	Yellow  = Color("\033[1;33m%s\033[0m")
	Purple  = Color("\033[1;34m%s\033[0m")
	Magenta = Color("\033[1;35m%s\033[0m")
	Teal    = Color("\033[1;36m%s\033[0m")
	White   = Color("\033[1;37m%s\033[0m")
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}