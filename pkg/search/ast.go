package search

import (
	"fmt"
	"regexp"
	"sync"
)

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
	// Implicit is true when this OR was created by combining adjacent terms
	// under a default-or operator (rather than an explicit "OR" in the query).
	// Filter hoisting only reaches through implicit OR groupings.
	Implicit bool
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
//
// A pattern is compiled MULTI-LINE, so ^ and $ mean the start and end of a
// line: search runs over a whole file, and without (?m) "/^\s{8}- Name: /"
// could only ever match a file that began with it, which nobody writing ^
// means. \A and \z still anchor to the file. It also follows the search's
// case setting the way every other term does, so "/api_key/" finds API_KEY
// unless the search is case-sensitive. A pattern's own flags still apply:
// "(?-i)" inside it restores case.
type RegexNode struct {
	Pattern  string
	compiled *regexp.Regexp
	folded   *regexp.Regexp
	err      error
	once     sync.Once
	foldOnce sync.Once
}

// Compile compiles the pattern, caching the result, and reports why an invalid
// pattern could not be compiled so the caller can tell the user rather than
// letting it look like a query that simply matched nothing.
func (n *RegexNode) Compile() error {
	n.once.Do(func() {
		// Checked as written first, so an error quotes what the user typed
		// rather than a "(?m)" they never saw. A valid pattern stays valid
		// with a leading flag group.
		if _, n.err = regexp.Compile(n.Pattern); n.err == nil {
			n.compiled, n.err = regexp.Compile("(?m)" + n.Pattern)
		}
	})
	return n.err
}

// Regexp returns the case-sensitive compiled regexp, compiling once on first
// call. Returns nil if the pattern is invalid.
func (n *RegexNode) Regexp() *regexp.Regexp {
	_ = n.Compile()
	return n.compiled
}

// RegexpFor returns the regexp for a search with the given case setting.
// Returns nil if the pattern is invalid.
func (n *RegexNode) RegexpFor(caseSensitive bool) *regexp.Regexp {
	if caseSensitive || n.Compile() != nil {
		return n.Regexp()
	}
	n.foldOnce.Do(func() {
		n.folded, _ = regexp.Compile("(?mi)" + n.Pattern)
	})
	return n.folded
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
