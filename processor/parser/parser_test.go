package parser

import "testing"

func TestParseArgumentsEmpty(t *testing.T) {
	res := ParseArguments([]string{})

	if len(res) != 0 {
		t.Error("Expected 0")
	}
}

func TestParseArgumentsSingle(t *testing.T) {
	res := ParseArguments([]string{"test"})

	if res[0].Term != "test" || res[0].Type != Default {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsFuzzy1(t *testing.T) {
	res := ParseArguments([]string{"test~1"})

	if res[0].Term != "test~1" || res[0].Type != Fuzzy1 {
		t.Error("Expected single match")
	}
}

func TestParseArgumentsFuzzy2(t *testing.T) {
	res := ParseArguments([]string{"test~2"})

	if res[0].Term != "test~2" || res[0].Type != Fuzzy2 {
		t.Error("Expected single match")
	}
}
