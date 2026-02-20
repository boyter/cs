package search

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

// Parser creates an AST from a stream of tokens.
type Parser struct {
	l          *Lexer
	tok        Token
	peekTok    Token
	notices    []string
	parenDepth int
}

// NewParser creates a new Parser.
func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.tok = p.peekTok
	p.peekTok = p.l.scan()
	for p.peekTok.Type == WS {
		p.peekTok = p.l.scan()
	}
	slog.Debug("nextToken", "current", p.tok.Literal, "peek", p.peekTok.Literal)
}

// ParseQuery is the entry point for parsing.
func (p *Parser) ParseQuery() (Node, []string) {
	if p.tok.Type == EOF {
		return nil, nil // Handle empty query
	}

	node := p.parseExpression(0)

	// Final checks after parsing is complete
	if p.tok.Type == AND || p.tok.Type == OR {
		p.notices = append(p.notices, fmt.Sprintf("Warning: Trailing '%s' was ignored.", p.tok.Literal))
	}
	if p.tok.Type == RPAREN {
		p.notices = append(p.notices, "Notice: Unmatched ')' was ignored.")
	}
	if p.parenDepth > 0 {
		p.notices = append(p.notices, "Notice: Missing ')' was added to the end of the query.")
	}

	return node, p.notices
}

// getPrecedence returns the precedence for the current token.
// This is the heart of the Pratt parser.
func (p *Parser) getPrecedence() int {
	switch p.tok.Type {
	case OR:
		return 1
	case AND:
		return 2
	// These tokens can imply an AND if they appear after another expression.
	case IDENTIFIER, PHRASE, REGEX, FUZZY, LPAREN, NOT:
		return 3
	default:
		return 0 // No precedence for other tokens like EOF, RPAREN
	}
}

func (p *Parser) parseExpression(precedence int) Node {
	slog.Debug("parseExpression", "precedence", precedence)
	left := p.parsePrefix()
	if left == nil {
		return nil
	}

	// This is the core Pratt parser loop.
	// It continues as long as the current token has a higher precedence than the context.
	for precedence < p.getPrecedence() {
		switch p.tok.Type {
		case AND, OR:
			// *** THIS IS THE FIX ***
			// If we find an operator, but the next token indicates the end of an
			// expression (EOF or the end of a group), it's a dangling operator.
			// We must not proceed, as it will cause an infinite loop. Instead, we
			// return the expression we have so far, leaving the dangling operator
			// to be handled by the final checks in ParseQuery.
			if p.peekTok.Type == EOF || p.peekTok.Type == RPAREN {
				return left
			}
			left = p.parseInfixExpression(left)
		case IDENTIFIER, PHRASE, REGEX, FUZZY, LPAREN, NOT: // An expression-starting token here means an implicit AND.
			left = p.parseImplicitAndExpression(left)
		default:
			// If the token can't be part of an infix expression, stop.
			return left
		}
	}
	return left
}

func isFilterField(s string) bool {
	switch strings.ToLower(s) {
	case "file", "filename", "ext", "extension", "lang", "language", "complexity", "path", "filepath":
		return true
	}
	return false
}

func (p *Parser) parsePrefix() Node {
	var node Node
	switch p.tok.Type {
	case IDENTIFIER:
		// Check for colon filter syntax: "file:test", "ext:go", etc.
		if colonIdx := strings.Index(p.tok.Literal, ":"); colonIdx > 0 {
			field := p.tok.Literal[:colonIdx]
			value := p.tok.Literal[colonIdx+1:]
			if isFilterField(field) {
				if value != "" {
					node = &FilterNode{Field: field, Operator: "=", Value: value}
					p.nextToken() // Consume the identifier
					return node
				}
				// Colon-filter with empty value followed by operator: complexity:<=25
				// Strip the trailing colon so parseFilterExpression gets clean field name
				if p.peekTok.Type == OPERATOR {
					p.tok.Literal = field
					node = p.parseFilterExpression()
					return node
				}
			}
		}
		if p.peekTok.Type == OPERATOR {
			node = p.parseFilterExpression()
			return node // Filter expression consumes its own tokens
		}
		// If an IDENTIFIER is not followed by an operator, it's treated as a keyword.
		node = &KeywordNode{Value: p.tok.Literal}
	case NUMBER:
		node = &KeywordNode{Value: p.tok.Literal}
	case PHRASE:
		node = &PhraseNode{Value: p.tok.Literal}
	case REGEX:
		node = &RegexNode{Pattern: p.tok.Literal}
	case FUZZY:
		// Parse "term~N" into FuzzyNode
		literal := p.tok.Literal
		idx := strings.LastIndex(literal, "~")
		if idx <= 0 {
			return nil
		}
		term := literal[:idx]
		dist, _ := strconv.Atoi(literal[idx+1:])
		node = &FuzzyNode{Value: term, Distance: dist}
	case NOT:
		p.nextToken()                // Consume 'NOT'
		expr := p.parseExpression(5) // High precedence for what NOT applies to
		if expr == nil {
			// This indicates a dangling NOT operator (e.g., "cat AND NOT")
			// which is a syntax error.
			return nil
		}
		node = &NotNode{Expr: expr}
		return node
	case LPAREN:
		p.parenDepth++
		p.nextToken() // Consume '('
		node = p.parseExpression(0)
		if p.tok.Type == RPAREN {
			p.parenDepth--
			p.nextToken() // Consume ')'
		}
		return node
	case RPAREN:
		p.notices = append(p.notices, "Notice: Unmatched ')' was ignored.")
		p.nextToken() // Consume the bad token
		return p.parsePrefix()
	default:
		return nil
	}

	p.nextToken() // Consume the token that was just parsed as a prefix
	return node
}

// We will modify parseFilterExpression
func (p *Parser) parseFilterExpression() Node {
	node := &FilterNode{Field: p.tok.Literal}
	p.nextToken() // Consume field
	node.Operator = p.tok.Literal
	p.nextToken() // Consume operator

	// --- START OF MODIFIED LOGIC ---

	// Check if the next token is a value or the start of a list of values.
	// This check is tricky because a leading comma is valid (e.g., ",go,py").
	// So we just start the loop.

	// Collect all values. It might be one, or it might be a list.
	var values []interface{}
	for {
		// Add the current value to our list if it's a valid value type.
		switch p.tok.Type {
		case NUMBER:
			val, _ := strconv.Atoi(p.tok.Literal)
			values = append(values, val)
			p.nextToken() // Consume the value token
		case STRING_ALIAS, IDENTIFIER, PHRASE:
			values = append(values, p.tok.Literal)
			p.nextToken() // Consume the value token
		}

		// If the next token is not a comma, we are done with the list.
		if p.tok.Type != COMMA {
			break
		}
		p.nextToken() // Consume the comma and loop again for the next value.
	}

	// After attempting to parse values, if we didn't find any,
	// it's a syntax error (e.g., "lang=" or "lang>>5").
	if len(values) == 0 {
		return nil // This will trigger an ErrInvalidQuery
	}

	// If we only found one value, store it directly to maintain compatibility
	// with simple filters like complexity=5. If we found multiple, store the slice.
	if len(values) == 1 {
		node.Value = values[0]
	} else {
		node.Value = values
	}

	// --- END OF MODIFIED LOGIC ---

	return node
}

func (p *Parser) parseInfixExpression(left Node) Node {
	tokType := p.tok.Type
	precedence := p.getPrecedence()
	p.nextToken() // Consume the operator (e.g., AND)
	right := p.parseExpression(precedence)

	if right == nil {
		// This case handles situations like `(cat AND)`. We've consumed the AND,
		// but there's nothing valid after it. We cannot form a valid infix expression.
		// By returning only `left`, we effectively discard the operator.
		// The calling `parseExpression` will then terminate, and the final checks
		// in `ParseQuery` will not see the consumed operator. This is a conscious
		// choice to favor graceful failure over producing a notice for this specific edge case.
		return left
	}

	if tokType == AND {
		return &AndNode{Left: left, Right: right}
	}
	return &OrNode{Left: left, Right: right}
}

func (p *Parser) parseImplicitAndExpression(left Node) Node {
	precedence := p.getPrecedence()
	// DO NOT consume a token here because there is no operator.
	right := p.parseExpression(precedence)
	return &AndNode{Left: left, Right: right}
}
