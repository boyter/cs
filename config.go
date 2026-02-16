// SPDX-License-Identifier: MIT

package main

import (
	"github.com/boyter/cs/pkg/snippet"
	"github.com/boyter/gocodewalker"
)

// resolveSnippetMode returns the effective snippet mode for a file.
// If globalMode is "auto", it selects based on the file extension.
func resolveSnippetMode(globalMode, filename string) string {
	if globalMode != "auto" {
		return globalMode
	}
	return snippet.SnippetModeForExtension(gocodewalker.GetExtension(filename))
}

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
	LanguageTypes          []string
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
	NoSyntax   bool // disable syntax highlighting

	// HTTP
	Address         string
	HttpServer      bool
	SearchTemplate  string
	DisplayTemplate string
	TemplateStyle   string
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
		TemplateStyle:          "dark",
	}
}
