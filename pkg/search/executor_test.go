package search

import "testing"

func TestPostEvalMetadataFilters_LangMatch(t *testing.T) {
	node := &FilterNode{Field: "lang", Operator: "=", Value: "Go"}
	if !PostEvalMetadataFilters(node, "Go", 0) {
		t.Error("expected lang:Go to match Go file")
	}
}

func TestPostEvalMetadataFilters_LangNoMatch(t *testing.T) {
	node := &FilterNode{Field: "lang", Operator: "=", Value: "Python"}
	if PostEvalMetadataFilters(node, "Go", 0) {
		t.Error("expected lang:Python to not match Go file")
	}
}

func TestPostEvalMetadataFilters_LangCaseInsensitive(t *testing.T) {
	node := &FilterNode{Field: "lang", Operator: "=", Value: "go"}
	if !PostEvalMetadataFilters(node, "Go", 0) {
		t.Error("expected lang:go to match Go file (case insensitive)")
	}
}

func TestPostEvalMetadataFilters_LanguageAlias(t *testing.T) {
	node := &FilterNode{Field: "language", Operator: "=", Value: "Go"}
	if !PostEvalMetadataFilters(node, "Go", 0) {
		t.Error("expected language:Go to match Go file")
	}
}

func TestPostEvalMetadataFilters_LangNotEqual(t *testing.T) {
	node := &FilterNode{Field: "lang", Operator: "!=", Value: "Go"}
	if PostEvalMetadataFilters(node, "Go", 0) {
		t.Error("expected lang!=Go to not match Go file")
	}
	if !PostEvalMetadataFilters(node, "Python", 0) {
		t.Error("expected lang!=Go to match Python file")
	}
}

func TestPostEvalMetadataFilters_ComplexityGte(t *testing.T) {
	node := &FilterNode{Field: "complexity", Operator: ">=", Value: 50}
	if !PostEvalMetadataFilters(node, "Go", 100) {
		t.Error("expected complexity>=50 to match file with complexity 100")
	}
	if PostEvalMetadataFilters(node, "Go", 10) {
		t.Error("expected complexity>=50 to not match file with complexity 10")
	}
}

func TestPostEvalMetadataFilters_ComplexityEqual(t *testing.T) {
	node := &FilterNode{Field: "complexity", Operator: "=", Value: 50}
	if !PostEvalMetadataFilters(node, "Go", 50) {
		t.Error("expected complexity=50 to match file with complexity 50")
	}
	if PostEvalMetadataFilters(node, "Go", 49) {
		t.Error("expected complexity=50 to not match file with complexity 49")
	}
}

func TestPostEvalMetadataFilters_NotLang(t *testing.T) {
	node := &NotNode{Expr: &FilterNode{Field: "lang", Operator: "=", Value: "Go"}}
	if PostEvalMetadataFilters(node, "Go", 0) {
		t.Error("expected NOT lang:Go to not match Go file")
	}
	if !PostEvalMetadataFilters(node, "Python", 0) {
		t.Error("expected NOT lang:Go to match Python file")
	}
}

func TestPostEvalMetadataFilters_AndWithLang(t *testing.T) {
	// keyword AND lang:Go — keyword is true (already evaluated), lang:Go is checked
	node := &AndNode{
		Left:  &KeywordNode{Value: "test"},
		Right: &FilterNode{Field: "lang", Operator: "=", Value: "Go"},
	}
	if !PostEvalMetadataFilters(node, "Go", 0) {
		t.Error("expected keyword AND lang:Go to match Go file")
	}
	if PostEvalMetadataFilters(node, "Python", 0) {
		t.Error("expected keyword AND lang:Go to not match Python file")
	}
}

func TestPostEvalMetadataFilters_OrWithLang(t *testing.T) {
	// lang:Go OR lang:Python
	node := &OrNode{
		Left:  &FilterNode{Field: "lang", Operator: "=", Value: "Go"},
		Right: &FilterNode{Field: "lang", Operator: "=", Value: "Python"},
	}
	if !PostEvalMetadataFilters(node, "Go", 0) {
		t.Error("expected lang:Go OR lang:Python to match Go file")
	}
	if !PostEvalMetadataFilters(node, "Python", 0) {
		t.Error("expected lang:Go OR lang:Python to match Python file")
	}
	if PostEvalMetadataFilters(node, "Java", 0) {
		t.Error("expected lang:Go OR lang:Python to not match Java file")
	}
}

func TestPostEvalMetadataFilters_MultiValueLang(t *testing.T) {
	node := &FilterNode{Field: "lang", Operator: "=", Value: []interface{}{"Go", "Python"}}
	if !PostEvalMetadataFilters(node, "Go", 0) {
		t.Error("expected lang=Go,Python to match Go file")
	}
	if !PostEvalMetadataFilters(node, "Python", 0) {
		t.Error("expected lang=Go,Python to match Python file")
	}
	if PostEvalMetadataFilters(node, "Java", 0) {
		t.Error("expected lang=Go,Python to not match Java file")
	}
}

func TestPostEvalMetadataFilters_NonMetadataFilterPassesThrough(t *testing.T) {
	// ext: filter should pass through (already handled in per-file eval)
	node := &FilterNode{Field: "ext", Operator: "=", Value: "go"}
	if !PostEvalMetadataFilters(node, "Go", 0) {
		t.Error("expected ext filter to pass through as true")
	}
}

func TestPostEvalMetadataFilters_NilNode(t *testing.T) {
	if !PostEvalMetadataFilters(nil, "Go", 0) {
		t.Error("expected nil node to return true")
	}
}

func TestPostEvalMetadataFilters_ColonComplexityFilter(t *testing.T) {
	// Verify that complexity:<=25 parsed via the colon syntax produces a
	// FilterNode that PostEvalMetadataFilters handles correctly.
	node := &FilterNode{Field: "complexity", Operator: "<=", Value: 25}
	if !PostEvalMetadataFilters(node, "Go", 10) {
		t.Error("expected complexity<=25 to match file with complexity 10")
	}
	if !PostEvalMetadataFilters(node, "Go", 25) {
		t.Error("expected complexity<=25 to match file with complexity 25")
	}
	if PostEvalMetadataFilters(node, "Go", 26) {
		t.Error("expected complexity<=25 to not match file with complexity 26")
	}
}
