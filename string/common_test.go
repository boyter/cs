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

func TestPermuteCaseFoldingUnicodeNoFolding(t *testing.T) {
	permutations := PermuteCaseFolding("ſ")

	if len(permutations) != 3 {
		t.Error("Expected 3 returns")
	}
}

func TestPermuteCaseFolding(t *testing.T) {
	folded := PermuteCaseFolding("nſ")

	if len(folded) != 6 {
		t.Error("Should get 6 results got", len(folded))
	}
}

func TestPermuteCaseFoldingNumbers(t *testing.T) {
	folded := PermuteCaseFolding("07123E1")

	if len(folded) != 6 {
		t.Error("Should get 6 results got", len(folded))
	}
}

func TestPermuteCaseFoldingComparison(t *testing.T) {
	r1 := PermuteCase("groß")
	r2 := PermuteCaseFolding("groß")

	if len(r1) >= len(r2) {
		t.Error("Should not be of equal length")
	}
}
