package search

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

// TokenType represents the type of a token.
type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	WS // Whitespace

	PHRASE       // "hello world"
	REGEX        // /[cb]at/
	OPERATOR     // >=, <=, =, !=
	LPAREN       // (
	RPAREN       // )
	AND          // AND, and
	OR           // OR, or
	NOT          // NOT, not
	IDENTIFIER   // complexity, author, #define
	NUMBER       // 5, 10
	STRING_ALIAS // high, low
	COMMA
	FUZZY // term~1, term~2
)

// Token represents a single token from the query string.
type Token struct {
	Type    TokenType
	Literal string
}

// Lexer scans the input string and produces tokens.
type Lexer struct {
	r *bufio.Reader
}

// NewLexer creates a new Lexer.
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{r: bufio.NewReader(r)}
}

// isWhitespace checks if a rune is a whitespace character.
func isWhitespace(ch rune) bool {
	return unicode.IsSpace(ch)
}

// isSpecialSyntaxChar checks if a rune is a character that always has a special meaning
// and thus cannot be part of an identifier.
func isSpecialSyntaxChar(ch rune) bool {
	switch ch {
	case '(', ')', '"', ',', '=', '!', '>', '<', '/':
		return true
	default:
		return false
	}
}

// scan is the main token-producing function.
func (l *Lexer) scan() Token {
	ch := l.read()

	if isWhitespace(ch) {
		l.unread()
		return l.scanWhitespace()
	}

	switch ch {
	case eof:
		return Token{Type: EOF, Literal: ""}
	case '"':
		return l.scanPhrase()
	case '/':
		return l.scanRegex()
	case '(':
		return Token{Type: LPAREN, Literal: string(ch)}
	case ')':
		return Token{Type: RPAREN, Literal: string(ch)}
	case ',':
		return Token{Type: COMMA, Literal: string(ch)}
	case '=', '!', '>', '<':
		l.unread()
		return l.scanOperator()
	default:
		l.unread()
		return l.scanIdentifier()
	}
}

func (l *Lexer) scanWhitespace() Token {
	var sb strings.Builder
	for {
		ch := l.read()
		if ch == eof || !isWhitespace(ch) {
			l.unread()
			break
		}
		sb.WriteRune(ch)
	}
	return Token{Type: WS, Literal: sb.String()}
}

// scanIdentifier consumes characters until a delimiter is found and returns the appropriate token.
func (l *Lexer) scanIdentifier() Token {
	var sb strings.Builder
	for {
		ch := l.read()
		if ch == eof || isWhitespace(ch) || isSpecialSyntaxChar(ch) {
			l.unread()
			break
		}
		sb.WriteRune(ch)
	}
	literal := sb.String()

	upper := strings.ToUpper(literal)
	switch upper {
	case "AND":
		return Token{Type: AND, Literal: literal}
	case "OR":
		return Token{Type: OR, Literal: literal}
	case "NOT":
		return Token{Type: NOT, Literal: literal}
	}

	if isAlias(literal) {
		return Token{Type: STRING_ALIAS, Literal: literal}
	}

	if isNumber(literal) {
		return Token{Type: NUMBER, Literal: literal}
	}

	// Check for fuzzy syntax: term~1 or term~2
	if idx := strings.LastIndex(literal, "~"); idx > 0 {
		suffix := literal[idx+1:]
		if suffix == "1" || suffix == "2" {
			return Token{Type: FUZZY, Literal: literal}
		}
	}

	if literal == "" {
		// This can happen if the query is just a special character, e.g., ">"
		// The switch in scan() will catch it and call scanOperator, but if it's
		// somehow missed, we return ILLEGAL.
		return Token{Type: ILLEGAL, Literal: ""}
	}

	return Token{Type: IDENTIFIER, Literal: literal}
}

func isAlias(s string) bool {
	switch strings.ToLower(s) {
	case "high", "medium", "low":
		return true
	}
	return false
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (l *Lexer) scanPhrase() Token {
	var sb strings.Builder
	for {
		ch := l.read()
		if ch == eof || ch == '"' {
			break
		}
		sb.WriteRune(ch)
	}
	return Token{Type: PHRASE, Literal: sb.String()}
}

func (l *Lexer) scanRegex() Token {
	var sb strings.Builder
	for {
		ch := l.read()
		if ch == eof || ch == '/' {
			break
		}
		sb.WriteRune(ch)
	}
	return Token{Type: REGEX, Literal: sb.String()}
}

func (l *Lexer) scanOperator() Token {
	ch1 := l.read()
	ch2 := l.read()

	if ch1 == '!' && ch2 == '=' {
		return Token{Type: OPERATOR, Literal: "!="}
	}
	if ch1 == '>' && ch2 == '=' {
		return Token{Type: OPERATOR, Literal: ">="}
	}
	if ch1 == '<' && ch2 == '=' {
		return Token{Type: OPERATOR, Literal: "<="}
	}

	l.unread()
	return Token{Type: OPERATOR, Literal: string(ch1)}
}

var eof = rune(0)

func (l *Lexer) read() rune {
	ch, _, err := l.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (l *Lexer) unread() { _ = l.r.UnreadRune() }
