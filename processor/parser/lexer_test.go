// SPDX-License-Identifier: MIT OR Unlicense

package parser

import (
	"testing"
)

func TestNext(t *testing.T) {
	lex := NewLexer(`test`)

	for lex.Next() != 0 {
		if 0 > 1 {
			t.Error("wot the")
		}
	}
}

func TestPeek(t *testing.T) {
	lex := NewLexer(`test`)

	for i := 0; i < 100; i++ {
		lex.Peek()
	}
}

func TestNextEnd(t *testing.T) {
	lex := NewLexer(``)

	token := lex.NextToken()
	if token.Type != "END" {
		t.Error(`expected END got`, token.Type)
	}
}

func TestNextTokenParenOpen(t *testing.T) {
	lex := NewLexer(`(`)

	token := lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}
}

func TestNextTokenParenClose(t *testing.T) {
	lex := NewLexer(`)`)

	token := lex.NextToken()
	if token.Type != "PAREN_CLOSE" {
		t.Error(`expected PAREN_CLOSE got`, token.Type)
	}
}

func TestNextTokenParenOpenParenClose(t *testing.T) {
	lex := NewLexer(`()`)

	token := lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "PAREN_CLOSE" {
		t.Error(`expected PAREN_CLOSE got`, token.Type)
	}
}

func TestNextTokenQuote(t *testing.T) {
	lex := NewLexer(`"`)

	token := lex.NextToken()
	if token.Type != "QUOTED_TERM" {
		t.Error(`expected QUOTED_TERM got`, token.Type)
	}
}

func TestNextTokenMultipleEmptyQuote(t *testing.T) {
	lex := NewLexer(`("")`)

	token := lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "QUOTED_TERM" {
		t.Error(`expected QUOTED_TERM got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "PAREN_CLOSE" {
		t.Error(`expected PAREN_CLOSE got`, token.Type)
	}
}

func TestNextTokenIgnoresSpace(t *testing.T) {
	lex := NewLexer(` (`)

	token := lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}
}

func TestNextTokenQuotedTerm(t *testing.T) {
	lex := NewLexer(`"test"`)

	token := lex.NextToken()
	if token.Type != "QUOTED_TERM" {
		t.Error(`expected QUOTED_TERM got`, token.Type)
	}

	if token.Value != `test` {
		t.Error("expected test got", token.Value)
	}
}

func TestNextTokenQuotedTermNoEnd(t *testing.T) {
	lex := NewLexer(`"test`)

	token := lex.NextToken()
	if token.Type != "QUOTED_TERM" {
		t.Error(`expected QUOTED_TERM got`, token.Type)
	}

	if token.Value != `test` {
		t.Error("expected test got", token.Value)
	}
}

func TestNextTokenQuotedTermMultiple(t *testing.T) {
	lex := NewLexer(`"test" "something"`)

	token := lex.NextToken()
	if token.Type != "QUOTED_TERM" {
		t.Error(`expected QUOTED_TERM got`, token.Type)
	}

	if token.Value != `test` {
		t.Error("expected test got", token.Value)
	}

	token = lex.NextToken()
	if token.Type != "QUOTED_TERM" {
		t.Error(`expected QUOTED_TERM got`, token.Type)
	}

	if token.Value != `something` {
		t.Error("expected something got", token.Value)
	}
}

func TestNextTokenMultipleSomethingQuote(t *testing.T) {
	lex := NewLexer(`("test")`)

	token := lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "QUOTED_TERM" {
		t.Error(`expected QUOTED_TERM got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "PAREN_CLOSE" {
		t.Error(`expected PAREN_CLOSE got`, token.Type)
	}
}

func TestNextTokenMultipleEverythingQuote(t *testing.T) {
	lex := NewLexer(`("test") ("test")`)

	token := lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "QUOTED_TERM" {
		t.Error(`expected QUOTED_TERM got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "PAREN_CLOSE" {
		t.Error(`expected PAREN_CLOSE got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "QUOTED_TERM" {
		t.Error(`expected QUOTED_TERM got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "PAREN_CLOSE" {
		t.Error(`expected PAREN_CLOSE got`, token.Type)
	}
}
