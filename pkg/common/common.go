// SPDX-License-Identifier: MIT

package common

// Query complexity limits — stricter for MCP, relaxed for interactive modes.
const (
	MaxQueryCharsMCP     = 250
	MaxQueryTermsMCP     = 12
	MaxQueryCharsDefault = 1000
	MaxQueryTermsDefault = 50
)

// FileJob is a struct used to hold the results of processing internally before sent to the formatter
type FileJob struct {
	Filename           string
	Extension          string
	Location           string
	Content            []byte
	ContentByteType    []byte // Per-byte classification from scc (code/comment/string/blank)
	Bytes              int
	Binary             bool
	Score              float64
	MatchLocations     map[string][][]int
	Minified           bool
	Language           string
	Lines              int64
	Code               int64
	Comment            int64
	Blank              int64
	Complexity         int64
	MatchHash          string
	DuplicateCount     int
	DuplicateLocations []string
}
