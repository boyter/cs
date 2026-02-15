package search

import "fmt"

// Node is the interface that all nodes in the AST must implement.
type Node interface {
	// String is used for debugging and visualizing the AST.
	String() string
}

// AndNode represents a logical AND operation between two other nodes.
type AndNode struct {
	Left  Node
	Right Node
}

func (n *AndNode) String() string { return fmt.Sprintf("(%s AND %s)", n.Left, n.Right) }

// OrNode represents a logical OR operation between two other nodes.
type OrNode struct {
	Left  Node
	Right Node
}

func (n *OrNode) String() string { return fmt.Sprintf("(%s OR %s)", n.Left, n.Right) }

// NotNode represents a logical NOT operation on a single node.
type NotNode struct {
	Expr Node
}

func (n *NotNode) String() string { return fmt.Sprintf("NOT %s", n.Expr) }

// KeywordNode represents a simple keyword search term.
type KeywordNode struct {
	Value string
}

func (n *KeywordNode) String() string { return fmt.Sprintf("KEYWORD(%s)", n.Value) }

// PhraseNode represents a quoted phrase search term.
type PhraseNode struct {
	Value string
}

func (n *PhraseNode) String() string { return fmt.Sprintf("PHRASE(\"%s\")", n.Value) }

// RegexNode represents a regular expression search term.
type RegexNode struct {
	Pattern string
}

func (n *RegexNode) String() string { return fmt.Sprintf("REGEX(/%s/)", n.Pattern) }

// FilterNode represents a generic metadata filter.
type FilterNode struct {
	Field    string
	Operator string
	Value    interface{} // Can be a string, number, etc.
}

func (n *FilterNode) String() string {
	return fmt.Sprintf("FILTER(%s %s %v)", n.Field, n.Operator, n.Value)
}

// FuzzyNode represents a fuzzy search term with a maximum edit distance.
type FuzzyNode struct {
	Value    string
	Distance int
}

func (n *FuzzyNode) String() string {
	return fmt.Sprintf("FUZZY(%s~%d)", n.Value, n.Distance)
}
