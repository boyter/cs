package processor

import (
	"os"
	"strconv"
)

func getLocated(res *fileJob, v3 Snippet) [][]int {
	l := [][]int{}

	// For all of the match locations we have only keep the ones that should be inside
	// where we are matching
	for _, value := range res.MatchLocations {
		// TODO bug search for "collins lizzy mr" in TUI view does not highlight mr but does pure command line probably in here somewhere
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
