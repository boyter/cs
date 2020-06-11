// SPDX-License-Identifier: MIT OR Unlicense

package parser

import (
	"strings"
	"unicode"
)

// Requirements
// Should parse boolean search queries and get something out
// that we can use for searching
// Should be forgiving of errors EG
// 		(this OR something
// the above should not fail, it should assume that the user was going to add the ) and just has not yet
// Example query that uses everything
//		(this OR that) NOT (yes AND "no way") stuff~1
// the above should search for this OR that in documents that don't have yes and "no way" AND fuzzy search for stuff

// first tokenise and turn into tree
// https://groups.google.com/forum/#!topic/golang-nuts/phpOipu03G0
// https://stackoverflow.com/questions/1312986/how-to-make-a-logical-boolean-parser-for-text-input
// https://godoc.org/github.com/cznic/goyacc
// https://github.com/Meyhem/go-simple-expression-eval
// https://about.sourcegraph.com/go/gophercon-2018-how-to-write-a-parser-in-go
// http://web.eecs.utk.edu/~azh/blog/teenytinycompiler1.html
// http://web.eecs.utk.edu/~azh/blog/teenytinycompiler2.html
// https://ruslanspivak.com/lsbasi-part1/

// https://www.aaronraff.dev/blog/how-to-write-a-lexer-in-go

// This is for the parser
type Expr struct {
	Op    string
	Left  *Expr
	Right *Expr
	Val   []string
}

// This is for the lexer
type Token struct {
	Type     string
	StartPos int
	Value    string
}

type Lexer struct {
	query string
	pos   int
}

func NewLexer(query string) Lexer {
	return Lexer{
		query: query,
	}
}

// Peek at the next byte
func (l *Lexer) Peek() byte {
	if l.pos < len(l.query) {
		return l.query[l.pos]
	}

	// return null byte when at the end
	return 0
}

// Return the next byte
func (l *Lexer) Next() byte {
	if l.pos < len(l.query) {
		l.pos++
		return l.query[l.pos-1]
	}

	// return null byte when at the end
	return 0
}

func (l *Lexer) NextToken() Token {

	// at the end so return end token
	if l.Peek() == 0 {
		return Token{
			Type:     "END",
			StartPos: l.pos,
			Value:    "",
		}
	}

	// skip whitespace
	for unicode.IsSpace(rune(l.Peek())) {
		l.Next()
	}

	switch c := l.Next(); c {
	case '(':
		return Token{
			Type:     "PAREN_OPEN",
			StartPos: l.pos,
			Value:    "(",
		}
	case ')':
		return Token{
			Type:     "PAREN_CLOSE",
			StartPos: l.pos,
			Value:    ")",
		}
	case '"':
		// loop till we hit another " or the end
		var sb strings.Builder
		for l.Peek() != '"' && l.Peek() != 0 {
			sb.WriteByte(l.Next())
		}

		if l.Peek() == '"' {
			l.Next()
		}

		return Token{
			Type:     "QUOTED_TERM",
			StartPos: l.pos,
			Value:    sb.String(),
		}
	}

	return Token{}
}

func (l *Lexer) Tokens() []Token {
	return nil
}
