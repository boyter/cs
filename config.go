// SPDX-License-Identifier: MIT

package main

import (
	"strings"

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
	GravityIntent string
	NoiseIntent   string
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
		GravityIntent:          "default",
		NoiseIntent:            "default",
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

// ResolveNoiseSensitivity maps the NoiseIntent string to a numeric sensitivity value
// used by the signal-to-noise penalty.
func (c *Config) ResolveNoiseSensitivity() float64 {
	switch strings.ToLower(c.NoiseIntent) {
	case "silence":
		return 0.1
	case "quiet":
		return 0.5
	case "default", "":
		return 1.0
	case "loud":
		return 2.0
	case "raw":
		return 100.0
	default:
		return 1.0
	}
}

// ResolveGravityStrength maps the GravityIntent string to a numeric strength value.
func (c *Config) ResolveGravityStrength() float64 {
	switch strings.ToLower(c.GravityIntent) {
	case "brain":
		return 2.5
	case "logic":
		return 1.5
	case "default", "":
		return 1.0
	case "low":
		return 0.2
	case "off":
		return 0.0
	default:
		return 1.0
	}
}
