package parser

import "testing"

func TestParser_Parse(t *testing.T) {
	lex := NewLexer(`test`)
	parser := NewParser(lex)

	parser.Parse()
}
