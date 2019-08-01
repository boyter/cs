package processor

import (
	"fmt"
	"github.com/boyter/sc/processor/snippet"
	"github.com/fatih/color"
	"os"
	"sort"
	"time"
)

func fileSummarize(input chan *FileJob) string {
	//switch {
	//case strings.ToLower(Format) == "json":
	//	return toJSON(input)
	//}

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
		return results[i].Score > results[j].Score
	})

	for _, res := range results {
		color.Magenta("%s (%.3f)", res.Location, res.Score)

		locs := []int{}
		for k := range res.Locations {
			locs = append(locs, res.Locations[k]...)
		}
		locs = RemoveIntDuplicates(locs)

		rel := snippet.ExtractRelevant(SearchString, string(res.Content), locs, int(SnippetLength), snippet.GetPrevCount(int(SnippetLength)), "…")
		fmt.Println(rel)

		// break up and highlight
		// base the highligt off lower so we can ensure we match correctly

		// find all of the matching sections so we can highlight them in the relevant part
		//fmt.Print(rel[:10])
		//color.Set(color.FgHiRed)
		//fmt.Print(rel[10:20])
		//color.Unset()
		//fmt.Print(rel[20:])

		//fmt.Println(rel)
		fmt.Println("")
	}

	return ""
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
	_, _ = fmt.Fprintf(os.Stderr, "ERROR %s: %s", getFormattedTime(), msg)
}
