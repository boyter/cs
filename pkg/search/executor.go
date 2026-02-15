package search

import (
	"errors"
	"regexp"
	"strings"

	str "github.com/boyter/go-string"
)

// ErrInvalidQuery is returned when a query is syntactically incorrect and cannot be healed.
var ErrInvalidQuery = errors.New("invalid or unparseable query")

// FilterHandler is a function that checks whether a single document matches a filter.
type FilterHandler func(op string, val interface{}, doc *Document) bool

// SearchEngine holds the data and configuration for searching.
type SearchEngine struct {
	documents      []*Document
	filterHandlers map[string]FilterHandler
}

// NewSearchEngine creates a new engine and initializes it.
func NewSearchEngine(docs []*Document) *SearchEngine {
	se := &SearchEngine{documents: docs}
	se.registerFilterHandlers()
	return se
}

func (se *SearchEngine) registerFilterHandlers() {
	se.filterHandlers = make(map[string]FilterHandler)
	se.filterHandlers["complexity"] = handleComplexityFilter
	se.filterHandlers["lang"] = handleLanguageFilter
	se.filterHandlers["language"] = handleLanguageFilter
	se.filterHandlers["file"] = handleFilenameFilter
	se.filterHandlers["filename"] = handleFilenameFilter
	se.filterHandlers["ext"] = handleExtensionFilter
	se.filterHandlers["extension"] = handleExtensionFilter
}

// Search is the main public method to run a query.
func (se *SearchEngine) Search(query string, caseSensitive bool) (*SearchResult, error) {
	// 1. Parse
	lexer := NewLexer(strings.NewReader(query))
	parser := NewParser(lexer)
	ast, notices := parser.ParseQuery()
	if ast == nil && query != "" {
		return nil, ErrInvalidQuery
	}

	// 2. Transform
	transformer := &Transformer{}
	ast, transformNotices := transformer.TransformAST(ast)
	notices = append(notices, transformNotices...)

	// 3. Plan
	ast = PlanAST(ast)

	// 4. Execute
	var results []*Document
	if ast != nil {
		results = se.evaluate(ast, se.documents, caseSensitive)
	}

	// 5. Extract terms for highlighting
	termsToHighlight := ExtractTerms(ast)

	return &SearchResult{
		Documents:        results,
		Notices:          notices,
		TermsToHighlight: termsToHighlight,
	}, nil
}

// evaluate is the entry point for the recursive AST execution.
func (se *SearchEngine) evaluate(node Node, docs []*Document, caseSensitive bool) []*Document {
	if node == nil {
		return docs
	}

	switch n := node.(type) {
	case *AndNode:
		leftResults := se.evaluate(n.Left, docs, caseSensitive)
		return se.evaluate(n.Right, leftResults, caseSensitive)
	case *OrNode:
		leftResults := se.evaluate(n.Left, docs, caseSensitive)
		rightResults := se.evaluate(n.Right, docs, caseSensitive)
		return union(leftResults, rightResults)
	case *NotNode:
		toExclude := se.evaluate(n.Expr, docs, caseSensitive)
		return difference(docs, toExclude)
	case *KeywordNode:
		var results []*Document
		for _, doc := range docs {
			var match bool
			if caseSensitive {
				match = strings.Contains(string(doc.Content), n.Value)
			} else {
				match = len(str.IndexAllIgnoreCase(string(doc.Content), n.Value, -1)) > 0
			}
			if match {
				results = append(results, doc)
			}
		}
		return results
	case *PhraseNode:
		var results []*Document
		for _, doc := range docs {
			var match bool
			if caseSensitive {
				match = strings.Contains(string(doc.Content), n.Value)
			} else {
				match = len(str.IndexAllIgnoreCase(string(doc.Content), n.Value, -1)) > 0
			}
			if match {
				results = append(results, doc)
			}
		}
		return results
	case *RegexNode:
		var results []*Document
		re, err := regexp.Compile(n.Pattern)
		if err != nil {
			return []*Document{}
		}
		for _, doc := range docs {
			if re.Match(doc.Content) {
				results = append(results, doc)
			}
		}
		return results
	case *FuzzyNode:
		// Placeholder: fuzzy matching not yet implemented
		return []*Document{}
	case *FilterNode:
		if handler, ok := se.filterHandlers[n.Field]; ok {
			var results []*Document
			for _, doc := range docs {
				if handler(n.Operator, n.Value, doc) {
					results = append(results, doc)
				}
			}
			return results
		}
		return []*Document{}
	}

	return []*Document{}
}

// EvaluateFile evaluates a parsed AST against a single file's content.
// It returns whether the file matches and a map of term → match locations.
// Filter nodes pass through as true since metadata is not available in per-file mode.
func EvaluateFile(node Node, content []byte, filename string, caseSensitive bool) (bool, map[string][][]int) {
	if node == nil {
		return true, nil
	}
	locations := make(map[string][][]int)
	matched := evalFile(node, content, caseSensitive, locations)
	return matched, locations
}

func evalFile(node Node, content []byte, caseSensitive bool, locations map[string][][]int) bool {
	if node == nil {
		return true
	}

	switch n := node.(type) {
	case *AndNode:
		if !evalFile(n.Left, content, caseSensitive, locations) {
			return false
		}
		return evalFile(n.Right, content, caseSensitive, locations)
	case *OrNode:
		left := evalFile(n.Left, content, caseSensitive, locations)
		right := evalFile(n.Right, content, caseSensitive, locations)
		return left || right
	case *NotNode:
		return !evalFile(n.Expr, content, caseSensitive, locations)
	case *KeywordNode:
		s := string(content)
		if caseSensitive {
			if strings.Contains(s, n.Value) {
				locs := str.IndexAll(s, n.Value, -1)
				if len(locs) > 0 {
					locations[n.Value] = locs
				}
				return true
			}
			return false
		}
		locs := str.IndexAllIgnoreCase(s, n.Value, -1)
		if len(locs) > 0 {
			locations[n.Value] = locs
			return true
		}
		return false
	case *PhraseNode:
		s := string(content)
		if caseSensitive {
			if strings.Contains(s, n.Value) {
				locs := str.IndexAll(s, n.Value, -1)
				if len(locs) > 0 {
					locations[n.Value] = locs
				}
				return true
			}
			return false
		}
		locs := str.IndexAllIgnoreCase(s, n.Value, -1)
		if len(locs) > 0 {
			locations[n.Value] = locs
			return true
		}
		return false
	case *RegexNode:
		re, err := regexp.Compile(n.Pattern)
		if err != nil {
			return false
		}
		loc := re.FindIndex(content)
		if loc != nil {
			locations[n.Pattern] = [][]int{loc}
			return true
		}
		return false
	case *FuzzyNode:
		// Placeholder: fuzzy matching not yet implemented
		return false
	case *FilterNode:
		// Filters pass through as true in per-file mode (no metadata available)
		return true
	}

	return false
}

// --- Filter Handlers ---

func handleComplexityFilter(op string, val interface{}, doc *Document) bool {
	complexity, ok := val.(int)
	if !ok {
		return false
	}
	complexity64 := int64(complexity)

	switch op {
	case "=":
		return doc.Complexity == complexity64
	case "!=":
		return doc.Complexity != complexity64
	case ">=":
		return doc.Complexity >= complexity64
	case "<=":
		return doc.Complexity <= complexity64
	}
	return false
}

func handleLanguageFilter(op string, val interface{}, doc *Document) bool {
	isEquality := (op == "=")

	switch v := val.(type) {
	case string:
		lang := strings.ToLower(v)
		return (strings.ToLower(doc.Language) == lang) == isEquality
	case []interface{}:
		valueSet := make(map[string]bool)
		for _, item := range v {
			if strItem, ok := item.(string); ok {
				valueSet[strings.ToLower(strItem)] = true
			}
		}
		_, exists := valueSet[strings.ToLower(doc.Language)]
		return exists == isEquality
	}
	return false
}

func handleFilenameFilter(op string, val interface{}, doc *Document) bool {
	filename, ok := val.(string)
	if !ok {
		return false
	}
	filename = strings.ToLower(filename)

	switch op {
	case "=":
		return strings.ToLower(doc.Filename) == filename
	case "!=":
		return strings.ToLower(doc.Filename) != filename
	}
	return false
}

func handleExtensionFilter(op string, val interface{}, doc *Document) bool {
	isEquality := (op == "=")

	switch v := val.(type) {
	case string:
		ext := strings.ToLower(v)
		return (strings.ToLower(doc.Extension) == ext) == isEquality
	case []interface{}:
		valueSet := make(map[string]bool)
		for _, item := range v {
			if strItem, ok := item.(string); ok {
				valueSet[strings.ToLower(strItem)] = true
			}
		}
		_, exists := valueSet[strings.ToLower(doc.Extension)]
		return exists == isEquality
	}
	return false
}

// --- Set helpers ---

func union(a, b []*Document) []*Document {
	pathMap := make(map[string]bool)
	var result []*Document
	for _, doc := range a {
		if !pathMap[doc.Path] {
			pathMap[doc.Path] = true
			result = append(result, doc)
		}
	}
	for _, doc := range b {
		if !pathMap[doc.Path] {
			pathMap[doc.Path] = true
			result = append(result, doc)
		}
	}
	return result
}

func difference(all, toExclude []*Document) []*Document {
	excludeMap := make(map[string]bool)
	for _, doc := range toExclude {
		excludeMap[doc.Path] = true
	}
	var result []*Document
	for _, doc := range all {
		if !excludeMap[doc.Path] {
			result = append(result, doc)
		}
	}
	return result
}
