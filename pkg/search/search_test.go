// TODO: This file is part of an ongoing test-driven development session.
//
// Progress Summary:
// 1. Multi-Value Filters: Added tests for comma-separated filter values (e.g., `lang=go,python`).
//    - Decision: Normalizing language aliases (e.g., 'py' to 'Python') is the responsibility
//      of an upstream process, not the search engine itself. A comment was added to the
//      executor to clarify this.
//
// 2. Malformed Queries: Added tests for invalid queries that should be handled gracefully.
//    - Implementation: Created a specific `ErrInvalidQuery` for queries that are syntactically
//      incorrect and cannot be healed by the parser (e.g., a query containing only "AND").
//      This makes the API more robust for callers.
//
// 3. Case-Sensitivity: Implemented a global case-sensitivity flag for the entire search operation.
//    - Implementation: The `Search` method now accepts a `caseSensitive` boolean flag.
//    - Default: The default search is now case-insensitive, using a Unicode-aware
//      comparison (`github.com/boyter/go-string`).
//    - Tests: All tests were updated to use the new flag, and specific tests for both
//      case-sensitive and case-insensitive matching have been added.
//
// 4. Edge Cases for Multi-Value Filters: Added comprehensive tests for edge cases
//    that could break production functionality.
// Next Test Cases to Add:
// 1. NOT Operator Precedence: Test the behavior of queries like `lazy AND NOT dog` to ensure
//    that the `NOT` operator has the expected precedence and correctly filters results.
//
// 2. NOT with Filters: Test the `NOT` operator's interaction with filters, such as `NOT lang=go`,
//    to verify it correctly excludes documents based on metadata.
//
// 3. Filter Case-Sensitivity: Add tests to ensure all filters (`lang`, `ext`, `file`) behave
//    consistently with respect to case-insensitivity. For example, `ext=PY` should match `.py`.
//
// 4. Empty/Whitespace Query: Add a test for an empty or whitespace-only query (`""`) to ensure
//    it returns zero results and no error, which is the expected behavior for a "search for nothing".
//
// 5. Multi-Value Filter Edge Cases:
//    - Mixed data types in filter values
//    - Empty values in lists
//    - Special characters and escaping
//    - Performance with large value lists
//    - Operator variations with multi-values
//    - Integration with other complex filters
//    - Type safety and error handling
//    - Unicode and international character handling

package search

import (
	"errors"
	"strings"
	"testing"
)

var testDocs = []*Document{
	{Path: "src/main/file1.go", Filename: "file1.go", Language: "Go", Extension: "go", Content: []byte("A brown cat is in the house."), Complexity: 2},
	{Path: "src/main/file2.go", Filename: "file2.go", Language: "Go", Extension: "go", Content: []byte("A quick brown dog jumps over the lazy fox."), Complexity: 5},
	{Path: "pkg/search/file3.py", Filename: "file3.py", Language: "Python", Extension: "py", Content: []byte("The lazy cat sat on the mat."), Complexity: 3},
	{Path: "pkg/search/file4.py", Filename: "file4.py", Language: "Python", Extension: "py", Content: []byte("A bat and a cat are friends."), Complexity: 8},
	{Path: "vendor/lib/file5.rs", Filename: "file5.rs", Language: "Rust", Extension: "rs", Content: []byte("This is a complex document about Go programming."), Complexity: 9},
	{Path: "src/main/file6.txt", Filename: "file6.txt", Language: "Text", Extension: "txt", Content: []byte("ByteOrderMarks = [][]byte{ {254, 255}, // UTF-16 BE"), Complexity: 1},
}

func TestExecutor(t *testing.T) {
	se := NewSearchEngine(testDocs)
	testCases := []struct {
		name          string
		query         string
		caseSensitive bool
		wantPaths     []string
	}{
		{"Simple Keyword", "cat", false, []string{"src/main/file1.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Simple AND", "brown AND cat", false, []string{"src/main/file1.go"}},
		{"Implicit AND", "lazy cat", false, []string{"pkg/search/file3.py"}},
		{"OR", "dog OR fox", false, []string{"src/main/file2.go"}},
		{"NOT", "brown NOT dog", false, []string{"src/main/file1.go"}},
		{"Grouping", "(lazy OR house) AND cat", false, []string{"src/main/file1.go", "pkg/search/file3.py"}},
		{"Regex", "/[cb]at/", false, []string{"src/main/file1.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Filter =", "complexity=5", false, []string{"src/main/file2.go"}},
		{"Filter >=", "complexity>=8", false, []string{"pkg/search/file4.py", "vendor/lib/file5.rs"}},
		{"Filter !=", "complexity!=3", false, []string{"src/main/file1.go", "src/main/file2.go", "pkg/search/file4.py", "vendor/lib/file5.rs", "src/main/file6.txt"}},
		{"Combined Filter and Keyword", "lazy AND complexity<=3", false, []string{"pkg/search/file3.py"}},
		{"Language Filter", "lang=go", false, []string{"src/main/file1.go", "src/main/file2.go"}},
		{"Extension Filter", "ext=py", false, []string{"pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Multi-Value Language Filter", "lang=go,py", false, []string{"src/main/file1.go", "src/main/file2.go"}}, // must use Python/python
		{"Multi-Value Language Filter", "lang=go,python", false, []string{"src/main/file1.go", "src/main/file2.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Multi-Value Language Filter Ignore Case", "lang=go,Python", false, []string{"src/main/file1.go", "src/main/file2.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Case Insensitive Keyword", "Cat", false, []string{"src/main/file1.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Case Sensitive Keyword Match", "cat", true, []string{"src/main/file1.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Case Sensitive Keyword No Match", "Cat", true, []string{}},
		// Edge cases for multi-value filters
		{"Multi-Value with Empty Values", "lang=go,,python", false, []string{"src/main/file1.go", "src/main/file2.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Multi-Value with Leading Empty", "lang=,go,python", false, []string{"src/main/file1.go", "src/main/file2.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Multi-Value with Trailing Empty", "lang=go,python,", false, []string{"src/main/file1.go", "src/main/file2.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Multi-Value Extension Filter", "ext=go,py", false, []string{"src/main/file1.go", "src/main/file2.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Multi-Value with Special Characters", "lang=go,py+thon,java", false, []string{"src/main/file1.go", "src/main/file2.go"}},
		{"Multi-Value Filter With NOT", "lang=go NOT python", false, []string{"src/main/file1.go", "src/main/file2.go"}},
		{"Multi-Value Filter with Complex Conditions", "(lang=go OR lang=python) AND complexity>=5", false, []string{"src/main/file2.go", "pkg/search/file4.py"}},
		{"Multi-Value Filter Case Sensitivity", "lang=GO,PYTHON", false, []string{"src/main/file1.go", "src/main/file2.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Search for Numbers", "254 255", false, []string{"src/main/file6.txt"}},
		{"Fuzzy Distance 1", "houss~1", false, []string{"src/main/file1.go"}}, // "houss" is distance 1 from "house" (only in file1)
		{"Fuzzy Distance 1 No Match", "zzz~1", false, []string{}},
		{"Fuzzy AND Keyword", "houss~1 AND brown", false, []string{"src/main/file1.go"}},
		{"Fuzzy Distance 2", "hovze~2", false, []string{"src/main/file1.go"}}, // "hovze" is distance 2 from "house" (u→v, s→z), only in file1

		// Colon filter syntax
		{"Colon file filter", "cat file:file1", false, []string{"src/main/file1.go"}},
		{"Colon ext filter", "cat ext:go", false, []string{"src/main/file1.go"}},
		{"Colon lang filter", "cat lang:go", false, []string{"src/main/file1.go"}},

		// Path filter
		{"Path filter colon src/main", "path:src/main", false, []string{"src/main/file1.go", "src/main/file2.go", "src/main/file6.txt"}},
		{"Path filter colon pkg/search", "path:pkg/search", false, []string{"pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Path filter vendor", "path=vendor", false, []string{"vendor/lib/file5.rs"}},
		{"Path filter with keyword", "cat path:pkg/search", false, []string{"pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Path filter != excludes", "path!=vendor", false, []string{"src/main/file1.go", "src/main/file2.go", "pkg/search/file3.py", "pkg/search/file4.py", "src/main/file6.txt"}},
		{"Path filter case insensitive", "path:SRC/MAIN", false, []string{"src/main/file1.go", "src/main/file2.go", "src/main/file6.txt"}},
		{"Path filter quoted value", `path="src/main"`, false, []string{"src/main/file1.go", "src/main/file2.go", "src/main/file6.txt"}},
		{"Colon path filter with keyword", "cat path:pkg/search", false, []string{"pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Filepath alias", "filepath=vendor", false, []string{"vendor/lib/file5.rs"}},

		// TODO: Add tests for NOT operator precedence and filter interaction.
		// {"NOT Operator Precedence", "lazy AND NOT dog", false, []string{"file3.py"}},
		// {"NOT with Filter", "NOT lang=go", false, []string{"file3.py", "file4.py", "file5.rs"}},
		// {"Empty Query", "", false, []string{}},
		// {"Whitespace Query", " ", false, []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := se.Search(tc.query, tc.caseSensitive)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if len(res.Documents) != len(tc.wantPaths) {
				t.Fatalf("got %d results, want %d", len(res.Documents), len(tc.wantPaths))
			}

			gotPaths := make(map[string]bool)
			for _, doc := range res.Documents {
				gotPaths[doc.Path] = true
			}
			for _, wantPath := range tc.wantPaths {
				if !gotPaths[wantPath] {
					t.Errorf("missing expected document path: %s", wantPath)
				}
			}
		})
	}
}

func TestParserHealing(t *testing.T) {
	se := NewSearchEngine(testDocs)
	testCases := []struct {
		name       string
		query      string
		wantNotice string
		wantPaths  []string
	}{
		{"Dangling AND", "cat AND", "Trailing 'AND' was ignored", []string{"src/main/file1.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Dangling OR", "dog OR", "Trailing 'OR' was ignored", []string{"src/main/file2.go"}},
		{"Missing Closing Paren", "(cat OR dog", "Missing ')' was added", []string{"src/main/file1.go", "src/main/file2.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
		{"Extra Closing Paren", "cat)", "Unmatched ')' was ignored", []string{"src/main/file1.go", "pkg/search/file3.py", "pkg/search/file4.py"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := se.Search(tc.query, false)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			foundNotice := false
			for _, notice := range res.Notices {
				if strings.Contains(notice, tc.wantNotice) {
					foundNotice = true
					break
				}
			}
			if !foundNotice {
				t.Errorf("did not find expected notice '%s' in notices: %v", tc.wantNotice, res.Notices)
			}

			if len(res.Documents) != len(tc.wantPaths) {
				t.Errorf("got %d results, want %d", len(res.Documents), len(tc.wantPaths))
			}
		})
	}
}

func TestTransformer(t *testing.T) {
	se := NewSearchEngine(testDocs)
	res, err := se.Search("complexity=high", false)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(res.Documents) != 2 { // Should match docs with path pkg/search/file4.py and vendor/lib/file5.rs
		t.Errorf("got %d results, want 2", len(res.Documents))
	}

	wantNotice := "'complexity=high' was interpreted as 'complexity >= 8'"
	foundNotice := false
	for _, notice := range res.Notices {
		if strings.Contains(notice, wantNotice) {
			foundNotice = true
			break
		}
	}
	if !foundNotice {
		t.Errorf("did not find expected transformation notice in: %v", res.Notices)
	}
}

func TestSearchErrors(t *testing.T) {
	se := NewSearchEngine(testDocs)
	testCases := []struct {
		name    string
		query   string
		wantErr error
	}{
		{"Query just AND", "AND", ErrInvalidQuery},
		{"Query just operator", ">", ErrInvalidQuery},
		{"Query just OR", "OR", ErrInvalidQuery},
		{"Query just NOT", "NOT", ErrInvalidQuery},
		{"Filter without value", "lang=", ErrInvalidQuery},
		{"Filter with multiple operators", "lang>>5", ErrInvalidQuery},
		{"Query just comma", ",", ErrInvalidQuery},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := se.Search(tc.query, false)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Search() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestLexer(t *testing.T) {
	testCases := []struct {
		name  string
		query string
		want  []Token
	}{
		{
			name:  "Hashtag define",
			query: "#define MAX_CUBES 30",
			want: []Token{
				{Type: IDENTIFIER, Literal: "#define"},
				{Type: IDENTIFIER, Literal: "MAX_CUBES"},
				{Type: NUMBER, Literal: "30"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "C++ Comment",
			query: `"//"`,
			want: []Token{
				{Type: PHRASE, Literal: "//"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "Pointer ->",
			query: `"my_var->field"`,
			want: []Token{
				{Type: PHRASE, Literal: "my_var->field"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "Complex filter",
			query: "lang=go,c++",
			want: []Token{
				{Type: IDENTIFIER, Literal: "lang"},
				{Type: OPERATOR, Literal: "="},
				{Type: IDENTIFIER, Literal: "go"},
				{Type: COMMA, Literal: ","},
				{Type: IDENTIFIER, Literal: "c++"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "Fuzzy distance 1",
			query: "cat~1",
			want: []Token{
				{Type: FUZZY, Literal: "cat~1"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "Fuzzy distance 2",
			query: "wickham~2",
			want: []Token{
				{Type: FUZZY, Literal: "wickham~2"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "Fuzzy with other tokens",
			query: "cat~1 AND dog",
			want: []Token{
				{Type: FUZZY, Literal: "cat~1"},
				{Type: AND, Literal: "AND"},
				{Type: IDENTIFIER, Literal: "dog"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "Tilde without valid distance is identifier",
			query: "cat~3",
			want: []Token{
				{Type: IDENTIFIER, Literal: "cat~3"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "Colon in identifier (file:test)",
			query: "file:test",
			want: []Token{
				{Type: IDENTIFIER, Literal: "file:test"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "Colon path filter (path:pkg/search)",
			query: "path:pkg/search",
			want: []Token{
				{Type: IDENTIFIER, Literal: "path:pkg/search"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "Colon non-filter (std::cout)",
			query: "std::cout",
			want: []Token{
				{Type: IDENTIFIER, Literal: "std::cout"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "Parens and symbols",
			query: "func(a *int)",
			want: []Token{
				{Type: IDENTIFIER, Literal: "func"},
				{Type: LPAREN, Literal: "("},
				{Type: IDENTIFIER, Literal: "a"},
				{Type: IDENTIFIER, Literal: "*int"},
				{Type: RPAREN, Literal: ")"},
				{Type: EOF, Literal: ""},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tc.query))
			var got []Token
			for {
				tok := lexer.scan()
				if tok.Type == WS {
					continue
				}
				got = append(got, tok)
				if tok.Type == EOF {
					break
				}
			}

			if len(got) != len(tc.want) {
				t.Fatalf("got %d tokens, want %d.\ngot:  %v\nwant: %v", len(got), len(tc.want), got, tc.want)
			}

			for i := range got {
				if got[i].Type != tc.want[i].Type || got[i].Literal != tc.want[i].Literal {
					t.Errorf("token %d mismatch.\ngot:  %v\nwant: %v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestTermExtractor(t *testing.T) {
	testCases := []struct {
		name  string
		query string
		want  []string
	}{
		{"Simple Keyword", "cat", []string{"cat"}},
		{"Simple AND", "brown AND cat", []string{"brown", "cat"}},
		{"Implicit AND", "lazy cat", []string{"lazy", "cat"}},
		{"OR", "dog OR fox", []string{"dog", "fox"}},
		{"Simple NOT", "brown NOT dog", []string{"brown"}},
		{"Leading NOT", "NOT brown", []string{}},
		{"Grouping", "(lazy OR house) AND cat", []string{"lazy", "house", "cat"}},
		{"Regex", "/[cb]at/", []string{"[cb]at"}},
		{"Phrase", `"lazy cat"`, []string{"lazy cat"}},
		{"Mixed Types", `cat AND "lazy dog" AND /[a-z]/`, []string{"cat", "lazy dog", "[a-z]"}},
		{"Nested NOT", `cat AND NOT (dog OR fox)`, []string{"cat"}},
		{"Complex Nested", `(cat AND NOT dog) OR (fox AND NOT bird)`, []string{"cat", "fox"}},
		{"Fuzzy Term", "cat~1", []string{"cat"}},
		{"Fuzzy with AND", "cat~1 AND dog", []string{"cat", "dog"}},
		{"Filter Only", "lang=go", []string{}},
		{"Filter with Keyword", "cat lang=go", []string{"cat"}},
		{"Colon Filter Only", "file:test", []string{}},
		{"Colon Filter with Keyword", "cat file:test", []string{"cat"}},
		{"Colon Path Filter Only", "path:pkg/search", []string{}},
		{"Colon Path Filter with Keyword", "cat path:pkg/search", []string{"cat"}},
		{"Empty Query", "", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tc.query))
			parser := NewParser(lexer)
			ast, _ := parser.ParseQuery()

			got := ExtractTerms(ast)

			if len(got) != len(tc.want) {
				t.Fatalf("got %d terms, want %d. got: %v, want: %v", len(got), len(tc.want), got, tc.want)
			}

			gotMap := make(map[string]bool)
			for _, term := range got {
				gotMap[term] = true
			}
			for _, term := range tc.want {
				if !gotMap[term] {
					t.Errorf("missing expected term: %s", term)
				}
			}
		})
	}
}

func TestEvaluateFile(t *testing.T) {
	parse := func(query string) Node {
		lexer := NewLexer(strings.NewReader(query))
		parser := NewParser(lexer)
		ast, _ := parser.ParseQuery()
		return ast
	}

	testCases := []struct {
		name          string
		query         string
		content       string
		caseSensitive bool
		wantMatch     bool
		wantTerms     []string // terms expected in locations map
	}{
		{
			name:      "Simple match",
			query:     "cat",
			content:   "the cat sat on the mat",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "No match",
			query:     "dog",
			content:   "the cat sat on the mat",
			wantMatch: false,
		},
		{
			name:      "AND both present",
			query:     "cat AND mat",
			content:   "the cat sat on the mat",
			wantMatch: true,
			wantTerms: []string{"cat", "mat"},
		},
		{
			name:      "AND one missing",
			query:     "cat AND dog",
			content:   "the cat sat on the mat",
			wantMatch: false,
		},
		{
			name:      "OR first present",
			query:     "cat OR dog",
			content:   "the cat sat on the mat",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "OR second present",
			query:     "dog OR mat",
			content:   "the cat sat on the mat",
			wantMatch: true,
			wantTerms: []string{"mat"},
		},
		{
			name:      "OR neither present",
			query:     "dog OR bird",
			content:   "the cat sat on the mat",
			wantMatch: false,
		},
		{
			name:      "NOT excludes",
			query:     "NOT cat",
			content:   "the cat sat on the mat",
			wantMatch: false,
		},
		{
			name:      "NOT does not exclude",
			query:     "NOT dog",
			content:   "the cat sat on the mat",
			wantMatch: true,
		},
		{
			name:      "Regex match",
			query:     "/[cm]at/",
			content:   "the cat sat on the mat",
			wantMatch: true,
			wantTerms: []string{"[cm]at"},
		},
		{
			name:      "Regex no match",
			query:     "/[dz]og/",
			content:   "the cat sat on the mat",
			wantMatch: false,
		},
		{
			name:          "Case sensitive match",
			query:         "cat",
			content:       "the cat sat on the mat",
			caseSensitive: true,
			wantMatch:     true,
			wantTerms:     []string{"cat"},
		},
		{
			name:          "Case sensitive no match",
			query:         "Cat",
			content:       "the cat sat on the mat",
			caseSensitive: true,
			wantMatch:     false,
		},
		{
			name:      "Case insensitive match",
			query:     "Cat",
			content:   "the cat sat on the mat",
			wantMatch: true,
			wantTerms: []string{"Cat"},
		},
		{
			name:      "Lang filter passes through in per-file mode",
			query:     "cat AND lang=go",
			content:   "the cat sat on the mat",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Complex AND OR NOT",
			query:     "(cat OR dog) AND NOT bird",
			content:   "the cat sat on the mat",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Phrase match",
			query:     `"cat sat"`,
			content:   "the cat sat on the mat",
			wantMatch: true,
			wantTerms: []string{"cat sat"},
		},
		{
			name:      "Phrase no match",
			query:     `"cat mat"`,
			content:   "the cat sat on the mat",
			wantMatch: false,
		},
	}

	// Add file filter tests with specific filenames
	fileFilterCases := []struct {
		name          string
		query         string
		content       string
		filename      string
		location      string
		caseSensitive bool
		wantMatch     bool
		wantTerms     []string
	}{
		{
			name:      "File filter matches substring",
			query:     "cat file=test",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "File filter no match",
			query:     "cat file=test",
			content:   "the cat sat on the mat",
			filename:  "main.go",
			location:  "src/main.go",
			wantMatch: false,
		},
		{
			name:      "File filter != operator",
			query:     "cat file!=test",
			content:   "the cat sat on the mat",
			filename:  "main.go",
			location:  "src/main.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "File filter != excludes matching file",
			query:     "cat file!=test",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: false,
		},
		{
			name:      "Extension filter in per-file mode",
			query:     "cat ext=go",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Extension filter no match in per-file mode",
			query:     "cat ext=py",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: false,
		},
		{
			name:      "Filename filter case insensitive",
			query:     "cat file=TEST",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Extension filter case insensitive",
			query:     "cat ext=GO",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Lang filter passes through in per-file mode",
			query:     "cat lang=go",
			content:   "the cat sat on the mat",
			filename:  "anything.py",
			location:  "src/anything.py",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		// Colon syntax tests
		{
			name:      "Colon file filter matches",
			query:     "cat file:test",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Colon file filter no match",
			query:     "cat file:test",
			content:   "the cat sat on the mat",
			filename:  "main.go",
			location:  "src/main.go",
			wantMatch: false,
		},
		{
			name:      "Colon ext filter matches",
			query:     "cat ext:go",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		// Path filter tests in per-file mode
		{
			name:      "Path filter matches directory",
			query:     "cat path:pkg/search",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Path filter no match",
			query:     "cat path=vendor",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: false,
		},
		{
			name:      "Path filter != excludes",
			query:     "cat path!=vendor",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Path filter case insensitive",
			query:     "cat path:PKG/SEARCH",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Colon path filter matches",
			query:     "cat path:pkg/search",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Colon path filter no match",
			query:     "cat path:vendor",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: false,
		},
		{
			name:      "Filepath alias in per-file mode",
			query:     "cat filepath:pkg/search",
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
		{
			name:      "Path filter with quoted value",
			query:     `cat path="pkg/search"`,
			content:   "the cat sat on the mat",
			filename:  "search_test.go",
			location:  "pkg/search/search_test.go",
			wantMatch: true,
			wantTerms: []string{"cat"},
		},
	}

	for _, tc := range fileFilterCases {
		t.Run(tc.name, func(t *testing.T) {
			node := parse(tc.query)
			matched, locations := EvaluateFile(node, []byte(tc.content), tc.filename, tc.location, tc.caseSensitive)

			if matched != tc.wantMatch {
				t.Errorf("EvaluateFile() matched = %v, want %v", matched, tc.wantMatch)
			}

			if tc.wantMatch && tc.wantTerms != nil {
				for _, term := range tc.wantTerms {
					if _, ok := locations[term]; !ok {
						t.Errorf("EvaluateFile() missing locations for term %q, got keys: %v", term, mapKeys(locations))
					}
				}
			}
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := parse(tc.query)
			matched, locations := EvaluateFile(node, []byte(tc.content), "test.txt", "test.txt", tc.caseSensitive)

			if matched != tc.wantMatch {
				t.Errorf("EvaluateFile() matched = %v, want %v", matched, tc.wantMatch)
			}

			if tc.wantMatch && tc.wantTerms != nil {
				for _, term := range tc.wantTerms {
					if _, ok := locations[term]; !ok {
						t.Errorf("EvaluateFile() missing locations for term %q, got keys: %v", term, mapKeys(locations))
					}
				}
			}
		})
	}
}

func mapKeys(m map[string][][]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// --- PlanAST tests ---

func TestPlanASTReordersByCost(t *testing.T) {
	// filter (cost 10) should be evaluated before keyword (cost 20)
	// Build: keyword AND filter — planner should reorder to filter AND keyword
	ast := &AndNode{
		Left:  &KeywordNode{Value: "cat"},
		Right: &FilterNode{Field: "lang", Operator: "=", Value: "go"},
	}
	planned := PlanAST(ast)

	and, ok := planned.(*AndNode)
	if !ok {
		t.Fatalf("expected *AndNode, got %T", planned)
	}
	if _, ok := and.Left.(*FilterNode); !ok {
		t.Errorf("expected FilterNode on left after planning, got %T", and.Left)
	}
	if _, ok := and.Right.(*KeywordNode); !ok {
		t.Errorf("expected KeywordNode on right after planning, got %T", and.Right)
	}
}

func TestPlanASTFlattensNestedAnd(t *testing.T) {
	// ((filter AND keyword) AND regex) should flatten and sort: filter, keyword, regex
	ast := &AndNode{
		Left: &AndNode{
			Left:  &RegexNode{Pattern: "[a-z]"},
			Right: &KeywordNode{Value: "cat"},
		},
		Right: &FilterNode{Field: "ext", Operator: "=", Value: "go"},
	}
	planned := PlanAST(ast)

	// Walk the AND tree left-to-right and collect leaf nodes
	var leaves []Node
	var collect func(n Node)
	collect = func(n Node) {
		if a, ok := n.(*AndNode); ok {
			collect(a.Left)
			collect(a.Right)
		} else {
			leaves = append(leaves, n)
		}
	}
	collect(planned)

	if len(leaves) != 3 {
		t.Fatalf("expected 3 leaves, got %d", len(leaves))
	}
	if _, ok := leaves[0].(*FilterNode); !ok {
		t.Errorf("first leaf should be FilterNode (cost 10), got %T", leaves[0])
	}
	if _, ok := leaves[1].(*KeywordNode); !ok {
		t.Errorf("second leaf should be KeywordNode (cost 20), got %T", leaves[1])
	}
	if _, ok := leaves[2].(*RegexNode); !ok {
		t.Errorf("third leaf should be RegexNode (cost 50), got %T", leaves[2])
	}
}

func TestPlanASTNilAndSingleNode(t *testing.T) {
	if PlanAST(nil) != nil {
		t.Error("PlanAST(nil) should return nil")
	}

	kw := &KeywordNode{Value: "cat"}
	if PlanAST(kw) != kw {
		t.Error("PlanAST on single node should return same node")
	}
}

func TestPlanASTPreservesOrAndNot(t *testing.T) {
	// OR node children should be planned recursively
	ast := &OrNode{
		Left: &AndNode{
			Left:  &KeywordNode{Value: "cat"},
			Right: &FilterNode{Field: "lang", Operator: "=", Value: "go"},
		},
		Right: &NotNode{
			Expr: &KeywordNode{Value: "dog"},
		},
	}
	planned := PlanAST(ast)

	or, ok := planned.(*OrNode)
	if !ok {
		t.Fatalf("expected *OrNode, got %T", planned)
	}
	// Left AND should be reordered: filter before keyword
	and, ok := or.Left.(*AndNode)
	if !ok {
		t.Fatalf("expected *AndNode on left of OR, got %T", or.Left)
	}
	if _, ok := and.Left.(*FilterNode); !ok {
		t.Errorf("expected FilterNode first in planned AND, got %T", and.Left)
	}
	// NOT should be preserved
	if _, ok := or.Right.(*NotNode); !ok {
		t.Errorf("expected *NotNode on right of OR, got %T", or.Right)
	}
}

func TestGetCostAllNodeTypes(t *testing.T) {
	tests := []struct {
		name string
		node Node
		want int
	}{
		{"Filter", &FilterNode{}, filterCost},
		{"Keyword", &KeywordNode{}, keywordCost},
		{"Phrase", &PhraseNode{}, phraseCost},
		{"Regex", &RegexNode{}, regexCost},
		{"Fuzzy", &FuzzyNode{}, regexCost},
		{"Or", &OrNode{}, orCost},
		{"Not", &NotNode{}, notCost},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := getCost(tc.node); got != tc.want {
				t.Errorf("getCost(%T) = %d, want %d", tc.node, got, tc.want)
			}
		})
	}
}

// --- FuzzyNode tests ---

func TestFuzzyNodeExtractTerms(t *testing.T) {
	// Positive context: term should be extracted
	ast := &AndNode{
		Left:  &KeywordNode{Value: "cat"},
		Right: &FuzzyNode{Value: "dog", Distance: 1},
	}
	terms := ExtractTerms(ast)
	termSet := make(map[string]bool)
	for _, t := range terms {
		termSet[t] = true
	}
	if !termSet["cat"] {
		t.Error("expected 'cat' in extracted terms")
	}
	if !termSet["dog"] {
		t.Error("expected 'dog' from FuzzyNode in extracted terms")
	}
}

func TestFuzzyNodeExtractTermsNotContext(t *testing.T) {
	ast := &NotNode{Expr: &FuzzyNode{Value: "dog", Distance: 2}}
	terms := ExtractTerms(ast)
	if len(terms) != 0 {
		t.Errorf("expected no terms from FuzzyNode in NOT context, got %v", terms)
	}
}

func TestFuzzyNodeEvaluate(t *testing.T) {
	docs := []*Document{
		{Path: "a.go", Filename: "a.go", Content: []byte("hello world")},
		{Path: "b.go", Filename: "b.go", Content: []byte("goodbye mars")},
	}
	se := NewSearchEngine(docs)

	t.Run("Exact match within distance", func(t *testing.T) {
		node := &FuzzyNode{Value: "hello", Distance: 1}
		results := se.evaluate(node, se.documents, false)
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
		if len(results) == 1 && results[0].Path != "a.go" {
			t.Errorf("expected a.go, got %s", results[0].Path)
		}
	})

	t.Run("Fuzzy match distance 1", func(t *testing.T) {
		// "hallo" is distance 1 from "hello"
		node := &FuzzyNode{Value: "hallo", Distance: 1}
		results := se.evaluate(node, se.documents, false)
		if len(results) != 1 {
			t.Errorf("expected 1 result for hallo~1, got %d", len(results))
		}
	})

	t.Run("Fuzzy match distance 2", func(t *testing.T) {
		// "hxllo" is distance 1 from "hello", "hxlly" is distance 2
		node := &FuzzyNode{Value: "hxlly", Distance: 2}
		results := se.evaluate(node, se.documents, false)
		if len(results) != 1 {
			t.Errorf("expected 1 result for hxlly~2, got %d", len(results))
		}
	})

	t.Run("No match beyond distance", func(t *testing.T) {
		// "zzzzz" is far from anything in docs
		node := &FuzzyNode{Value: "zzzzz", Distance: 1}
		results := se.evaluate(node, se.documents, false)
		if len(results) != 0 {
			t.Errorf("expected 0 results for zzzzz~1, got %d", len(results))
		}
	})
}

func TestFuzzyNodeEvaluateFile(t *testing.T) {
	t.Run("Exact match", func(t *testing.T) {
		node := &FuzzyNode{Value: "hello", Distance: 1}
		matched, locs := EvaluateFile(node, []byte("hello world"), "test.txt", "test.txt", false)
		if !matched {
			t.Error("FuzzyNode in EvaluateFile should match 'hello' in 'hello world'")
		}
		if len(locs) == 0 {
			t.Error("expected locations for fuzzy match")
		}
	})

	t.Run("Fuzzy match", func(t *testing.T) {
		node := &FuzzyNode{Value: "hallo", Distance: 1}
		matched, locs := EvaluateFile(node, []byte("hello world"), "test.txt", "test.txt", false)
		if !matched {
			t.Error("FuzzyNode should match 'hello' for query 'hallo~1'")
		}
		if _, ok := locs["hallo"]; !ok {
			t.Errorf("expected locations under key 'hallo', got keys: %v", mapKeys(locs))
		}
	})

	t.Run("No match", func(t *testing.T) {
		node := &FuzzyNode{Value: "zzzzz", Distance: 1}
		matched, _ := EvaluateFile(node, []byte("hello world"), "test.txt", "test.txt", false)
		if matched {
			t.Error("FuzzyNode should not match 'zzzzz~1' in 'hello world'")
		}
	})
}

// --- Invalid regex tests ---

func TestInvalidRegexSearch(t *testing.T) {
	se := NewSearchEngine(testDocs)
	res, err := se.Search("/[invalid/", false)
	if err != nil {
		t.Fatalf("Search with invalid regex should not error, got: %v", err)
	}
	if len(res.Documents) != 0 {
		t.Errorf("invalid regex should match 0 docs, got %d", len(res.Documents))
	}
}

func TestInvalidRegexEvaluateFile(t *testing.T) {
	node := &RegexNode{Pattern: "[invalid"}
	matched, _ := EvaluateFile(node, []byte("hello world"), "test.txt", "test.txt", false)
	if matched {
		t.Error("EvaluateFile with invalid regex should return false")
	}
}

// --- Unknown filter field ---

func TestUnknownFilterField(t *testing.T) {
	se := NewSearchEngine(testDocs)
	res, err := se.Search("author=bob", false)
	if err != nil {
		t.Fatalf("Search with unknown filter should not error, got: %v", err)
	}
	if len(res.Documents) != 0 {
		t.Errorf("unknown filter should match 0 docs, got %d", len(res.Documents))
	}
}

// --- Filter handler edge cases ---

func TestFilterHandlerEdgeCases(t *testing.T) {
	t.Run("Complexity with non-int value", func(t *testing.T) {
		doc := &Document{Path: "a.go", Complexity: 5}
		if handleComplexityFilter("=", "notanumber", doc) {
			t.Error("complexity filter with string value should return false")
		}
	})

	t.Run("Complexity with unsupported operator", func(t *testing.T) {
		doc := &Document{Path: "a.go", Complexity: 5}
		if handleComplexityFilter("~", 5, doc) {
			t.Error("complexity filter with unsupported operator should return false")
		}
	})

	t.Run("Filename with non-string value", func(t *testing.T) {
		doc := &Document{Path: "a.go", Filename: "a.go"}
		if handleFilenameFilter("=", 123, doc) {
			t.Error("filename filter with int value should return false")
		}
	})

	t.Run("Filename with unsupported operator", func(t *testing.T) {
		doc := &Document{Path: "a.go", Filename: "a.go"}
		if handleFilenameFilter(">=", "a.go", doc) {
			t.Error("filename filter with >= should return false")
		}
	})

	t.Run("Language with non-string non-slice value", func(t *testing.T) {
		doc := &Document{Path: "a.go", Language: "Go"}
		if handleLanguageFilter("=", 123, doc) {
			t.Error("language filter with int value should return false")
		}
	})

	t.Run("Extension with non-string non-slice value", func(t *testing.T) {
		doc := &Document{Path: "a.go", Extension: "go"}
		if handleExtensionFilter("=", 123, doc) {
			t.Error("extension filter with int value should return false")
		}
	})

	t.Run("Language != single value", func(t *testing.T) {
		doc := &Document{Path: "a.go", Language: "Go"}
		if !handleLanguageFilter("!=", "python", doc) {
			t.Error("Go != python should be true")
		}
		if handleLanguageFilter("!=", "go", doc) {
			t.Error("Go != go should be false")
		}
	})

	t.Run("Extension != single value", func(t *testing.T) {
		doc := &Document{Path: "a.go", Extension: "go"}
		if !handleExtensionFilter("!=", "py", doc) {
			t.Error("go != py should be true")
		}
		if handleExtensionFilter("!=", "go", doc) {
			t.Error("go != go should be false")
		}
	})

	t.Run("Language != multi-value", func(t *testing.T) {
		doc := &Document{Path: "a.go", Language: "Go"}
		vals := []interface{}{"python", "rust"}
		if !handleLanguageFilter("!=", vals, doc) {
			t.Error("Go not in [python, rust] with != should be true")
		}
		valsWithGo := []interface{}{"go", "rust"}
		if handleLanguageFilter("!=", valsWithGo, doc) {
			t.Error("Go in [go, rust] with != should be false")
		}
	})

	t.Run("Extension != multi-value", func(t *testing.T) {
		doc := &Document{Path: "a.go", Extension: "go"}
		vals := []interface{}{"py", "rs"}
		if !handleExtensionFilter("!=", vals, doc) {
			t.Error("go not in [py, rs] with != should be true")
		}
		valsWithGo := []interface{}{"go", "rs"}
		if handleExtensionFilter("!=", valsWithGo, doc) {
			t.Error("go in [go, rs] with != should be false")
		}
	})

	t.Run("Path with non-string value", func(t *testing.T) {
		doc := &Document{Path: "pkg/search/a.go"}
		if handlePathFilter("=", 123, doc) {
			t.Error("path filter with int value should return false")
		}
	})

	t.Run("Path with unsupported operator", func(t *testing.T) {
		doc := &Document{Path: "pkg/search/a.go"}
		if handlePathFilter(">=", "pkg", doc) {
			t.Error("path filter with >= should return false")
		}
	})
}

// --- Filter alias tests ---

func TestFilterAliases(t *testing.T) {
	se := NewSearchEngine(testDocs)

	tests := []struct {
		name      string
		query     string
		wantPaths []string
	}{
		{"language=go", "language=go", []string{"src/main/file1.go", "src/main/file2.go"}},
		{"lang=go", "lang=go", []string{"src/main/file1.go", "src/main/file2.go"}},
		{"extension=py", "extension=py", []string{"pkg/search/file3.py", "pkg/search/file4.py"}},
		{"ext=py", "ext=py", []string{"pkg/search/file3.py", "pkg/search/file4.py"}},
		{"filename=file1.go", "filename=file1.go", []string{"src/main/file1.go"}},
		{"file=file1.go", "file=file1.go", []string{"src/main/file1.go"}},
		{"file substring match", "file=file1", []string{"src/main/file1.go"}},
		{"file substring match .go", "file=.go", []string{"src/main/file1.go", "src/main/file2.go"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := se.Search(tc.query, false)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			if len(res.Documents) != len(tc.wantPaths) {
				t.Fatalf("got %d results, want %d", len(res.Documents), len(tc.wantPaths))
			}
			gotPaths := make(map[string]bool)
			for _, doc := range res.Documents {
				gotPaths[doc.Path] = true
			}
			for _, p := range tc.wantPaths {
				if !gotPaths[p] {
					t.Errorf("missing expected path: %s", p)
				}
			}
		})
	}
}

// --- Transformer expanded tests ---

func TestTransformerExpanded(t *testing.T) {
	se := NewSearchEngine(testDocs)

	t.Run("complexity=HIGH case insensitive", func(t *testing.T) {
		res, err := se.Search("complexity=HIGH", false)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(res.Documents) != 2 {
			t.Errorf("complexity=HIGH should match 2 docs, got %d", len(res.Documents))
		}
	})

	t.Run("complexity=High mixed case", func(t *testing.T) {
		res, err := se.Search("complexity=High", false)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(res.Documents) != 2 {
			t.Errorf("complexity=High should match 2 docs, got %d", len(res.Documents))
		}
	})

	t.Run("Non-complexity filter not transformed", func(t *testing.T) {
		// lang=high should NOT be transformed — should match nothing (no language called "high")
		res, err := se.Search("lang=high", false)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(res.Documents) != 0 {
			t.Errorf("lang=high should match 0 docs, got %d", len(res.Documents))
		}
	})

	t.Run("Numeric complexity not transformed", func(t *testing.T) {
		// complexity=5 has int value, not string — transformer should leave it alone
		res, err := se.Search("complexity=5", false)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(res.Documents) != 1 {
			t.Errorf("complexity=5 should match 1 doc, got %d", len(res.Documents))
		}
	})

	t.Run("Transformer with nil AST", func(t *testing.T) {
		tr := &Transformer{}
		result, notices := tr.TransformAST(nil)
		if result != nil {
			t.Error("TransformAST(nil) should return nil")
		}
		if len(notices) != 0 {
			t.Errorf("TransformAST(nil) should produce no notices, got %v", notices)
		}
	})

	t.Run("Nested transform in AND", func(t *testing.T) {
		res, err := se.Search("cat AND complexity=high", false)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		// file4.py has cat + complexity 8, file5.rs has complexity 9 but no cat
		if len(res.Documents) != 1 {
			t.Errorf("expected 1 doc, got %d", len(res.Documents))
		}
		if len(res.Documents) == 1 && res.Documents[0].Path != "pkg/search/file4.py" {
			t.Errorf("expected file4.py, got %s", res.Documents[0].Path)
		}
	})

	t.Run("Nested transform in OR", func(t *testing.T) {
		res, err := se.Search("complexity=high OR lang=rust", false)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		// complexity>=8: file4.py, file5.rs. lang=rust: file5.rs. Union: file4.py, file5.rs
		if len(res.Documents) != 2 {
			t.Errorf("expected 2 docs, got %d", len(res.Documents))
		}
	})
}

// --- EvaluateFile edge cases ---

func TestEvaluateFileNilNode(t *testing.T) {
	matched, locs := EvaluateFile(nil, []byte("hello"), "test.txt", "test.txt", false)
	if !matched {
		t.Error("EvaluateFile(nil) should return true")
	}
	if locs != nil {
		t.Errorf("EvaluateFile(nil) should return nil locations, got %v", locs)
	}
}

// --- Empty document list ---

func TestEmptyDocumentList(t *testing.T) {
	se := NewSearchEngine([]*Document{})

	t.Run("Keyword on empty", func(t *testing.T) {
		res, err := se.Search("cat", false)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(res.Documents) != 0 {
			t.Errorf("expected 0 results, got %d", len(res.Documents))
		}
	})

	t.Run("Filter on empty", func(t *testing.T) {
		res, err := se.Search("lang=go", false)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(res.Documents) != 0 {
			t.Errorf("expected 0 results, got %d", len(res.Documents))
		}
	})
}
