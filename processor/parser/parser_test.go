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
