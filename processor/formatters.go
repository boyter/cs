package processor

import (
	"fmt"
	"github.com/alecthomas/chroma/quick"
	"os"
	"sort"
	"time"
)


func fileSummarize(input chan *FileJob) string {
	//switch {
	//case strings.ToLower(Format) == "json":
	//	return toJSON(input)
	//}

	// Collect results
	results := []*FileJob{}
	for res := range input {
		results = append(results, res)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	for _, res := range results {
		fmt.Println(res.Filename, res.Score)
		rel := extractRelevant(SearchString, string(res.Content), 300, 5, "…")

		// NB the below does not work in the old CMD.exe on windows
		_ = quick.Highlight(os.Stdout, rel, "go", "terminal16m", "monokai")
		fmt.Println("")
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
