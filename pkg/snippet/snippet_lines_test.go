// SPDX-License-Identifier: MIT

package snippet

import (
	"testing"

	"github.com/boyter/cs/v3/pkg/common"
)

func TestFindMatchingLinesEmpty(t *testing.T) {
	res := &common.FileJob{
		Content:        []byte{},
		MatchLocations: map[string][][]int{},
	}
	result := FindMatchingLines(res, 2)
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestFindMatchingLinesNoMatches(t *testing.T) {
	res := &common.FileJob{
		Content:        []byte("hello world\nfoo bar\n"),
		MatchLocations: map[string][][]int{},
	}
	result := FindMatchingLines(res, 2)
	if result != nil {
		t.Errorf("expected nil for no matches, got %v", result)
	}
}

func TestFindMatchingLinesSingleMatch(t *testing.T) {
	content := []byte("line one\nline two\nline three\nline four\nline five")
	// Match "two" at byte offset 14-17 (within "line two\n")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"two": {{14, 17}},
		},
	}
	result := FindMatchingLines(res, 2)
	if len(result) == 0 {
		t.Fatal("expected results")
	}

	// The matching line should be line 2 (1-based)
	found := false
	for _, lr := range result {
		if lr.LineNumber == 2 {
			found = true
			if lr.Content != "line two" {
				t.Errorf("expected 'line two', got %q", lr.Content)
			}
			if len(lr.Locs) != 1 {
				t.Errorf("expected 1 loc, got %d", len(lr.Locs))
			} else {
				// "two" starts at position 5 within "line two"
				if lr.Locs[0][0] != 5 || lr.Locs[0][1] != 8 {
					t.Errorf("expected loc [5,8], got %v", lr.Locs[0])
				}
			}
			if lr.Score == 0 {
				t.Error("expected non-zero score")
			}
		}
	}
	if !found {
		t.Error("line 2 not found in results")
	}

	// Should include surrounding context (lines 1 and 3-4 with surroundLines=2)
	lineNums := make(map[int]bool)
	for _, lr := range result {
		lineNums[lr.LineNumber] = true
	}
	// Line 1 (context before), line 2 (match), lines 3-4 (context after)
	for _, expected := range []int{1, 2, 3, 4} {
		if !lineNums[expected] {
			t.Errorf("expected line %d in results", expected)
		}
	}
}

func TestFindMatchingLinesMultipleMatches(t *testing.T) {
	content := []byte("foo bar\nbaz foo\nqux quux\nfoo again")
	// "foo" appears on lines 1, 2, and 4
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"foo": {{0, 3}, {12, 15}, {24, 27}},
		},
	}
	result := FindMatchingLines(res, 1)
	if len(result) == 0 {
		t.Fatal("expected results")
	}

	// All lines with matches should have Locs
	matchCount := 0
	for _, lr := range result {
		if len(lr.Locs) > 0 {
			matchCount++
		}
	}
	if matchCount < 3 {
		t.Errorf("expected at least 3 matching lines, got %d", matchCount)
	}
}

func TestFindMatchingLinesSorted(t *testing.T) {
	content := []byte("aaa\nbbb\nccc match\nddd\neee match")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"match": {{8, 13}, {22, 27}},
		},
	}
	result := FindMatchingLines(res, 1)

	// Verify results are sorted by line number
	for i := 1; i < len(result); i++ {
		if result[i].LineNumber <= result[i-1].LineNumber {
			t.Errorf("results not sorted: line %d after line %d",
				result[i].LineNumber, result[i-1].LineNumber)
		}
	}
}

func TestFindMatchingLinesCRLF(t *testing.T) {
	content := []byte("line one\r\nline two\r\nline three")
	// "two" within "line two\r\n" - byte offsets account for \r\n
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"two": {{15, 18}},
		},
	}
	result := FindMatchingLines(res, 0)
	if len(result) == 0 {
		t.Fatal("expected results")
	}

	for _, lr := range result {
		if lr.LineNumber == 2 {
			if lr.Content != "line two" {
				t.Errorf("expected 'line two' (no \\r), got %q", lr.Content)
			}
			return
		}
	}
	t.Error("line 2 not found")
}

func TestFindMatchingLines1Based(t *testing.T) {
	content := []byte("match here")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"match": {{0, 5}},
		},
	}
	result := FindMatchingLines(res, 0)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].LineNumber != 1 {
		t.Errorf("expected 1-based line number 1, got %d", result[0].LineNumber)
	}
}

func TestFindMatchingLinesIgnoresShortTerms(t *testing.T) {
	// "a" appears on every line, "year" on one line only.
	// With minTermLen filtering, "a" should be ignored.
	content := []byte("a cat sat\na dog ran\na year passed")
	// "a" at positions 0,1  10,11  20,21
	// "year" at position 22-26
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"a":    {{0, 1}, {10, 11}, {20, 21}},
			"year": {{22, 26}},
		},
	}
	result := FindMatchingLines(res, 0)

	// Only line 3 ("a year passed") should match via "year"
	if len(result) != 1 {
		t.Fatalf("expected 1 matching line, got %d", len(result))
	}
	if result[0].LineNumber != 3 {
		t.Errorf("expected line 3, got line %d", result[0].LineNumber)
	}
	// Score should be 4.0 (one "year" hit), not inflated by "a"
	if result[0].Score != 4.0 {
		t.Errorf("expected score 4.0, got %f", result[0].Score)
	}
}

func TestFindMatchingLinesKeepsPhraseWithShortWords(t *testing.T) {
	// A phrase key like "ten thousand a year" contains "a" but the key
	// itself is long, so it should NOT be filtered out.
	content := []byte("ten thousand a year is nice")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"ten thousand a year": {{0, 20}},
		},
	}
	result := FindMatchingLines(res, 0)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Score == 0 {
		t.Error("expected non-zero score for phrase match")
	}
}

func TestFindMatchingLinesSingleCharOnly(t *testing.T) {
	// When all search terms are single characters, they should NOT be
	// filtered out — otherwise no snippets are produced at all.
	content := []byte("one\ntwo\nthree")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"e": {{2, 3}, {10, 11}, {12, 13}},
		},
	}
	result := FindMatchingLines(res, 0)
	if len(result) == 0 {
		t.Fatal("expected results for single-char-only search, got none")
	}

	// Lines 1 ("one") and 3 ("three") contain "e"
	lineNums := make(map[int]bool)
	for _, lr := range result {
		lineNums[lr.LineNumber] = true
	}
	if !lineNums[1] {
		t.Error("expected line 1 ('one') in results")
	}
	if !lineNums[3] {
		t.Error("expected line 3 ('three') in results")
	}

	// Verify match locations are populated
	for _, lr := range result {
		if len(lr.Locs) == 0 && (lr.LineNumber == 1 || lr.LineNumber == 3) {
			t.Errorf("expected Locs on line %d, got none", lr.LineNumber)
		}
	}
}

func TestAddPhraseMatchLocations(t *testing.T) {
	content := []byte("Ten Thousand a Year is a lot of money. ten thousand a year indeed.")

	t.Run("multi-word adds entry", func(t *testing.T) {
		locs := map[string][][]int{}
		AddPhraseMatchLocations(content, "ten thousand a year", locs)
		positions, ok := locs["ten thousand a year"]
		if !ok {
			t.Fatal("expected phrase entry in matchLocations")
		}
		if len(positions) != 2 {
			t.Fatalf("expected 2 phrase matches, got %d", len(positions))
		}
		// First match: "Ten Thousand a Year" starts at 0, length 19
		if positions[0][0] != 0 || positions[0][1] != 19 {
			t.Errorf("first match: expected [0,19], got %v", positions[0])
		}
	})

	t.Run("single-word is no-op", func(t *testing.T) {
		locs := map[string][][]int{}
		AddPhraseMatchLocations(content, "test", locs)
		if len(locs) != 0 {
			t.Errorf("expected no entries for single-word query, got %d", len(locs))
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		locs := map[string][][]int{}
		AddPhraseMatchLocations(content, "TEN THOUSAND A YEAR", locs)
		if _, ok := locs["TEN THOUSAND A YEAR"]; !ok {
			t.Fatal("expected case-insensitive match")
		}
	})

	t.Run("no match returns nothing", func(t *testing.T) {
		locs := map[string][][]int{}
		AddPhraseMatchLocations(content, "not found phrase", locs)
		if len(locs) != 0 {
			t.Errorf("expected no entries for non-matching phrase, got %d", len(locs))
		}
	})
}
