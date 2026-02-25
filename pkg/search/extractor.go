package search

// termExtractor is a visitor that walks the AST to find all positive terms
// that should be highlighted in the search results.
type termExtractor struct {
	terms        []string
	inNotContext bool
}

// ExtractTerms is the public entry point that traverses the AST and returns
// a slice of strings that should be highlighted.
func ExtractTerms(node Node) []string {
	visitor := &termExtractor{}
	visitor.walk(node)
	return visitor.terms
}

func (v *termExtractor) walk(node Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *AndNode:
		v.walk(n.Left)
		v.walk(n.Right)
	case *OrNode:
		v.walk(n.Left)
		v.walk(n.Right)
	case *NotNode:
		// Set the context to true for the subtree of this NOT node
		originalContext := v.inNotContext
		v.inNotContext = true
		v.walk(n.Expr)
		v.inNotContext = originalContext // Restore context
	case *KeywordNode:
		if !v.inNotContext {
			v.terms = append(v.terms, n.Value)
		}
	case *PhraseNode:
		if !v.inNotContext {
			v.terms = append(v.terms, n.Value)
		}
	case *RegexNode:
		if !v.inNotContext {
			v.terms = append(v.terms, n.Pattern)
		}
	case *FuzzyNode:
		if !v.inNotContext {
			v.terms = append(v.terms, n.Value)
		}
	// FilterNode is ignored as its values are not for content highlighting.
	case *FilterNode:
		// Do nothing
	}
}

// CountAllTerms counts all unique leaf-node values in the AST, including
// terms inside NOT and filter values. This is used for query complexity
// validation, unlike ExtractTerms which only returns positive terms for
// highlighting.
func CountAllTerms(node Node) int {
	seen := make(map[string]struct{})
	countWalk(node, seen)
	return len(seen)
}

func countWalk(node Node, seen map[string]struct{}) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *AndNode:
		countWalk(n.Left, seen)
		countWalk(n.Right, seen)
	case *OrNode:
		countWalk(n.Left, seen)
		countWalk(n.Right, seen)
	case *NotNode:
		countWalk(n.Expr, seen)
	case *KeywordNode:
		seen[n.Value] = struct{}{}
	case *PhraseNode:
		seen[n.Value] = struct{}{}
	case *RegexNode:
		seen[n.Pattern] = struct{}{}
	case *FuzzyNode:
		seen[n.Value] = struct{}{}
	case *FilterNode:
		if s, ok := n.Value.(string); ok {
			seen[s] = struct{}{}
		}
	}
}
