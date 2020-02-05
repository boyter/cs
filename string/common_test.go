package string

import "testing"

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
