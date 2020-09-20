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

	if expr.Op != "TERM" {
		t.Error("expected TERM got", expr.Op)
	}

	if expr.Val != "test" {
		t.Error("expected test got", expr.Val)
	}
}


func TestParser_SimpleTest(t *testing.T) {
	lex := NewLexer(`simple test`)
	parser := NewParser(lex)
	parser.Parse()
}

//func TestParser_ParseAnd(t *testing.T) {
//	lex := NewLexer(`test stuff`)
//	parser := NewParser(lex)
//	expr := parser.Parse()
//
//	if expr.Op != "AND" {
//		t.Error("expected AND got", expr.Op)
//	}
//
//	if expr.Val != "" {
//		t.Error("expected '' got", expr.Val)
//	}
//
//	if expr.Left.Op != "TERM" {
//		t.Error("expected TERM got", expr.Left.Op)
//	}
//
//	if expr.Left.Val != "test" {
//		t.Error("expected test got", expr.Left.Val)
//	}
//
//	if expr.Right.Val != "stuff" {
//		t.Error("expected TERM got", expr.Right.Val)
//	}
//}
