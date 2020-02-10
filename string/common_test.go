package string

import (
	"testing"
)

func TestRemoveStringDuplicates(t *testing.T) {
	r := []string{"test", "test"}

	if len(RemoveStringDuplicates(r)) != 1 {
		t.Error("Expected a single return")
	}
}

func TestPermuteCase(t *testing.T) {
	permutations := PermuteCase("fo")

	if len(permutations) != 4 {
		t.Error("Expected 4 returns")
	}
}

func TestPermuteCaseUnicode(t *testing.T) {
	permutations := PermuteCase("ȺȾ")

	if len(permutations) != 4 {
		t.Error("Expected 4 returns")
	}
}

func TestPermuteCaseUnicodeNoFolding(t *testing.T) {
	permutations := PermuteCase("ſ")

	if len(permutations) != 2 {
		t.Error("Expected 2 returns")
	}
}

func TestPermuteCaseFoldingUnicodeNoFolding(t *testing.T) {
	permutations := PermuteCase("ſ")

	if len(permutations) != 3 {
		t.Error("Expected 3 returns")
	}
}

func TestAllSimpleFoldAsciiNumber(t *testing.T) {
	folded := AllSimpleFold('1')

	if len(folded) != 1 {
		t.Error("Should get 1 result")
	}
}

func TestAllSimpleFoldAsciiLetter(t *testing.T) {
	folded := AllSimpleFold('z')

	if len(folded) != 2 {
		t.Error("Should get 2 results")
	}
}

func TestAllSimpleFoldMultipleReturn(t *testing.T) {
	folded := AllSimpleFold('ſ')

	if len(folded) != 3 {
		t.Error("Should get 3 results")
	}
}

func TestAllSimpleFoldNotFullFold(t *testing.T) {
	// ß (assuming I copied the lowercase one)
	// can with full fold rules turn into SS
	// https://www.w3.org/TR/charmod-norm/#definitionCaseFolding
	// however in this case its a simple fold
	// so we would not expect that
	folded := AllSimpleFold('ß')

	if len(folded) != 2 {
		t.Error("Should get 2 results")
	}
}