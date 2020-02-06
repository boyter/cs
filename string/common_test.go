package string

import (
	"fmt"
	"testing"
	"unicode"
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

func TestPermuteCaseUnicode2(t *testing.T) {
	permutations := PermuteCase("ſ")

	if len(permutations) != 3 {
		t.Error("Expected 3 returns")
	}
}

func TestSimpleFoldStuff(t *testing.T) {
	var s rune = 'ß'

	for i := 0; i < 3; i++ {
		fmt.Printf("%#U\n", unicode.SimpleFold(s))
		s = unicode.SimpleFold(s)
	}
}


var simpleFoldTests = []string{
	// SimpleFold(x) returns the next equivalent rune > x or wraps
	// around to smaller values.

	// Easy cases.
	"Aa",
	"δΔ",

	// ASCII special cases.
	"KkK",
	"Ssſ",

	// Non-ASCII special cases.
	"ρϱΡ",
	"ͅΙιι",

	// Extra special cases: has lower/upper but no case fold.
	"İ",
	"ı",

	// Upper comes before lower (Cherokee).
	"\u13b0\uab80",
}

func TestSimpleFold(t *testing.T) {
	for _, tt := range simpleFoldTests {
		cycle := []rune(tt)
		r := cycle[len(cycle)-1]
		for _, out := range cycle {
			if r := unicode.SimpleFold(r); r != out {
				t.Errorf("SimpleFold(%#U) = %#U, want %#U", r, r, out)
			}
			r = out
		}
	}

	if r := unicode.SimpleFold(-42); r != -42 {
		t.Errorf("SimpleFold(-42) = %v, want -42", r)
	}
}