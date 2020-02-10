package string

import (
	"testing"
)

func BenchmarkHorspool(b *testing.B) {
	var large string
	for i := 0; i <= 100; i++ {
		large += test_UnicodeMatchEndCaseLarge
	}
	l := []rune(large)

	var horspool = Horspool{}

	for i := 0; i < b.N; i++ {
		matches := horspool.Search(l, []rune("test"))
		if len(matches) != 101 {
			b.Error("Expected single match got", len(matches))
		}
	}
}


func BenchmarkHorspool2(b *testing.B) {
	var horspool = Horspool{}

	for i := 0; i < b.N; i++ {
		matches := horspool.Search([]rune(test_MatchEndCaseLarge), []rune("1test"))
		if len(matches) != 1 {
			b.Error("Expected single match got", len(matches))
		}
	}
}

func BenchmarkIndexesAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexAll(test_MatchEndCaseLarge, "1test", -1)
		if len(matches) != 1 {
			b.Error("Expected single match got", len(matches))
		}
	}
}

