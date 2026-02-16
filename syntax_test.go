// SPDX-License-Identifier: MIT

package main

import (
	"strings"
	"testing"
)

func TestTokenize_GoFuncLine(t *testing.T) {
	line := `func main() {`
	tokens := Tokenize(line)

	if len(tokens) == 0 {
		t.Fatal("expected tokens, got none")
	}

	// First token should be "func" keyword
	if tokens[0].Kind != TkKeyword {
		t.Errorf("expected first token to be TkKeyword, got %d", tokens[0].Kind)
	}
	if line[tokens[0].Start:tokens[0].End] != "func" {
		t.Errorf("expected 'func', got %q", line[tokens[0].Start:tokens[0].End])
	}

	// Should find "main" as an identifier (plain, lowercase)
	found := false
	for _, tok := range tokens {
		text := line[tok.Start:tok.End]
		if text == "main" {
			found = true
			if tok.Kind != TkPlain {
				t.Errorf("expected 'main' to be TkPlain, got %d", tok.Kind)
			}
		}
	}
	if !found {
		t.Error("did not find 'main' token")
	}
}

func TestTokenize_StringLiteral(t *testing.T) {
	line := `x := "hello world"`
	tokens := Tokenize(line)

	foundString := false
	for _, tok := range tokens {
		if tok.Kind == TkString {
			text := line[tok.Start:tok.End]
			if text == `"hello world"` {
				foundString = true
			}
		}
	}
	if !foundString {
		t.Error("expected to find string literal token")
	}
}

func TestTokenize_Comment(t *testing.T) {
	line := `x := 1 // a comment`
	tokens := Tokenize(line)

	foundComment := false
	for _, tok := range tokens {
		if tok.Kind == TkComment {
			foundComment = true
		}
	}
	if !foundComment {
		t.Error("expected to find comment token")
	}
}

func TestTokenize_Number(t *testing.T) {
	line := `y := 42 + 3.14`
	tokens := Tokenize(line)

	numCount := 0
	for _, tok := range tokens {
		if tok.Kind == TkNumber {
			numCount++
		}
	}
	if numCount < 2 {
		t.Errorf("expected at least 2 number tokens, got %d", numCount)
	}
}

func TestTokenize_UppercaseType(t *testing.T) {
	line := `var x MyType`
	tokens := Tokenize(line)

	found := false
	for _, tok := range tokens {
		text := line[tok.Start:tok.End]
		if text == "MyType" {
			found = true
			if tok.Kind != TkType {
				t.Errorf("expected 'MyType' to be TkType, got %d", tok.Kind)
			}
		}
	}
	if !found {
		t.Error("did not find 'MyType' token")
	}
}

func TestTokenize_CoversFullSource(t *testing.T) {
	line := `func Foo(x int) string { return "bar" }`
	tokens := Tokenize(line)

	// Verify tokens cover the entire source with no gaps
	if len(tokens) == 0 {
		t.Fatal("expected tokens")
	}
	if tokens[0].Start != 0 {
		t.Errorf("first token starts at %d, expected 0", tokens[0].Start)
	}
	if tokens[len(tokens)-1].End != len(line) {
		t.Errorf("last token ends at %d, expected %d", tokens[len(tokens)-1].End, len(line))
	}
	for i := 1; i < len(tokens); i++ {
		if tokens[i].Start != tokens[i-1].End {
			t.Errorf("gap between token %d (end=%d) and %d (start=%d)",
				i-1, tokens[i-1].End, i, tokens[i].Start)
		}
	}
}

func TestBuildKindArray_MatchOverridesSyntax(t *testing.T) {
	line := `func main()`
	tokens := Tokenize(line)
	matchLocs := [][]int{{0, 4}} // "func" match

	kinds := BuildKindArray(line, tokens, matchLocs)

	// Positions 0-3 should be TkMatch (overriding TkKeyword)
	for i := 0; i < 4; i++ {
		if kinds[i] != TkMatch {
			t.Errorf("position %d: expected TkMatch, got %d", i, kinds[i])
		}
	}

	// Position after "func " should NOT be TkMatch
	if kinds[5] == TkMatch {
		t.Error("position 5 should not be TkMatch")
	}
}

func TestBuildKindArray_EmptyLine(t *testing.T) {
	kinds := BuildKindArray("", nil, nil)
	if len(kinds) != 0 {
		t.Errorf("expected empty kinds for empty line, got %d", len(kinds))
	}
}

func TestRenderANSI_PlainText(t *testing.T) {
	line := "hello"
	kinds := make([]TokenKind, len(line)) // all TkPlain
	result := RenderANSI(line, kinds)
	if result != "hello" {
		t.Errorf("expected plain 'hello', got %q", result)
	}
}

func TestRenderANSI_WithKeyword(t *testing.T) {
	line := `func main`
	tokens := Tokenize(line)
	kinds := BuildKindArray(line, tokens, nil)
	result := RenderANSI(line, kinds)

	// Should contain ANSI escape for keyword
	if !strings.Contains(result, "\033[38;5;75m") {
		t.Error("expected keyword ANSI color in output")
	}
	if !strings.Contains(result, "func") {
		t.Error("expected 'func' in output")
	}
}

func TestRenderANSI_WithMatch(t *testing.T) {
	line := `func main`
	tokens := Tokenize(line)
	matchLocs := [][]int{{5, 9}} // "main"
	kinds := BuildKindArray(line, tokens, matchLocs)
	result := RenderANSI(line, kinds)

	// Should contain match ANSI (red bold)
	if !strings.Contains(result, "\033[1;31m") {
		t.Error("expected match ANSI color in output")
	}
}

func TestRenderANSILine_Convenience(t *testing.T) {
	line := `if x == 42 { return "yes" }`
	result := RenderANSILine(line, [][]int{{3, 5}})
	if !strings.Contains(result, "if") {
		t.Error("expected 'if' in output")
	}
	if !strings.Contains(result, "42") {
		t.Error("expected '42' in output")
	}
}

func TestRenderLipgloss_EmptyLine(t *testing.T) {
	result := RenderLipgloss("", nil, false)
	if result != "" {
		t.Errorf("expected empty string for empty line, got %q", result)
	}
}

func BenchmarkTokenize(b *testing.B) {
	line := `func extractRelevantV3(res *FileJob, documentTermFrequency map[string]int, snippetLength int) []Snippet {`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Tokenize(line)
	}
}

func BenchmarkRenderANSILine(b *testing.B) {
	line := `func extractRelevantV3(res *FileJob, documentTermFrequency map[string]int, snippetLength int) []Snippet {`
	matchLocs := [][]int{{5, 22}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderANSILine(line, matchLocs)
	}
}
