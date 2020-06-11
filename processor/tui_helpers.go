// SPDX-License-Identifier: MIT OR Unlicense

package processor

import (
	"os"
	"strconv"
)

func getLocated(res *fileJob, v3 Snippet) [][]int {
	var l [][]int

	// For all of the match locations we have only keep the ones that should be inside
	// where we are matching
	for _, value := range res.MatchLocations {
		for _, s := range value {
			if s[0] >= v3.StartPos && s[1] <= v3.EndPos {
				// Have to create a new one to avoid changing the position
				// unlike in others where we throw away the results afterwards
				t := []int{s[0] - v3.StartPos, s[1] - v3.StartPos}
				l = append(l, t)
			}
		}
	}

	return l
}

// Because in TUI mode there is no way to just dump logs out to screen this
// method exists which logs out to disk. Very cheap and nasty but works well enough.
func debugLogger(text string) {
	f, err := os.OpenFile("cs.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(strconv.Itoa(int(makeTimestampMilli())) + " " + text + "\n"); err != nil {
		panic(err)
	}
}
