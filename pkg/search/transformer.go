package search

import (
	"fmt"
	"strings"
)

// Transformer walks the AST and applies modifications.
type Transformer struct {
	notices []string
}

// TransformAST is the entry point for the transformation process.
func (t *Transformer) TransformAST(node Node) (Node, []string) {
	result := t.walk(node)
	return result, t.notices
}

func (t *Transformer) walk(node Node) Node {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *AndNode:
		n.Left = t.walk(n.Left)
		n.Right = t.walk(n.Right)
		return n
	case *OrNode:
		n.Left = t.walk(n.Left)
		n.Right = t.walk(n.Right)
		return n
	case *NotNode:
		n.Expr = t.walk(n.Expr)
		return n
	case *FilterNode:
		return t.transformFilterNode(n)
	default:
		return node // No transformation for Keyword, Phrase, etc.
	}
}

func (t *Transformer) transformFilterNode(node *FilterNode) Node {
	if node.Field == "complexity" && node.Operator == "=" {
		if val, ok := node.Value.(string); ok {
			if strings.ToLower(val) == "high" {
				// Replace this node with a new one
				newNode := &FilterNode{
					Field:    "complexity",
					Operator: ">=",
					Value:    8, // 'high' is defined as 8 or more
				}
				notice := fmt.Sprintf("Notice: '%s=%s' was interpreted as 'complexity >= 8'.", node.Field, val)
				t.notices = append(t.notices, notice)
				return newNode
			}
			// Can add more aliases like "medium" or "low" here
		}
	}
	return node
}
