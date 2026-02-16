// SPDX-License-Identifier: MIT

package main

import (
	"strings"
	"text/scanner"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// TokenKind classifies a byte in a source line for syntax highlighting.
type TokenKind int

const (
	TkPlain       TokenKind = iota
	TkKeyword               // language keyword
	TkString                // string / char literal
	TkComment               // comment
	TkNumber                // integer or float literal
	TkType                  // uppercase-initial identifier (heuristic)
	TkPunctuation           // operators, braces, etc.
	TkWhitespace            // spaces, tabs
	TkMatch                 // search match — always wins over syntax
)

// Token represents a classified span in source text.
type Token struct {
	Kind  TokenKind
	Start int // byte offset (inclusive)
	End   int // byte offset (exclusive)
}

// keywords is a language-independent union keyword map covering Go, Python,
// Java, JavaScript, TypeScript, C, C++, C#, Ruby, Rust, PHP, Swift, Kotlin,
// Scala, and more. Intentionally over-broad: a Python keyword will highlight
// in a Java file too. This matches the existing HTML highlighter behavior.
var keywords = map[string]struct{}{
	// Control flow
	"if": {}, "else": {}, "for": {}, "while": {}, "do": {},
	"switch": {}, "case": {}, "break": {}, "continue": {}, "default": {},
	"return": {}, "goto": {}, "fallthrough": {},

	// Boolean / nil / null
	"true": {}, "false": {}, "nil": {}, "null": {}, "None": {},
	"True": {}, "False": {}, "undefined": {},

	// Go
	"func": {}, "package": {}, "import": {}, "defer": {}, "go": {},
	"chan": {}, "select": {}, "range": {}, "interface": {}, "struct": {},
	"map": {}, "type": {}, "var": {}, "const": {},

	// Python
	"def": {}, "class": {}, "lambda": {}, "yield": {}, "from": {},
	"with": {}, "as": {}, "pass": {}, "raise": {}, "except": {},
	"assert": {}, "global": {}, "nonlocal": {}, "del": {},
	"elif": {}, "is": {}, "in": {}, "not": {}, "and": {}, "or": {},

	// Java / C# / OOP
	"new": {}, "delete": {}, "this": {}, "self": {}, "super": {},
	"throw": {}, "throws": {}, "catch": {}, "try": {}, "finally": {},
	"static": {}, "public": {}, "private": {}, "protected": {},
	"abstract": {}, "final": {}, "override": {}, "extends": {},
	"implements": {}, "enum": {}, "void": {}, "instanceof": {},
	"synchronized": {}, "volatile": {}, "transient": {}, "native": {},
	"strictfp": {},

	// JavaScript / TypeScript
	"async": {}, "await": {}, "export": {}, "require": {},
	"let": {}, "function": {},
	"typeof": {}, "of": {}, "debugger": {},
	"declare": {}, "namespace": {}, "module": {},

	// C / C++
	"auto": {}, "register": {}, "extern": {}, "signed": {}, "unsigned": {},
	"sizeof": {}, "typedef": {}, "union": {}, "inline": {},
	"template": {}, "typename": {}, "virtual": {}, "explicit": {},
	"friend": {}, "mutable": {}, "operator": {}, "using": {},
	"constexpr": {}, "noexcept": {}, "nullptr": {}, "static_cast": {},
	"dynamic_cast": {}, "reinterpret_cast": {}, "const_cast": {},

	// Rust
	"fn": {}, "mut": {}, "impl": {}, "trait": {},
	"pub": {}, "mod": {}, "use": {}, "crate": {}, "where": {},
	"move": {}, "ref": {}, "match": {}, "loop": {}, "unsafe": {},
	"dyn": {}, "box": {},

	// Ruby
	"begin": {}, "end": {}, "rescue": {}, "ensure": {}, "then": {},
	"unless": {}, "until": {}, "defined?": {},
	"attr_reader": {}, "attr_writer": {}, "attr_accessor": {},
	"include": {}, "prepend": {}, "extend": {},

	// PHP
	"echo": {}, "isset": {}, "unset": {}, "foreach": {},
	"elseif": {}, "endif": {}, "endfor": {}, "endforeach": {},
	"endwhile": {}, "endswitch": {},

	// Swift
	"guard": {}, "associatedtype": {}, "protocol": {},
	"convenience": {}, "required": {}, "weak": {}, "unowned": {},
	"fileprivate": {}, "internal": {}, "open": {},
	"willSet": {}, "didSet": {},

	// Kotlin
	"val": {}, "when": {}, "object": {}, "companion": {},
	"data": {}, "sealed": {}, "inner": {}, "crossinline": {},
	"noinline": {}, "reified": {}, "suspend": {},
	"init": {}, "constructor": {},

	// Scala
	"lazy": {}, "implicit": {}, "forSome": {},

	// Common type keywords
	"int": {}, "float": {}, "double": {}, "char": {}, "bool": {},
	"long": {}, "short": {}, "byte": {}, "string": {},
	"boolean": {},
}

// Tokenize splits source into classified tokens using text/scanner.
// Gaps between scanner tokens (e.g. # comments) are emitted as TkPlain.
func Tokenize(source string) []Token {
	var s scanner.Scanner
	s.Init(strings.NewReader(source))
	s.Whitespace = 0                                // don't skip whitespace
	s.Mode ^= scanner.SkipComments                  // return comments as tokens
	s.Error = func(_ *scanner.Scanner, _ string) {} // suppress errors

	var tokens []Token
	lastEnd := 0

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		start := s.Offset
		text := s.TokenText()
		end := start + len(text)

		// Fill gap between last token end and this token start
		if start > lastEnd {
			tokens = append(tokens, Token{Kind: TkPlain, Start: lastEnd, End: start})
		}

		kind := classifyToken(tok, text)
		tokens = append(tokens, Token{Kind: kind, Start: start, End: end})
		lastEnd = end
	}

	// Fill trailing gap
	if lastEnd < len(source) {
		tokens = append(tokens, Token{Kind: TkPlain, Start: lastEnd, End: len(source)})
	}

	return tokens
}

// classifyToken determines the TokenKind for a scanner token.
func classifyToken(tok rune, text string) TokenKind {
	switch tok {
	case scanner.Ident:
		if _, ok := keywords[text]; ok {
			return TkKeyword
		}
		if len(text) > 0 && unicode.IsUpper(rune(text[0])) {
			return TkType
		}
		return TkPlain

	case scanner.Int, scanner.Float:
		return TkNumber

	case scanner.String, scanner.Char, scanner.RawString:
		return TkString

	case scanner.Comment:
		return TkComment

	default:
		if text == " " || text == "\t" || text == "\n" || text == "\r" {
			return TkWhitespace
		}
		return TkPunctuation
	}
}

// BuildKindArray creates a per-byte TokenKind array for a line.
// Syntax tokens are stamped first, then match locations override as TkMatch.
func BuildKindArray(line string, tokens []Token, matchLocs [][]int) []TokenKind {
	kinds := make([]TokenKind, len(line))
	// Default is TkPlain (zero value)

	// Stamp syntax tokens
	for _, t := range tokens {
		start := t.Start
		end := t.End
		if start < 0 {
			start = 0
		}
		if end > len(line) {
			end = len(line)
		}
		for i := start; i < end; i++ {
			kinds[i] = t.Kind
		}
	}

	// Stamp match locations last — always wins
	for _, loc := range matchLocs {
		if len(loc) < 2 {
			continue
		}
		start := loc[0]
		end := loc[1]
		if start < 0 {
			start = 0
		}
		if end > len(line) {
			end = len(line)
		}
		for i := start; i < end; i++ {
			kinds[i] = TkMatch
		}
	}

	return kinds
}

// ANSI 256-color escape sequences for each token kind.
var ansiStyles = map[TokenKind]string{
	TkKeyword:     "\033[38;5;75m",  // cornflower blue
	TkString:      "\033[38;5;114m", // soft green
	TkComment:     "\033[38;5;243m", // medium gray
	TkNumber:      "\033[38;5;176m", // light purple
	TkType:        "\033[38;5;80m",  // teal/cyan
	TkPunctuation: "\033[38;5;250m", // light gray
	TkMatch:       "\033[1;31m",     // red bold (existing)
}

const ansiReset = "\033[0m"

// RenderANSI renders a line with ANSI color codes based on the per-byte kind array.
func RenderANSI(line string, kinds []TokenKind) string {
	if len(line) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(line) * 2) // rough estimate

	segStart := 0
	prevKind := kinds[0]

	for i := 1; i <= len(line); i++ {
		var curKind TokenKind
		if i < len(line) {
			curKind = kinds[i]
		}

		if i == len(line) || curKind != prevKind {
			seg := line[segStart:i]
			if style, ok := ansiStyles[prevKind]; ok {
				b.WriteString(style)
				b.WriteString(seg)
				b.WriteString(ansiReset)
			} else {
				b.WriteString(seg)
			}
			segStart = i
			prevKind = curKind
		}
	}

	return b.String()
}

// Lipgloss styles for TUI syntax highlighting.
var (
	syntaxKeywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	syntaxStringStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	syntaxCommentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	syntaxNumberStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("176"))
	syntaxTypeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("80"))
	syntaxPunctStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	// Selected variants (with background)
	syntaxKeywordSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Background(lipgloss.Color("236"))
	syntaxStringSelectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("114")).Background(lipgloss.Color("236"))
	syntaxCommentSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Background(lipgloss.Color("236"))
	syntaxNumberSelectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("176")).Background(lipgloss.Color("236"))
	syntaxTypeSelectedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("80")).Background(lipgloss.Color("236"))
	syntaxPunctSelectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Background(lipgloss.Color("236"))
)

// lipglossStyle returns the appropriate lipgloss style for a token kind.
func lipglossStyle(kind TokenKind, isSelected bool) lipgloss.Style {
	if isSelected {
		switch kind {
		case TkKeyword:
			return syntaxKeywordSelectedStyle
		case TkString:
			return syntaxStringSelectedStyle
		case TkComment:
			return syntaxCommentSelectedStyle
		case TkNumber:
			return syntaxNumberSelectedStyle
		case TkType:
			return syntaxTypeSelectedStyle
		case TkPunctuation:
			return syntaxPunctSelectedStyle
		case TkMatch:
			return selectedMatchStyle
		default:
			return selectedSnippetStyle
		}
	}

	switch kind {
	case TkKeyword:
		return syntaxKeywordStyle
	case TkString:
		return syntaxStringStyle
	case TkComment:
		return syntaxCommentStyle
	case TkNumber:
		return syntaxNumberStyle
	case TkType:
		return syntaxTypeStyle
	case TkPunctuation:
		return syntaxPunctStyle
	case TkMatch:
		return matchStyle
	default:
		return snippetStyle
	}
}

// RenderLipgloss renders a line with lipgloss styles based on the per-byte kind array.
func RenderLipgloss(line string, kinds []TokenKind, isSelected bool) string {
	if len(line) == 0 {
		if isSelected {
			return selectedSnippetStyle.Render("")
		}
		return ""
	}

	var b strings.Builder

	segStart := 0
	prevKind := kinds[0]

	for i := 1; i <= len(line); i++ {
		var curKind TokenKind
		if i < len(line) {
			curKind = kinds[i]
		}

		if i == len(line) || curKind != prevKind {
			seg := line[segStart:i]
			style := lipglossStyle(prevKind, isSelected)
			b.WriteString(style.Render(seg))
			segStart = i
			prevKind = curKind
		}
	}

	return b.String()
}

// RenderANSILine is a convenience that tokenizes, builds the kind array, and renders ANSI.
func RenderANSILine(line string, matchLocs [][]int) string {
	tokens := Tokenize(line)
	kinds := BuildKindArray(line, tokens, matchLocs)
	return RenderANSI(line, kinds)
}

// RenderLipglossLine is a convenience that tokenizes, builds the kind array, and renders lipgloss.
func RenderLipglossLine(line string, matchLocs [][]int, isSelected bool) string {
	tokens := Tokenize(line)
	kinds := BuildKindArray(line, tokens, matchLocs)
	return RenderLipgloss(line, kinds, isSelected)
}
