package search

import "testing"

func TestCountAllTerms_SingleKeyword(t *testing.T) {
	ast := &KeywordNode{Value: "hello"}
	if got := CountAllTerms(ast); got != 1 {
		t.Errorf("CountAllTerms = %d, want 1", got)
	}
}

func TestCountAllTerms_AndNode(t *testing.T) {
	ast := &AndNode{
		Left:  &KeywordNode{Value: "hello"},
		Right: &KeywordNode{Value: "world"},
	}
	if got := CountAllTerms(ast); got != 2 {
		t.Errorf("CountAllTerms = %d, want 2", got)
	}
}

func TestCountAllTerms_DuplicateTerms(t *testing.T) {
	ast := &AndNode{
		Left:  &KeywordNode{Value: "hello"},
		Right: &KeywordNode{Value: "hello"},
	}
	if got := CountAllTerms(ast); got != 1 {
		t.Errorf("CountAllTerms = %d, want 1 (duplicates collapsed)", got)
	}
}

func TestCountAllTerms_IncludesNotTerms(t *testing.T) {
	ast := &AndNode{
		Left:  &KeywordNode{Value: "hello"},
		Right: &NotNode{Expr: &KeywordNode{Value: "world"}},
	}
	if got := CountAllTerms(ast); got != 2 {
		t.Errorf("CountAllTerms = %d, want 2 (NOT terms counted)", got)
	}
}

func TestCountAllTerms_IncludesFilterValues(t *testing.T) {
	ast := &AndNode{
		Left:  &KeywordNode{Value: "hello"},
		Right: &FilterNode{Field: "lang", Operator: "=", Value: "go"},
	}
	if got := CountAllTerms(ast); got != 2 {
		t.Errorf("CountAllTerms = %d, want 2 (filter values counted)", got)
	}
}

func TestCountAllTerms_AllNodeTypes(t *testing.T) {
	ast := &AndNode{
		Left: &AndNode{
			Left:  &KeywordNode{Value: "keyword"},
			Right: &PhraseNode{Value: "exact phrase"},
		},
		Right: &AndNode{
			Left: &OrNode{
				Left:  &RegexNode{Pattern: "func\\s+Test"},
				Right: &FuzzyNode{Value: "fuzzy", Distance: 1},
			},
			Right: &FilterNode{Field: "ext", Operator: "=", Value: "go"},
		},
	}
	if got := CountAllTerms(ast); got != 5 {
		t.Errorf("CountAllTerms = %d, want 5", got)
	}
}

func TestCountAllTerms_NilNode(t *testing.T) {
	if got := CountAllTerms(nil); got != 0 {
		t.Errorf("CountAllTerms(nil) = %d, want 0", got)
	}
}

func TestCountAllTerms_ComplexTree(t *testing.T) {
	// Simulates: a b c NOT d lang:go ext:ts "phrase" /regex/ fuzzy~1
	ast := &AndNode{
		Left: &AndNode{
			Left: &AndNode{
				Left:  &KeywordNode{Value: "a"},
				Right: &KeywordNode{Value: "b"},
			},
			Right: &AndNode{
				Left:  &KeywordNode{Value: "c"},
				Right: &NotNode{Expr: &KeywordNode{Value: "d"}},
			},
		},
		Right: &AndNode{
			Left: &AndNode{
				Left:  &FilterNode{Field: "lang", Operator: "=", Value: "go"},
				Right: &FilterNode{Field: "ext", Operator: "=", Value: "ts"},
			},
			Right: &AndNode{
				Left:  &PhraseNode{Value: "phrase"},
				Right: &AndNode{
					Left:  &RegexNode{Pattern: "regex"},
					Right: &FuzzyNode{Value: "fuzzy", Distance: 1},
				},
			},
		},
	}
	// a, b, c, d, go, ts, phrase, regex, fuzzy = 9 unique terms
	if got := CountAllTerms(ast); got != 9 {
		t.Errorf("CountAllTerms = %d, want 9", got)
	}
}

func TestCountAllTerms_MultiValueFilter(t *testing.T) {
	ast := &AndNode{
		Left:  &KeywordNode{Value: "search"},
		Right: &FilterNode{Field: "lang", Operator: "=", Value: []interface{}{"go", "python"}},
	}
	if got := CountAllTerms(ast); got != 3 {
		t.Errorf("CountAllTerms = %d, want 3 (keyword + 2 filter values)", got)
	}
}

func TestExtractTerms_IgnoresNotTerms(t *testing.T) {
	ast := &AndNode{
		Left:  &KeywordNode{Value: "hello"},
		Right: &NotNode{Expr: &KeywordNode{Value: "world"}},
	}
	terms := ExtractTerms(ast)
	if len(terms) != 1 || terms[0] != "hello" {
		t.Errorf("ExtractTerms = %v, want [hello]", terms)
	}
}
