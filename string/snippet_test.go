package string

import (
	"strings"
	"testing"
)

func TestExtractRelevant(t *testing.T) {
	locations := [][]int{}
	locations = append(locations, []int{31, 35})

	fulltext := "this is some text (╯°□°）╯︵ ┻━┻) the thing we want is here"
	relevant, startPos, endPos := extractRelevantV1(fulltext, locations, 30, 20, "...")

	if len(relevant) == 0 && startPos == 0 && endPos == len(fulltext) {
		t.Error("Expected some value")
	}
}

func TestExtractRelevantOddCase(t *testing.T) {
	fulltext := `package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/boyter/cs/processor/printer"
	"github.com/boyter/cs/processor/snippet"
	"github.com/fatih/color"
)

func fileSummarize(input chan *FileJob) string {
	switch {
	case strings.ToLower(Format) == "json":
		return toJSON(input)
	}

	// Collect results so we can rank them
	results := []*FileJob{}
	for res := range input {
		results = append(results, res)
	}

	if int64(len(results)) > ResultLimit {
		results = results[:ResultLimit]
	}

	// Rank results then sort for display
	RankResults(results)
	SortResults(results)

	for _, res := range results {
		fmtBegin := "\033[1;31m"
		fmtEnd := "\033[0m"
		color.Magenta("%s (%.3f)", res.Location, res.Score)

		locations := GetResultLocations(res)
		coloredContent := printer.WriteHighlights(res.Content, res.Locations, fmtBegin, fmtEnd)
		rel := snippet.extractRelevantV1(coloredContent, locations, int(SnippetLength), snippet.CalculatePrevCount(int(SnippetLength), 6), "…")

		fmt.Println(rel)
		fmt.Println("")
	}

	return ""
}

func toJSON(input chan *FileJob) string {
	// Collect results so we can rank them
	results := []*FileJob{}
	for res := range input {
		results = append(results, res)
	}

	if int64(len(results)) > ResultLimit {
		results = results[:ResultLimit]
	}

	// Rank results then sort for display
	RankResults(results)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return strings.Compare(results[i].Location, results[j].Location) < 0
		}

		return results[i].Score > results[j].Score
	})

	r := []JsonResult{}

	for _, res := range results {
		locations := GetResultLocations(res)
		rel := snippet.extractRelevantV1(string(res.Content), locations, int(SnippetLength), snippet.getPrevCount(int(SnippetLength)), "…")

		r = append(r, JsonResult{
			Filename:  res.Filename,
			Extension: res.Extension,
			Location:  res.Location,
			Bytes:     res.Bytes,
			Score:     res.Score,
			Snippet:   rel,
		})
	}

	data, _ := json.Marshal(r)
	return string(data)
}

// Get the time as standard UTC/Zulu format
func getFormattedTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Prints a message to stdout if flag to enable warning output is set
func printWarn(msg string) {
	if Verbose {
		fmt.Println(fmt.Sprintf(" WARN %s: %s", getFormattedTime(), msg))
	}
}

// Prints a message to stdout if flag to enable debug output is set
func printDebug(msg string) {
	if Debug {
		fmt.Println(fmt.Sprintf("DEBUG %s: %s", getFormattedTime(), msg))
	}
}

// Prints a message to stdout if flag to enable trace output is set
func printTrace(msg string) {
	if Trace {
		fmt.Println(fmt.Sprintf("TRACE %s: %s", getFormattedTime(), msg))
	}
}

// Used when explicitly for os.exit output when crashing out
func printError(msg string) {
	if Error {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR %s: %s", getFormattedTime(), msg)
	}
}
`

	locations := IndexAllIgnoreCaseUnicode(fulltext, `printer`, -1)

	relevant, _, _ := extractRelevantV1(fulltext, locations, 300, 50, "...")

	if !strings.Contains(strings.ToLower(relevant), "printer") {
		t.Error("Expected printer to exist somewhere")
	}
}

func TestGetPrevCount(t *testing.T) {
	got := getPrevCount(300)

	if got != 50 {
		t.Error("Expected 50 got", got)
	}
}
