// SPDX-License-Identifier: MIT OR Unlicense

package snippet

// proseExtensions maps file extensions (without dot) that are considered
// prose/free-text to true. Everything else is treated as code.
var proseExtensions = map[string]bool{
	"md":        true,
	"markdown":  true,
	"txt":       true,
	"text":      true,
	"rst":       true,
	"adoc":      true,
	"asciidoc":  true,
	"html":      true,
	"htm":       true,
	"xml":       true,
	"svg":       true,
	"csv":       true,
	"tsv":       true,
	"log":       true,
	"org":       true,
	"tex":       true,
	"latex":     true,
	"wiki":      true,
	"rdoc":      true,
	"textile":   true,
	"pod":       true,
	"man":       true,
	"roff":      true,
}

// SnippetModeForExtension returns the snippet mode appropriate for a given
// file extension (without the leading dot). Prose files get "snippet" (free-text),
// code files get "lines" (line-based with line numbers).
func SnippetModeForExtension(ext string) string {
	if proseExtensions[ext] {
		return "snippet"
	}
	return "lines"
}
