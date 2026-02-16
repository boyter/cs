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
