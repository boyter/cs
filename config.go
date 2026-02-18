// SPDX-License-Identifier: MIT

package main

import (
	"github.com/boyter/cs/pkg/ranker"
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

	// Structural ranker weights
	WeightCode    float64
	WeightComment float64
	WeightString  float64
	OnlyCode      bool
	OnlyComments  bool

	// Output
	Format     string
	FileOutput string
	Verbose    bool
	NoSyntax   bool // disable syntax highlighting

	// MCP
	MCPServer bool

	// HTTP
	Address         string
	HttpServer      bool
	SearchTemplate  string
	DisplayTemplate string
	TemplateStyle   string
}

// DefaultConfig returns a Config with sensible defaults matching the root-level globals.
func DefaultConfig() Config {
	defaults := ranker.DefaultStructuralConfig()
	return Config{
		SnippetLength:          300,
		SnippetCount:           1,
		SnippetMode:            "auto",
		Ranker:                 "bm25",
		ResultLimit:            -1,
		PathDenylist:           []string{".git", ".hg", ".svn"},
		MinifiedLineByteLength: 255,
		MaxReadSizeBytes:       1_000_000,
		WeightCode:             defaults.WeightCode,
		WeightComment:          defaults.WeightComment,
		WeightString:           defaults.WeightString,
		Format:                 "text",
		Address:                ":8080",
		TemplateStyle:          "dark",
	}
}

// StructuralRankerConfig returns a StructuralConfig populated from this Config.
func (c *Config) StructuralRankerConfig() *ranker.StructuralConfig {
	return &ranker.StructuralConfig{
		WeightCode:    c.WeightCode,
		WeightComment: c.WeightComment,
		WeightString:  c.WeightString,
		OnlyCode:      c.OnlyCode,
		OnlyComments:  c.OnlyComments,
	}
}
