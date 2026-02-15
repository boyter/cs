package search

import "sort"

// Cost values for different node types. Lower cost is executed first.
const (
	filterCost  = 10
	keywordCost = 20
	phraseCost  = 25
	regexCost   = 50
	orCost      = 100
	notCost     = 110
)

// PlanAST optimizes the AST for execution.
func PlanAST(node Node) Node {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *AndNode:
		// Flatten nested AND nodes
		children := flattenAnd(n)

		// Recursively plan children
		for i, child := range children {
			children[i] = PlanAST(child)
		}

		// Sort children by their execution cost
		sort.Slice(children, func(i, j int) bool {
			return getCost(children[i]) < getCost(children[j])
		})

		// Rebuild the AND tree from the sorted children
		return buildAndTree(children)
	case *OrNode:
		n.Left = PlanAST(n.Left)
		n.Right = PlanAST(n.Right)
		return n
	case *NotNode:
		n.Expr = PlanAST(n.Expr)
		return n
	default:
		return node
	}
}

// getCost assigns a cost to a node.
func getCost(n Node) int {
	switch n.(type) {
	case *FilterNode:
		return filterCost
	case *KeywordNode:
		return keywordCost
	case *PhraseNode:
		return phraseCost
	case *RegexNode:
		return regexCost
	case *FuzzyNode:
		return regexCost
	case *OrNode:
		return orCost
	case *NotNode:
		return notCost
	default:
		return 1000 // High cost for unknown
	}
}

// flattenAnd collects all children of a nested AND structure.
func flattenAnd(n *AndNode) []Node {
	var children []Node
	if left, ok := n.Left.(*AndNode); ok {
		children = append(children, flattenAnd(left)...)
	} else {
		children = append(children, n.Left)
	}
	if right, ok := n.Right.(*AndNode); ok {
		children = append(children, flattenAnd(right)...)
	} else {
		children = append(children, n.Right)
	}
	return children
}

// buildAndTree reconstructs a balanced AND tree from a list of nodes.
func buildAndTree(nodes []Node) Node {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}
	// Build tree from right to left
	tree := &AndNode{Left: nodes[len(nodes)-2], Right: nodes[len(nodes)-1]}
	for i := len(nodes) - 3; i >= 0; i-- {
		tree = &AndNode{Left: nodes[i], Right: tree}
	}
	return tree
}
