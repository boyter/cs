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

// --- FindAllMatchingLines tests ---

func TestFindAllMatchingLinesEmpty(t *testing.T) {
	res := &common.FileJob{
		Content:        []byte{},
		MatchLocations: map[string][][]int{},
	}
	result := FindAllMatchingLines(res, -1, 0, 0)
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestFindAllMatchingLinesNoMatches(t *testing.T) {
	res := &common.FileJob{
		Content:        []byte("hello world\nfoo bar\n"),
		MatchLocations: map[string][][]int{},
	}
	result := FindAllMatchingLines(res, -1, 0, 0)
	if result != nil {
		t.Errorf("expected nil for no matches, got %v", result)
	}
}

func TestFindAllMatchingLinesReturnsAllMatches(t *testing.T) {
	content := []byte("foo bar\nbaz foo\nqux quux\nfoo again")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"foo": {{0, 3}, {12, 15}, {24, 27}},
		},
	}
	result := FindAllMatchingLines(res, -1, 0, 0)
	if len(result) != 3 {
		t.Fatalf("expected 3 matching lines, got %d", len(result))
	}
	// Lines 1, 2, 4 (1-based)
	expected := []int{1, 2, 4}
	for i, lr := range result {
		if lr.LineNumber != expected[i] {
			t.Errorf("result[%d]: expected line %d, got %d", i, expected[i], lr.LineNumber)
		}
	}
}

func TestFindAllMatchingLinesRespectsLimit(t *testing.T) {
	content := []byte("foo bar\nbaz foo\nqux quux\nfoo again")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"foo": {{0, 3}, {12, 15}, {24, 27}},
		},
	}
	result := FindAllMatchingLines(res, 2, 0, 0)
	if len(result) != 2 {
		t.Fatalf("expected 2 matching lines with limit=2, got %d", len(result))
	}
	if result[0].LineNumber != 1 || result[1].LineNumber != 2 {
		t.Errorf("expected lines 1,2 got %d,%d", result[0].LineNumber, result[1].LineNumber)
	}
}

func TestFindAllMatchingLinesNoContextLines(t *testing.T) {
	content := []byte("line one\nline two\nline three\nline four\nline five")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"three": {{19, 24}},
		},
	}
	result := FindAllMatchingLines(res, -1, 0, 0)
	if len(result) != 1 {
		t.Fatalf("expected exactly 1 line (no context), got %d", len(result))
	}
	if result[0].LineNumber != 3 {
		t.Errorf("expected line 3, got %d", result[0].LineNumber)
	}
}

func TestFindAllMatchingLines1Based(t *testing.T) {
	content := []byte("match here")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"match": {{0, 5}},
		},
	}
	result := FindAllMatchingLines(res, -1, 0, 0)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].LineNumber != 1 {
		t.Errorf("expected 1-based line number 1, got %d", result[0].LineNumber)
	}
}

func TestFindAllMatchingLinesCRLF(t *testing.T) {
	content := []byte("line one\r\nline two\r\nline three")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"two": {{15, 18}},
		},
	}
	result := FindAllMatchingLines(res, -1, 0, 0)
	if len(result) == 0 {
		t.Fatal("expected results")
	}
	if result[0].Content != "line two" {
		t.Errorf("expected 'line two' (no \\r), got %q", result[0].Content)
	}
}

func TestFindAllMatchingLinesLargeFile(t *testing.T) {
	// Build a 200-line file where every line contains the word "import"
	var lines []byte
	for i := 0; i < 200; i++ {
		if i > 0 {
			lines = append(lines, '\n')
		}
		line := []byte("import something")
		lines = append(lines, line...)
	}

	// Build match locations for "import" on every line
	var positions [][]int
	offset := 0
	for i := 0; i < 200; i++ {
		positions = append(positions, []int{offset, offset + 6})
		offset += len("import something") + 1
	}

	res := &common.FileJob{
		Content: lines,
		MatchLocations: map[string][][]int{
			"import": positions,
		},
	}
	result := FindAllMatchingLines(res, -1, 0, 0)
	if len(result) != 200 {
		t.Errorf("expected 200 matching lines (no cap), got %d", len(result))
	}
}

// --- FindAllMatchingLines context tests ---

func TestFindAllMatchingLinesContextNoOverlap(t *testing.T) {
	// 7 lines, match on line 5 (0-based index 4), -B 2 -A 2 → lines 3-7
	content := []byte("line one\nline two\nline three\nline four\nline five\nline six\nline seven")
	// "five" at byte 44-48 within "line five" (line starts at byte 39)
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"five": {{44, 48}},
		},
	}
	result := FindAllMatchingLines(res, -1, 2, 2)
	if len(result) != 5 {
		t.Fatalf("expected 5 lines (match + 2 before + 2 after), got %d", len(result))
	}
	expected := []int{3, 4, 5, 6, 7}
	for i, lr := range result {
		if lr.LineNumber != expected[i] {
			t.Errorf("result[%d]: expected line %d, got %d", i, expected[i], lr.LineNumber)
		}
	}
	// Only line 5 should have Locs
	for _, lr := range result {
		if lr.LineNumber == 5 {
			if len(lr.Locs) == 0 {
				t.Error("expected Locs on match line 5")
			}
		} else {
			if len(lr.Locs) != 0 {
				t.Errorf("expected no Locs on context line %d, got %v", lr.LineNumber, lr.Locs)
			}
		}
	}
}

func TestFindAllMatchingLinesContextOverlapping(t *testing.T) {
	// Matches on lines 1 and 3 (1-based), -C 1 → lines 1-4 merged
	content := []byte("foo bar\nbaz qux\nfoo quux\nend line")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"foo": {{0, 3}, {16, 19}},
		},
	}
	result := FindAllMatchingLines(res, -1, 1, 1)
	if len(result) != 4 {
		t.Fatalf("expected 4 lines (merged overlapping context), got %d", len(result))
	}
	expected := []int{1, 2, 3, 4}
	for i, lr := range result {
		if lr.LineNumber != expected[i] {
			t.Errorf("result[%d]: expected line %d, got %d", i, expected[i], lr.LineNumber)
		}
	}
	// Lines 1 and 3 are matches, 2 and 4 are context
	if len(result[0].Locs) == 0 {
		t.Error("line 1 should have Locs")
	}
	if len(result[1].Locs) != 0 {
		t.Error("line 2 should be context (no Locs)")
	}
	if len(result[2].Locs) == 0 {
		t.Error("line 3 should have Locs")
	}
	if len(result[3].Locs) != 0 {
		t.Error("line 4 should be context (no Locs)")
	}
}

func TestFindAllMatchingLinesContextAdjacentRanges(t *testing.T) {
	// Matches on lines 2 and 5 (1-based), -C 1 → two groups: 1-3, 4-6
	content := []byte("aaa\nbbb\nccc\nddd\neee\nfff")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"bbb": {{4, 7}},
			"eee": {{16, 19}},
		},
	}
	result := FindAllMatchingLines(res, -1, 1, 1)
	// Adjacent ranges (end of first = 2, start of second = 3) should merge → lines 1-6
	if len(result) != 6 {
		t.Fatalf("expected 6 lines (adjacent ranges merge), got %d", len(result))
	}
	expected := []int{1, 2, 3, 4, 5, 6}
	for i, lr := range result {
		if lr.LineNumber != expected[i] {
			t.Errorf("result[%d]: expected line %d, got %d", i, expected[i], lr.LineNumber)
		}
	}
}

func TestFindAllMatchingLinesContextClampedToBounds(t *testing.T) {
	// Match on line 1 with -B 3 → should clamp to start of file
	content := []byte("match\nsecond\nthird")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"match": {{0, 5}},
		},
	}
	result := FindAllMatchingLines(res, -1, 3, 0)
	if len(result) != 1 {
		t.Fatalf("expected 1 line (clamped before), got %d", len(result))
	}
	if result[0].LineNumber != 1 {
		t.Errorf("expected line 1, got %d", result[0].LineNumber)
	}
}

func TestFindAllMatchingLinesContextZeroBackwardCompat(t *testing.T) {
	// Zero context should return only match lines (same as before)
	content := []byte("aaa\nbbb\nccc\nddd")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"bbb": {{4, 7}},
		},
	}
	result := FindAllMatchingLines(res, -1, 0, 0)
	if len(result) != 1 {
		t.Fatalf("expected 1 line with zero context, got %d", len(result))
	}
	if result[0].LineNumber != 2 {
		t.Errorf("expected line 2, got %d", result[0].LineNumber)
	}
}

func TestFindAllMatchingLinesContextLimitAppliesToMatches(t *testing.T) {
	// 3 matches, limit=2, with -C 0 → only first 2 matches
	content := []byte("foo\nbar\nfoo\nbaz\nfoo")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"foo": {{0, 3}, {8, 11}, {16, 19}},
		},
	}
	result := FindAllMatchingLines(res, 2, 1, 1)
	// limit=2 means only 2 match lines; with -C 1 that's lines 1-4
	matchCount := 0
	for _, lr := range result {
		if len(lr.Locs) > 0 {
			matchCount++
		}
	}
	if matchCount != 2 {
		t.Errorf("expected 2 match lines with limit=2, got %d", matchCount)
	}
	// Should include context around both matches
	if len(result) < 3 {
		t.Errorf("expected at least 3 total lines (2 matches + context), got %d", len(result))
	}
}

func TestFindAllMatchingLinesContextBeforeOnly(t *testing.T) {
	// -B 2, -A 0: only lines before
	content := []byte("aaa\nbbb\nccc\nddd\neee")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"ddd": {{12, 15}},
		},
	}
	result := FindAllMatchingLines(res, -1, 2, 0)
	if len(result) != 3 {
		t.Fatalf("expected 3 lines (2 before + match), got %d", len(result))
	}
	expected := []int{2, 3, 4}
	for i, lr := range result {
		if lr.LineNumber != expected[i] {
			t.Errorf("result[%d]: expected line %d, got %d", i, expected[i], lr.LineNumber)
		}
	}
}

func TestFindAllMatchingLinesContextAfterOnly(t *testing.T) {
	// -B 0, -A 2: only lines after
	content := []byte("aaa\nbbb\nccc\nddd\neee")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"bbb": {{4, 7}},
		},
	}
	result := FindAllMatchingLines(res, -1, 0, 2)
	if len(result) != 3 {
		t.Fatalf("expected 3 lines (match + 2 after), got %d", len(result))
	}
	expected := []int{2, 3, 4}
	for i, lr := range result {
		if lr.LineNumber != expected[i] {
			t.Errorf("result[%d]: expected line %d, got %d", i, expected[i], lr.LineNumber)
		}
	}
}
