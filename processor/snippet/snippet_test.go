// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package snippet

import (
	"io/ioutil"
	"math/rand"
	"strings"
	"testing"
)

func TestExtractRelevant(t *testing.T) {
	relevant := ExtractRelevant("this is some text (╯°□°）╯︵ ┻━┻) the thing we want is here", []LocationType{
		{
			Term:     "the",
			Location: 31,
		},
	}, 30, 20, "...")

	if len(relevant) == 0 {
		t.Error("Expected some value")
	}
}

func TestExtractRelevantOddCase(t *testing.T) {
	relevant := ExtractRelevant(`package processor

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
		rel := snippet.ExtractRelevant(coloredContent, locations, int(SnippetLength), snippet.CalculatePrevCount(int(SnippetLength), 6), "…")

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
		rel := snippet.ExtractRelevant(string(res.Content), locations, int(SnippetLength), snippet.GetPrevCount(int(SnippetLength)), "…")

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
`, []LocationType{
		{
			Term:     `printer`,
			Location: 119,
		},
		{
			Term:     `printer`,
			Location: 810,
		},
		{
			Term:     `printer`,
			Location: 2811,
		},
	}, 300, 50, "...")

	if !strings.Contains(relevant, "printer") {
		t.Error("Expected printer to exist somewhere")
	}
}

func TestExtractLocation(t *testing.T) {
	content, _ := ioutil.ReadFile("blns.json")

	for i := 0; i < 10000; i++ {
		location := ExtractLocation([]byte(RandStringBytes(rand.Intn(2))), content, 50)

		for l := range location {
			if l > len([]rune(string(content))) {
				t.Error("Should not be longer")
			}
		}
	}
}

func TestExtractLocations(t *testing.T) {
	content, _ := ioutil.ReadFile("blns.json")

	locations := ExtractLocations([][]byte{[]byte("test"), []byte("something"), []byte("other")}, content)

	if len(locations) == 0 {
		t.Error("Expected at least one location")
	}
}

func TestGetPrevCount(t *testing.T) {
	got := GetPrevCount(300)

	if got != 50 {
		t.Error("Expected 50 got", got)
	}
}

// Designed to catch out any issues with unicode and the like
func TestFuzzy(t *testing.T) {
	content, _ := ioutil.ReadFile("blns.json")

	split := strings.Split("a b c d e f g h i j k l m n o p q r s t u v w x y z", " ")

	for i, t := range split {
		ExtractRelevant(string(content), []LocationType{
			{
				Term:     t,
				Location: i,
			},
		}, 300, 50, "...")
	}

	for i := 0; i < 10000; i++ {
		ExtractRelevant(string(content), []LocationType{
			{
				Term:     RandStringBytes(rand.Intn(10)),
				Location: rand.Intn(len(content)),
			},
		}, 300, 50, "...")
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~!@#$%^&*()_+`1234567890-=[]\\{}|;':\",./<>?"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
