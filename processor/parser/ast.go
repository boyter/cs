package parser

import (
	"strings"
)

// first tokenise and turn into tree
//https://groups.google.com/forum/#!topic/golang-nuts/phpOipu03G0
type Expr struct {
	Op string
	Left *Expr
	Right *Expr
	Val []string
}


type Parser struct{
	Query string
	pos int
}

func NewParser(query string) Parser {
	return Parser{
		Query: query,
	}
}

func (p *Parser) nextToken() string {
	// based on the pos find the next token location
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

// need a peek here to look at the next token

func (p *Parser) Parse(query string) Expr {
	return Expr{
		Op:    "AND",
		Left:  nil,
		Right: nil,
		Val:   strings.Split(query, " "),
	}
}
