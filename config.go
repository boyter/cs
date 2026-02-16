// SPDX-License-Identifier: MIT

package main

// Config holds all CLI-configurable fields for the search tool.
type Config struct {
	// Search
	SearchString  []string
	CaseSensitive bool
	SnippetLength int
	SnippetCount  int
	SnippetMode   string
	Ranker        string
	ResultLimit   int

	// File walker
	Directory              string
	FindRoot               bool
	AllowListExtensions    []string
	PathDenylist           []string
	LocationExcludePattern []string
	IgnoreGitIgnore        bool
	IgnoreIgnoreFile       bool
	IncludeHidden          bool
	IncludeBinaryFiles     bool
	IncludeMinified        bool
	MinifiedLineByteLength int
	MaxReadSizeBytes       int64

	// Output
	Format     string
	FileOutput string
	Verbose    bool

	// HTTP
	Address         string
	HttpServer      bool
	SearchTemplate  string
	DisplayTemplate string
}

// DefaultConfig returns a Config with sensible defaults matching the root-level globals.
func DefaultConfig() Config {
	return Config{
		SnippetLength:          300,
		SnippetCount:           1,
		SnippetMode:            "auto",
		Ranker:                 "bm25",
		ResultLimit:            -1,
		PathDenylist:           []string{".git", ".hg", ".svn"},
		MinifiedLineByteLength: 255,
		MaxReadSizeBytes:       1_000_000,
		Format:                 "text",
		Address:                ":8080",
	}
}
