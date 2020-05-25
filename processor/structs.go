// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package processor

// fileJob is a struct used to hold all of the results of processing internally before sent to the formatter
type fileJob struct {
	Filename       string
	Extension      string
	Location       string
	Content        []byte
	Bytes          int64
	Hash           []byte
	Binary         bool
	Score          float64
	MatchLocations map[string][][]int
	Minified       bool
}
