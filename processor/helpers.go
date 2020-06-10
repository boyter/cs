// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package processor

import (
	"fmt"
	str "github.com/boyter/cs/str"
	"os"
	"time"
)

// Returns the current time as a millisecond timestamp
func makeTimestampMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// Returns the current time as a nanosecond timestamp as some things
// are far too fast to measure using nanoseconds
func makeTimestampNano() int64 {
	return time.Now().UnixNano()
}

const letterDigitFuzzyBytes = `abcdefghijklmnopqrstuvwxyz1234567890~!@#$%^&*()_+-=[]\{}|;':"',./<>?~`

// Takes in a term and returns a slice of them which contains all the
// fuzzy versions of that str with things such as mis-spellings
// somewhat based on https://norvig.com/spell-correct.html
func makeFuzzyDistanceOne(term string) []string {
	vals := []string{term}

	if len(term) <= 2 {
		return vals
	}

	// This tends to produce bad results
	// Split apart so turn "test" into "t" "est" then "te" "st"
	//for i := 0; i < len(term); i++ {
	//	vals = append(vals, term[:i])
	//	vals = append(vals, term[i:])
	//}

	// Delete letters so turn "test" into "est" "tst" "tet"
	for i := 0; i < len(term); i++ {
		vals = append(vals, term[:i]+term[i+1:])
	}

	// Replace a letter or digit which effectively does transpose for us
	for i := 0; i < len(term); i++ {
		for _, b := range letterDigitFuzzyBytes {
			vals = append(vals, term[:i]+string(b)+term[i+1:])
		}
	}

	// Insert a letter or digit
	for i := 0; i < len(term); i++ {
		for _, b := range letterDigitFuzzyBytes {
			vals = append(vals, term[:i]+string(b)+term[i:])
		}
	}

	return str.RemoveStringDuplicates(vals)
}

// Similar to fuzzy 1 but in this case we add letters
// to make the distance larger
func makeFuzzyDistanceTwo(term string) []string {
	vals := makeFuzzyDistanceOne(term)

	// Maybe they forgot to type a letter? Try adding one
	for i := 0; i < len(term)+1; i++ {
		for _, b := range letterDigitFuzzyBytes {
			vals = append(vals, term[:i]+string(b)+term[i:])
		}
	}

	return str.RemoveStringDuplicates(vals)
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
