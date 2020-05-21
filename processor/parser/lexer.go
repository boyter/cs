package parser

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

// This is for the parser
type Expr struct {
	Op string
	Left *Expr
	Right *Expr
	Val []string
}

// This is for the lexer
type Token struct {
	Type string
	Pos int
}


type Lexer struct{
	Query string
	pos int
}

func NewLexer(query string) Lexer {
	return Lexer{
		Query: query,
	}
}

func (p *Lexer) Tokens() []Token {
	return nil
}

func (p *Lexer) nextToken() Token {
	// based on the pos find the next token location
	switch c := p.Query[p.pos]; c {
	case '(':
		return Token{
			Type: "PAREN_START",
			Pos:  p.pos,
		}
	case '"':

	}

	if p.Query[p.pos] == '"' {
		// scan from here till we fine the next or the end and return as the token
		for i, r := range p.Query[p.pos:] {
			if i > p.pos {
				if r == '"' {
					tok := p.Query[p.pos:i+1]
					p.pos = i+1
					return tok
				}
			}
		}
	}

	return p.Query[p.pos:]
}
