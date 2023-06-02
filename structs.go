// SPDX-License-Identifier: MIT OR Unlicense

package main

// FileJob is a struct used to hold the results of processing internally before sent to the formatter
type FileJob struct {
	Filename       string
	Extension      string
	Location       string
	Content        []byte
	Bytes          int
	Hash           []byte
	Binary         bool
	Score          float64
	MatchLocations map[string][][]int
	Minified       bool
}
