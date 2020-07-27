package parser

import "testing"


func TestParser_Empty(t *testing.T) {
	lex := NewLexer(``)
	parser := NewParser(lex)
	expr := parser.Parse()

	if expr != nil {
		t.Error("expected nil")
	}
}


func TestParser_Parse(t *testing.T) {
	lex := NewLexer(`test`)
	parser := NewParser(lex)
	expr := parser.Parse()

	if expr == nil {
		t.Error("should be something")
	}

	if expr.Op != "TERM" {
		t.Error("expected TERM got", expr.Op)
	}

	if expr.Val != "test" {
		t.Error("expected test got", expr.Val)
	}
}
