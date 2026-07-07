package search

// HoistFilters lifts metadata filters (path:, ext:, lang:, file:, complexity:,
// etc.) out of implicit OR groupings so that they constrain the whole query
// rather than being OR'd with individual terms.
//
// This matters when the default operator is OR: a query like
//
//	foo bar path:pkg
//
// parses as "foo OR bar OR path:pkg", but users expect the path filter to apply
// to the entire query, i.e. "(foo OR bar) AND path:pkg". Hoisting rewrites it to
// exactly that.
//
// Only implicit OR nodes (adjacent terms with no explicit operator) are
// traversed. An explicit "foo OR path:x", a filter inside a NOT, and any AND
// grouping are treated as opaque and left untouched, preserving the user's
// stated intent. When the default operator is AND there are no implicit OR
// nodes, so this pass returns the AST unchanged.
func HoistFilters(node Node) Node {
	if node == nil {
		return nil
	}
	content, filters := hoistFilters(node)
	for _, f := range filters {
		if content == nil {
			content = f
		} else {
			content = &AndNode{Left: content, Right: f}
		}
	}
	return content
}

// hoistFilters recursively separates an implicit-OR subtree into its non-filter
// content and the filters pulled out of it. A bare FilterNode becomes a hoisted
// filter; any other node (explicit OR/AND groups, NOT, keywords, ...) is opaque
// content that is not descended into.
func hoistFilters(node Node) (content Node, filters []Node) {
	switch n := node.(type) {
	case *OrNode:
		if !n.Implicit {
			return n, nil
		}
		lc, lf := hoistFilters(n.Left)
		rc, rf := hoistFilters(n.Right)
		filters = append(lf, rf...)
		return orJoin(lc, rc), filters
	case *FilterNode:
		return nil, []Node{n}
	default:
		return node, nil
	}
}

// orJoin combines two (possibly nil) content nodes with an implicit OR.
func orJoin(a, b Node) Node {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	default:
		return &OrNode{Left: a, Right: b, Implicit: true}
	}
}
