// SPDX-License-Identifier: MIT OR Unlicense

package parser

import (
	"testing"
)

func TestNext(t *testing.T) {
	lex := NewLexer(`test`)

	for lex.Next() != 0 {
		if 0 > 1 { // we just want to ensure this does not crash hence the oddness
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
	lex := NewLexer(`                         (`)

	token := lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}
}

func TestNextTokenIgnoresSpaceMultiple(t *testing.T) {
	lex := NewLexer(`                         (                         )                         `)

	token := lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "PAREN_CLOSE" {
		t.Error(`expected PAREN_CLOSE got`, token.Type)
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

func TestNextTokenQuotedTermSpace(t *testing.T) {
	lex := NewLexer(`"test things"`)

	token := lex.NextToken()
	if token.Type != "QUOTED_TERM" {
		t.Error(`expected QUOTED_TERM got`, token.Type)
	}

	if token.Value != `test things` {
		t.Error("expected test things got", token.Value)
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

func TestNextTokenTerm(t *testing.T) {
	lex := NewLexer(`something`)
	token := lex.NextToken()

	if token.Type != "TERM" {
		t.Error(`expected TERM got`, token.Type)
	}

	if token.Value != "something" {
		t.Error(`expected something got`, token.Value)
	}
}

func TestNextTokenMultipleTerm(t *testing.T) {
	lex := NewLexer(`something else`)
	token := lex.NextToken()

	if token.Type != "TERM" {
		t.Error(`expected TERM got`, token.Type)
	}

	if token.Value != "something" {
		t.Error(`expected something got`, token.Value)
	}

	token = lex.NextToken()

	if token.Type != "TERM" {
		t.Error(`expected TERM got`, token.Type)
	}

	if token.Value != "else" {
		t.Error(`expected else got`, token.Value)
	}
}

func TestNextTokenAnd(t *testing.T) {
	lex := NewLexer(`AND`)
	token := lex.NextToken()

	if token.Type != "AND" {
		t.Error(`expected AND got`, token.Type)
	}
}

func TestNextTokenOr(t *testing.T) {
	lex := NewLexer(`OR`)
	token := lex.NextToken()

	if token.Type != "OR" {
		t.Error(`expected OR got`, token.Type)
	}
}

func TestNextTokenNot(t *testing.T) {
	lex := NewLexer(`NOT`)
	token := lex.NextToken()

	if token.Type != "NOT" {
		t.Error(`expected NOT got`, token.Type)
	}
}

func TestNextTokenMultipleTermOperators(t *testing.T) {
	lex := NewLexer(`something AND else`)
	token := lex.NextToken()

	if token.Type != "TERM" {
		t.Error(`expected TERM got`, token.Type)
	}

	if token.Value != "something" {
		t.Error(`expected something got`, token.Value)
	}

	token = lex.NextToken()

	if token.Type != "AND" {
		t.Error(`expected AND got`, token.Type)
	}

	token = lex.NextToken()

	if token.Type != "TERM" {
		t.Error(`expected TERM got`, token.Type)
	}

	if token.Value != "else" {
		t.Error(`expected else got`, token.Value)
	}
}

func TestNextTokenMultiple(t *testing.T) {
	lex := NewLexer(`(something AND else) OR (other NOT this)`)

	token := lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "TERM" {
		t.Error(`expected TERM got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "AND" {
		t.Error(`expected AND got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "TERM" {
		t.Error(`expected TERM got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "PAREN_CLOSE" {
		t.Error(`expected PAREN_CLOSE got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "OR" {
		t.Error(`expected OR got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "PAREN_OPEN" {
		t.Error(`expected PAREN_OPEN got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "TERM" {
		t.Error(`expected TERM got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "NOT" {
		t.Error(`expected NOT got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "TERM" {
		t.Error(`expected TERM got`, token.Type)
	}

	token = lex.NextToken()
	if token.Type != "PAREN_CLOSE" {
		t.Error(`expected PAREN_CLOSE got`, token.Type)
	}
}

func TestTokens(t *testing.T) {
	lex := NewLexer(`(something AND else) OR (other NOT this)`)

	tokens := lex.Tokens()

	if len(tokens) != 11 {
		t.Error("expected 11 tokens got", len(tokens))
	}
}
