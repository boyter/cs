package parser

// For fuzz testing...
// https://github.com/dvyukov/go-fuzz
// install both go-fuzz-build and go-fuzz
// go-fuzz-build && go-fuzz
func Fuzz(data []byte) int {
	lex := NewLexer(string(data))

	lex.Peek()
	t := lex.NextToken()
	for t.Type != "END" {
		t = lex.NextToken()
	}

	return 1
}
