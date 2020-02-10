package string

import (
	"regexp"
	"testing"
)

func BenchmarkFindAllIndexCaseInsensitive(b *testing.B) {
	r := regexp.MustCompile(`(?i)test`)
	haystack := []byte(test_MatchEndCase)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkIndexesAllIgnoreCaseCaseInsensitive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexesAllIgnoreCase("test", test_MatchEndCase, -1)

		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkFindAllIndexLargeCaseInsensitive(b *testing.B) {
	r := regexp.MustCompile(`(?i)test`)
	haystack := []byte(test_MatchEndCaseLarge)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkIndexesAllIgnoreCaseLargeCaseInsensitive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexesAllIgnoreCase("test", test_MatchEndCaseLarge, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkFindAllIndexUnicodeCaseInsensitive(b *testing.B) {
	r := regexp.MustCompile(`(?i)test`)
	haystack := []byte(test_UnicodeMatchEndCase)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkIndexesAllIgnoreCaseUnicodeCaseInsensitive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexesAllIgnoreCase("test", test_UnicodeMatchEndCase, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkFindAllIndexUnicodeLargeCaseInsensitive(b *testing.B) {
	r := regexp.MustCompile(`(?i)test`)
	haystack := []byte(test_UnicodeMatchEndCaseLarge)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkIndexesAllIgnoreCaseUnicodeLargeCaseInsensitive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexesAllIgnoreCase("test", test_UnicodeMatchEndCaseLarge, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

// This benchmark simulates a bad case of there being many
// partial matches where the first character in the needle
// can be found throughout the haystack
func BenchmarkFindAllIndexManyPartialMatchesCaseInsensitive(b *testing.B) {
	r := regexp.MustCompile(`(?i)1test`)
	haystack := []byte(test_MatchEndCase)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkIndexesAllIgnoreCaseManyPartialMatchesCaseInsensitive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexesAllIgnoreCase("1test", test_MatchEndCase, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

// This benchmark simulates a bad case of there being many
// partial matches where the first character in the needle
// can be found throughout the haystack
func BenchmarkFindAllIndexUnicodeManyPartialMatchesCaseInsensitive(b *testing.B) {
	r := regexp.MustCompile(`(?i)Ⱥtest`)
	haystack := []byte(test_UnicodeMatchEndCase)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkIndexesAllIgnoreCaseUnicodeManyPartialMatchesCaseInsensitive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexesAllIgnoreCase("Ⱥtest", test_UnicodeMatchEndCase, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkFindAllIndexUnicodeCaseInsensitiveVeryLarge(b *testing.B) {
	var large string
	for i := 0; i <= 100; i++ {
		large += test_UnicodeMatchEndCaseLarge
	}

	r := regexp.MustCompile(`(?i)Ⱥtest`)
	haystack := []byte(large)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 101 {
			b.Error("Expected single match got", len(matches))
		}
	}
}

func BenchmarkIndexesAllIgnoreCaseUnicodeCaseInsensitiveVeryLarge(b *testing.B) {
	var large string
	for i := 0; i <= 100; i++ {
		large += test_UnicodeMatchEndCaseLarge
	}

	for i := 0; i < b.N; i++ {
		matches := IndexesAllIgnoreCase("Ⱥtest", large, -1)
		if len(matches) != 101 {
			b.Error("Expected single match got", len(matches))
		}
	}
}

func BenchmarkFindAllIndexFoldingCaseInsensitiveVeryLarge(b *testing.B) {
	var large string
	for i := 0; i <= 100; i++ {
		large += test_UnicodeMatchEndCaseLarge
	}

	r := regexp.MustCompile(`(?i)ſ`)
	haystack := []byte(large)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 101 {
			b.Error("Expected single match got", len(matches))
		}
	}
}

func BenchmarkIndexesAllIgnoreCaseFoldingCaseInsensitiveVeryLarge(b *testing.B) {
	var large string
	for i := 0; i <= 100; i++ {
		large += test_UnicodeMatchEndCaseLarge
	}

	for i := 0; i < b.N; i++ {
		matches := IndexesAllIgnoreCase("ſ", large, -1)
		if len(matches) != 101 {
			b.Error("Expected single match got", len(matches))
		}
	}
}
