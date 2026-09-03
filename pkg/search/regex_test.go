// SPDX-License-Identifier: MIT

package search

import (
	"strings"
	"testing"
)

// TestScanRegexDelimiter pins that a '/' can be expressed inside a /regex/ query,
// by escaping it or by putting it in a character class, rather than truncating
// the pattern into an uncompilable fragment plus stray keyword nodes.
func TestScanRegexDelimiter(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string // expected AST rendering
	}{
		{"plain", `/health/`, `REGEX(/health/)`},
		{"escaped slash", `/_cluster\/health/`, `REGEX(/_cluster\/health/)`},
		{"slash in class", `/_cluster[/]health/`, `REGEX(/_cluster[/]health/)`},
		{"leading escaped slash", `/\/health/`, `REGEX(/\/health/)`},
		{"literal bracket then slash in class", `/[]/]x/`, `REGEX(/[]/]x/)`},
		{"negated class with literal bracket", `/[^]/]x/`, `REGEX(/[^]/]x/)`},
		{"escaped backslash before delimiter", `/a\\/`, `REGEX(/a\\/)`},
		{"unterminated", `/foo`, `REGEX(/foo/)`},
		{"unterminated trailing backslash", `/foo\`, `REGEX(/foo\/)`},
		{"class spanning to eof", `/foo[/`, `REGEX(/foo[//)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, notices := NewParser(NewLexer(strings.NewReader(tt.query))).ParseQuery()
			if node == nil {
				t.Fatalf("ParseQuery(%q) returned nil node, notices=%v", tt.query, notices)
			}
			if got := node.String(); got != tt.want {
				t.Errorf("ParseQuery(%q) = %s, want %s", tt.query, got, tt.want)
			}
		})
	}
}

// TestScanRegexDelimiterMatches pins that the patterns above are not merely
// preserved but actually compile and match, since RE2 accepts \/ and [/] as-is.
func TestScanRegexDelimiterMatches(t *testing.T) {
	tests := []struct {
		query   string
		subject string
	}{
		{`/_cluster\/health/`, "GET _cluster/health"},
		{`/_cluster[/]health/`, "GET _cluster/health"},
		{`/\/health/`, "GET /health"},
		{`/[]/]x/`, "/x"},
		{`/[]/]x/`, "]x"},
	}

	for _, tt := range tests {
		node, notices := NewParser(NewLexer(strings.NewReader(tt.query))).ParseQuery()
		regexNode, ok := node.(*RegexNode)
		if !ok {
			t.Fatalf("ParseQuery(%q) = %v (%T), want a single *RegexNode", tt.query, node, node)
		}
		if len(notices) != 0 {
			t.Errorf("ParseQuery(%q) notices = %v, want none", tt.query, notices)
		}
		re := regexNode.Regexp()
		if re == nil {
			t.Fatalf("ParseQuery(%q) produced pattern %q which did not compile: %v", tt.query, regexNode.Pattern, regexNode.Compile())
		}
		if !re.MatchString(tt.subject) {
			t.Errorf("pattern %q from query %q did not match %q", regexNode.Pattern, tt.query, tt.subject)
		}
	}
}

// TestFilterValueSlashUnaffected guards the separate scanIdentifier path that
// lets '/' through inside a colon filter value.
func TestFilterValueSlashUnaffected(t *testing.T) {
	node, notices := NewParser(NewLexer(strings.NewReader(`path:pkg/search`))).ParseQuery()
	if len(notices) != 0 {
		t.Errorf("notices = %v, want none", notices)
	}
	if got, want := node.String(), `FILTER(path = pkg/search)`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// TestInvalidRegexNotice pins that an uncompilable pattern is reported rather
// than silently behaving like a query that matched nothing. Asserting only the
// (still zero) match count would pass against the defect this covers.
func TestInvalidRegexNotice(t *testing.T) {
	tests := []struct {
		query      string
		wantReason string // substring of regexp.Compile's own error text
	}{
		{`/(unclosed/`, "missing closing )"},
		{`/[unclosed/`, "missing closing ]"},
		{`/_cluster[a-z/`, "missing closing ]"},
		{`/a**/`, "invalid nested repetition operator"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			node, notices := NewParser(NewLexer(strings.NewReader(tt.query))).ParseQuery()

			regexNode, ok := node.(*RegexNode)
			if !ok {
				t.Fatalf("ParseQuery(%q) = %v (%T), want a single *RegexNode", tt.query, node, node)
			}
			// The executor's nil guards stay correct; the change is that the
			// caller is told, not that matching behaves differently.
			if regexNode.Regexp() != nil {
				t.Errorf("Regexp() = non-nil for invalid pattern %q, want nil", regexNode.Pattern)
			}
			if regexNode.Compile() == nil {
				t.Errorf("Compile() = nil error for invalid pattern %q", regexNode.Pattern)
			}

			if len(notices) == 0 {
				t.Fatalf("ParseQuery(%q) returned no notices, so an invalid regex is indistinguishable from zero matches", tt.query)
			}
			joined := strings.Join(notices, "\n")
			if !strings.Contains(joined, tt.wantReason) {
				t.Errorf("notices %q do not include the reason %q", joined, tt.wantReason)
			}
			if !strings.Contains(joined, regexNode.Pattern) {
				t.Errorf("notices %q do not name the offending pattern %q", joined, regexNode.Pattern)
			}
		})
	}
}

// TestValidRegexNoNotice makes sure the notice above is not emitted for patterns
// that are fine, including the delimiter cases fixed in the lexer.
func TestValidRegexNoNotice(t *testing.T) {
	for _, query := range []string{`/health/`, `/_cluster\/health/`, `/_cluster[/]health/`, `/\/health/`, `/[]/]x/`, `/a\\/`, `/foo`} {
		_, notices := NewParser(NewLexer(strings.NewReader(query))).ParseQuery()
		if len(notices) != 0 {
			t.Errorf("ParseQuery(%q) notices = %v, want none", query, notices)
		}
	}
}
