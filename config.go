// SPDX-License-Identifier: MIT

package main

import (
	"strings"

	"github.com/boyter/cs/v3/pkg/common"
	"github.com/boyter/cs/v3/pkg/ranker"
	"github.com/boyter/cs/v3/pkg/snippet"
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
	TestPenalty   float64
	ResultLimit   int
	LineLimit     int

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
	WeightCode       float64
	WeightComment    float64
	WeightString     float64
	OnlyCode         bool
	OnlyComments     bool
	OnlyStrings      bool
	OnlyDeclarations bool
	OnlyUsages       bool

	// Grep context
	ContextBefore int // -B: lines before each match
	ContextAfter  int // -A: lines after each match
	ContextAround int // -C: lines before and after (sets both)

	// Output
	Format     string
	FileOutput string
	Verbose    bool
	NoSyntax   bool   // disable syntax highlighting
	Dedup      bool   // collapse byte-identical matches
	Color      string // color mode: auto, always, never
	Reverse    bool   // reverse result order (lowest score first)

	// Query complexity limits (0 = no limit)
	MaxQueryChars int
	MaxQueryTerms int

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
		TestPenalty:            0.4,
		ResultLimit:            -1,
		LineLimit:              -1,
		PathDenylist:           []string{".git", ".hg", ".svn"},
		MinifiedLineByteLength: 255,
		MaxReadSizeBytes:       1_000_000,
		WeightCode:             defaults.WeightCode,
		WeightComment:          defaults.WeightComment,
		WeightString:           defaults.WeightString,
		MaxQueryChars:          common.MaxQueryCharsDefault,
		MaxQueryTerms:          common.MaxQueryTermsDefault,
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
		OnlyStrings:   c.OnlyStrings,
	}
}

// HasContentFilter returns true if any content-type filter is active.
func (c *Config) HasContentFilter() bool {
	return c.OnlyCode || c.OnlyComments || c.OnlyStrings || c.OnlyDeclarations || c.OnlyUsages
}

// ContentFilterCachePrefix returns a cache key prefix for the active content filter.
func (c *Config) ContentFilterCachePrefix() string {
	switch {
	case c.OnlyCode:
		return "[[only-code]]"
	case c.OnlyComments:
		return "[[only-comments]]"
	case c.OnlyStrings:
		return "[[only-strings]]"
	case c.OnlyDeclarations:
		return "[[only-declarations]]"
	case c.OnlyUsages:
		return "[[only-usages]]"
	default:
		return ""
	}
}

// ResolveContext returns the effective before/after context line counts.
// -C sets both; -B/-A override individually (matching grep semantics).
func (c *Config) ResolveContext() (before, after int) {
	before = c.ContextAround
	after = c.ContextAround
	if c.ContextBefore > 0 {
		before = c.ContextBefore
	}
	if c.ContextAfter > 0 {
		after = c.ContextAfter
	}
	return
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
