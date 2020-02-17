package processor

import "testing"

func TestParseArgumentsEmpty(t *testing.T) {
	res := parseArguments([]string{})

	if len(res) != 0 {
		t.Error("Expected 0")
	}
}

func TestParseArgumentsSingle(t *testing.T) {
	res := parseArguments([]string{"test"})

	if res[0].Term != "test" || res[0].Type != Default {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsTwo(t *testing.T) {
	res := parseArguments([]string{"test", "test"})

	if res[0].Term != "test" || res[0].Type != Default {
		t.Error("Expected single match")
	}

	if res[1].Term != "test" || res[1].Type != Default {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsFuzzy1(t *testing.T) {
	res := parseArguments([]string{"test~1"})

	if res[0].Term != "test~1" || res[0].Type != Fuzzy1 {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsFuzzy2(t *testing.T) {
	res := parseArguments([]string{"test~2"})

	if res[0].Term != "test~2" || res[0].Type != Fuzzy2 {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsNOTFirst(t *testing.T) {
	res := parseArguments([]string{"NOT"})

	if len(res) != 0 {
		t.Error("Expected 0")
	}
}

func TestParseArgumentsNOTSecond(t *testing.T) {
	res := parseArguments([]string{"test", "NOT"})

	if res[0].Term != `test` || res[0].Type != Default {
		t.Error("Expected single match")
	}

	if res[1].Term != `NOT` || res[1].Type != Negated {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsRegex(t *testing.T) {
	res := parseArguments([]string{"/test/"})

	if res[0].Term != "/test/" || res[0].Type != Regex {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsQuoted(t *testing.T) {
	res := parseArguments([]string{`"test"`})

	if res[0].Term != `"test"` || res[0].Type != Quoted {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsQuotedEmpty(t *testing.T) {
	res := parseArguments([]string{`""`})

	if res[0].Term != `""` || res[0].Type != Quoted {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsQuotedMultiple(t *testing.T) {
	res := parseArguments([]string{`"test`, `something"`})

	if res[0].Term != `"test something"` || res[0].Type != Quoted {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsRegexMultiple(t *testing.T) {
	res := parseArguments([]string{`/test`, `something/`})

	if res[0].Term != `/test something/` || res[0].Type != Regex {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsMultiple(t *testing.T) {
	res := parseArguments([]string{`test~1`, `NOT`, `/test`, `something/`})

	if res[0].Type != Fuzzy1 && res[1].Type != Negated && res[2].Type != Regex {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsNotClosedRegex(t *testing.T) {
	res := parseArguments([]string{`/test`})

	if res[0].Term != "/test/" && res[0].Type != Regex {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsNotClosedQuoted(t *testing.T) {
	res := parseArguments([]string{`"test`})

	if res[0].Term != `"test"` && res[0].Type != Regex {
		t.Error("Expected single match")
	}
}
