package search

import (
	"errors"
	"path"
	"path/filepath"
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
	se.filterHandlers["path"] = handlePathFilter
	se.filterHandlers["filepath"] = handlePathFilter
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
		var results []*Document
		termLen := len(n.Value)
		for _, doc := range docs {
			s := string(doc.Content)
			if caseSensitive {
				if fuzzyContains(s, n.Value, n.Distance, termLen) {
					results = append(results, doc)
				}
			} else {
				if fuzzyContains(strings.ToLower(s), strings.ToLower(n.Value), n.Distance, termLen) {
					results = append(results, doc)
				}
			}
		}
		return results
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
// File/extension filters are evaluated against the filename; other filters
// (lang, complexity) pass through as true since metadata is not available.
func EvaluateFile(node Node, content []byte, filename string, location string, caseSensitive bool) (bool, map[string][][]int) {
	if node == nil {
		return true, nil
	}
	locations := make(map[string][][]int)
	matched := evalFile(node, content, filename, location, caseSensitive, locations)
	return matched, locations
}

func evalFile(node Node, content []byte, filename string, location string, caseSensitive bool, locations map[string][][]int) bool {
	if node == nil {
		return true
	}

	switch n := node.(type) {
	case *AndNode:
		if !evalFile(n.Left, content, filename, location, caseSensitive, locations) {
			return false
		}
		return evalFile(n.Right, content, filename, location, caseSensitive, locations)
	case *OrNode:
		left := evalFile(n.Left, content, filename, location, caseSensitive, locations)
		right := evalFile(n.Right, content, filename, location, caseSensitive, locations)
		return left || right
	case *NotNode:
		return !evalFile(n.Expr, content, filename, location, caseSensitive, locations)
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
		locs := re.FindAllIndex(content, -1)
		if len(locs) > 0 {
			locations[n.Pattern] = locs
			return true
		}
		return false
	case *FuzzyNode:
		s := string(content)
		searchContent := s
		searchTerm := n.Value
		if !caseSensitive {
			searchContent = strings.ToLower(s)
			searchTerm = strings.ToLower(n.Value)
		}
		locs := fuzzyFind(searchContent, searchTerm, n.Distance, len(n.Value))
		if len(locs) > 0 {
			locations[n.Value] = locs
			return true
		}
		return false
	case *FilterNode:
		return evalFileFilter(n, filename, location)
	}

	return false
}

// evalFileFilter evaluates a FilterNode against a filename/location in per-file mode.
// Filters that require document metadata not available per-file (lang, complexity)
// pass through as true.
func evalFileFilter(n *FilterNode, filename string, location string) bool {
	filterVal, ok := n.Value.(string)
	if !ok {
		return true
	}

	field := strings.ToLower(n.Field)
	switch field {
	case "file", "filename":
		var match bool
		if containsGlobMeta(filterVal) {
			match = matchGlob(filterVal, filename)
		} else {
			match = strings.Contains(strings.ToLower(filename), strings.ToLower(filterVal))
		}
		if n.Operator == "!=" {
			return !match
		}
		return match
	case "ext", "extension":
		ext := strings.TrimPrefix(filepath.Ext(filename), ".")
		match := strings.EqualFold(ext, filterVal)
		if n.Operator == "!=" {
			return !match
		}
		return match
	case "path", "filepath":
		var match bool
		if containsGlobMeta(filterVal) {
			match = matchPathGlob(filterVal, location)
		} else {
			match = strings.Contains(strings.ToLower(location), strings.ToLower(filterVal))
		}
		if n.Operator == "!=" {
			return !match
		}
		return match
	default:
		// lang, language, complexity, etc. — metadata not available in per-file mode
		return true
	}
}

// --- Glob matching helpers ---

// containsGlobMeta reports whether s contains any glob metacharacters (*, ?, [).
func containsGlobMeta(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// matchGlob performs a case-insensitive glob match using path.Match.
// Returns false on malformed patterns (no panic).
func matchGlob(pattern, name string) bool {
	matched, err := path.Match(strings.ToLower(pattern), strings.ToLower(name))
	if err != nil {
		return false
	}
	return matched
}

// matchPathGlob matches a multi-segment glob pattern against a full path using
// a sliding-window approach over /-separated segments.
func matchPathGlob(pattern, fullPath string) bool {
	patSegs := strings.Split(strings.ToLower(pattern), "/")
	pathSegs := strings.Split(strings.ToLower(fullPath), "/")

	if len(patSegs) > len(pathSegs) {
		return false
	}

	for i := 0; i <= len(pathSegs)-len(patSegs); i++ {
		allMatch := true
		for j, ps := range patSegs {
			matched, err := path.Match(ps, pathSegs[i+j])
			if err != nil || !matched {
				allMatch = false
				break
			}
		}
		if allMatch {
			return true
		}
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

	var match bool
	if containsGlobMeta(filename) {
		match = matchGlob(filename, doc.Filename)
	} else {
		match = strings.Contains(strings.ToLower(doc.Filename), strings.ToLower(filename))
	}

	switch op {
	case "=":
		return match
	case "!=":
		return !match
	}
	return false
}

func handlePathFilter(op string, val interface{}, doc *Document) bool {
	p, ok := val.(string)
	if !ok {
		return false
	}

	var match bool
	if containsGlobMeta(p) {
		match = matchPathGlob(p, doc.Path)
	} else {
		match = strings.Contains(strings.ToLower(doc.Path), strings.ToLower(p))
	}

	switch op {
	case "=":
		return match
	case "!=":
		return !match
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

// --- Fuzzy matching helpers ---

// levenshtein computes the Levenshtein edit distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Use single row of DP table
	prev := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr := make([]int, lb+1)
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = min3(del, ins, sub)
		}
		prev = curr
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// fuzzyContains checks if any same-length substring of content matches
// the term within the given edit distance (substitution-based matching).
func fuzzyContains(content, term string, maxDist, termLen int) bool {
	contentLen := len(content)
	if contentLen == 0 || termLen == 0 || termLen > contentLen {
		return false
	}

	for i := 0; i <= contentLen-termLen; i++ {
		window := content[i : i+termLen]
		if levenshtein(window, term) <= maxDist {
			return true
		}
	}
	return false
}

// fuzzyFind finds all match locations in content that are within the given
// edit distance of the term. Returns [][]int where each entry is [start, end].
func fuzzyFind(content, term string, maxDist, termLen int) [][]int {
	contentLen := len(content)
	if contentLen == 0 || termLen == 0 || termLen > contentLen {
		return nil
	}

	var locs [][]int
	for i := 0; i <= contentLen-termLen; i++ {
		window := content[i : i+termLen]
		if levenshtein(window, term) <= maxDist {
			locs = append(locs, []int{i, i + termLen})
		}
	}
	return locs
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
