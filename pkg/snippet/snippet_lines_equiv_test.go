// SPDX-License-Identifier: MIT

package snippet

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/boyter/cs/v3/pkg/common"
)

// findAllMatchingLinesSlow is the original O(lines*matches) implementation of
// FindAllMatchingLines, kept verbatim as the reference the fast version is
// checked against. Do not "fix" it — its whole value is being the old code.
func findAllMatchingLinesSlow(res *common.FileJob, limit int, contextBefore, contextAfter int) []LineResult {
	if len(res.MatchLocations) == 0 || len(res.Content) == 0 {
		return nil
	}

	rawLines := bytes.Split(res.Content, []byte("\n"))
	lineOffsets := make([]int, len(rawLines))
	offset := 0
	for i, line := range rawLines {
		lineOffsets[i] = offset
		offset += len(line) + 1
	}

	filterShort := shouldFilterShortTerms(res.MatchLocations)

	type lineMatch struct {
		index int
		locs  [][]int
	}
	var matches []lineMatch
	for i, rawLine := range rawLines {
		lineStart := lineOffsets[i]
		lineEnd := lineStart + len(rawLine)

		var locs [][]int

		for term, positions := range res.MatchLocations {
			if filterShort && len(term) < minTermLen {
				continue
			}
			for _, pos := range positions {
				mStart, mEnd := pos[0], pos[1]
				if mStart < lineEnd && mEnd > lineStart {
					relStart := mStart - lineStart
					relEnd := mEnd - lineStart
					if relStart < 0 {
						relStart = 0
					}
					if relEnd > len(rawLine) {
						relEnd = len(rawLine)
					}
					locs = append(locs, []int{relStart, relEnd})
				}
			}
		}

		if len(locs) > 0 {
			matches = append(matches, lineMatch{index: i, locs: locs})
			if limit > 0 && len(matches) >= limit {
				break
			}
		}
	}

	if len(matches) == 0 {
		return nil
	}

	if contextBefore <= 0 && contextAfter <= 0 {
		results := make([]LineResult, len(matches))
		for i, m := range matches {
			content := strings.TrimRight(string(rawLines[m.index]), "\r")
			for j := range m.locs {
				if m.locs[j][1] > len(content) {
					m.locs[j][1] = len(content)
				}
			}
			results[i] = LineResult{
				LineNumber: m.index + 1,
				Content:    content,
				Locs:       m.locs,
			}
		}
		return results
	}

	type lineRange struct{ start, end int }
	ranges := make([]lineRange, 0, len(matches))
	maxLine := len(rawLines) - 1
	for _, m := range matches {
		s := m.index - contextBefore
		if s < 0 {
			s = 0
		}
		e := m.index + contextAfter
		if e > maxLine {
			e = maxLine
		}
		ranges = append(ranges, lineRange{s, e})
	}

	merged := []lineRange{ranges[0]}
	for _, r := range ranges[1:] {
		last := &merged[len(merged)-1]
		if r.start <= last.end+1 {
			if r.end > last.end {
				last.end = r.end
			}
		} else {
			merged = append(merged, r)
		}
	}

	matchByLine := make(map[int][][]int, len(matches))
	for _, m := range matches {
		matchByLine[m.index] = m.locs
	}

	var results []LineResult
	for _, r := range merged {
		for i := r.start; i <= r.end; i++ {
			content := strings.TrimRight(string(rawLines[i]), "\r")
			lr := LineResult{
				LineNumber: i + 1,
				Content:    content,
			}
			if locs, ok := matchByLine[i]; ok {
				for j := range locs {
					if locs[j][1] > len(content) {
						locs[j][1] = len(content)
					}
				}
				lr.Locs = locs
			}
			results = append(results, lr)
		}
	}

	return results
}

// sortLocs normalises the per-line Locs ordering. The old implementation
// emitted them in map iteration order, so both sides have to be sorted before
// they can be compared.
func sortLocs(results []LineResult) []LineResult {
	out := make([]LineResult, len(results))
	for i, lr := range results {
		locs := make([][]int, len(lr.Locs))
		copy(locs, lr.Locs)
		sort.Slice(locs, func(a, b int) bool {
			if locs[a][0] != locs[b][0] {
				return locs[a][0] < locs[b][0]
			}
			return locs[a][1] < locs[b][1]
		})
		lr.Locs = locs
		if len(locs) == 0 {
			lr.Locs = nil
		}
		out[i] = lr
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// randomCase builds a file and a match set with the awkward shapes that break
// naive implementations: multi-line matches, overlapping matches, matches on
// the first and last lines, empty lines, CRLF endings and short terms.
func randomCase(rnd *rand.Rand) (*common.FileJob, string) {
	numLines := 1 + rnd.Intn(30)
	crlf := rnd.Intn(2) == 0

	var buf bytes.Buffer
	for i := 0; i < numLines; i++ {
		if i > 0 {
			if crlf {
				buf.WriteString("\r")
			}
			buf.WriteString("\n")
		}
		// Length 0 lines are deliberately possible, including as the last line.
		lineLen := rnd.Intn(12)
		for j := 0; j < lineLen; j++ {
			buf.WriteByte(byte('a' + rnd.Intn(6)))
		}
	}
	content := buf.Bytes()

	matchLocations := map[string][][]int{}
	if len(content) > 0 {
		numTerms := 1 + rnd.Intn(3)
		for t := 0; t < numTerms; t++ {
			// Term length of 1 exercises the filterShort path, since
			// shouldFilterShortTerms only kicks in when a longer term exists.
			term := strings.Repeat(string(rune('a'+rnd.Intn(6))), 1+rnd.Intn(4))
			numPos := rnd.Intn(8)
			var positions [][]int
			for p := 0; p < numPos; p++ {
				start := rnd.Intn(len(content))
				// Spans of up to 20 bytes reliably cross line boundaries.
				end := start + rnd.Intn(20)
				if end > len(content) {
					end = len(content)
				}
				positions = append(positions, []int{start, end})
			}
			if len(positions) > 0 {
				matchLocations[term] = positions
			}
		}
	}

	return &common.FileJob{Content: content, MatchLocations: matchLocations},
		fmt.Sprintf("lines=%d crlf=%v content=%q locs=%v", numLines, crlf, content, matchLocations)
}

func TestFindAllMatchingLinesEquivalence(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))

	limits := []int{-1, 0, 1, 2, 5, 1000}
	contexts := [][2]int{{0, 0}, {1, 1}, {2, 0}, {0, 2}, {3, 3}}

	for iter := 0; iter < 2000; iter++ {
		res, desc := randomCase(rnd)

		for _, limit := range limits {
			for _, ctx := range contexts {
				// Each implementation clamps Locs in place, so give both a
				// pristine copy of the input.
				want := sortLocs(findAllMatchingLinesSlow(cloneFileJob(res), limit, ctx[0], ctx[1]))
				got := sortLocs(FindAllMatchingLines(cloneFileJob(res), limit, ctx[0], ctx[1]))

				if !reflect.DeepEqual(want, got) {
					t.Fatalf("mismatch (limit=%d before=%d after=%d)\n%s\nwant %+v\ngot  %+v",
						limit, ctx[0], ctx[1], desc, want, got)
				}
			}
		}
	}
}

func cloneFileJob(res *common.FileJob) *common.FileJob {
	locs := make(map[string][][]int, len(res.MatchLocations))
	for term, positions := range res.MatchLocations {
		cp := make([][]int, len(positions))
		for i, p := range positions {
			cp[i] = []int{p[0], p[1]}
		}
		locs[term] = cp
	}
	return &common.FileJob{Content: res.Content, MatchLocations: locs}
}

// TestFindAllMatchingLinesMultiLineMatch pins the interval-overlap behaviour
// down explicitly: a match that spans a line boundary belongs to every line it
// touches, clamped to each line.
func TestFindAllMatchingLinesMultiLineMatch(t *testing.T) {
	content := []byte("aaaa\nbbbb\ncccc")
	res := &common.FileJob{
		Content: content,
		// Byte 2 through 12 covers the tail of line 1, all of line 2 and the
		// head of line 3.
		MatchLocations: map[string][][]int{"span": {{2, 12}}},
	}

	result := FindAllMatchingLines(res, -1, 0, 0)
	if len(result) != 3 {
		t.Fatalf("expected the match to land on 3 lines, got %d: %+v", len(result), result)
	}

	expected := [][]int{{2, 4}, {0, 4}, {0, 2}}
	for i, lr := range result {
		if lr.LineNumber != i+1 {
			t.Errorf("result[%d]: expected line %d, got %d", i, i+1, lr.LineNumber)
		}
		if !reflect.DeepEqual(lr.Locs, [][]int{expected[i]}) {
			t.Errorf("result[%d]: expected locs %v, got %v", i, [][]int{expected[i]}, lr.Locs)
		}
	}
}

// TestFindAllMatchingLinesLocsSorted covers the one deliberate behaviour
// change: within a line, Locs come out ordered by start offset regardless of
// map iteration order.
func TestFindAllMatchingLinesLocsSorted(t *testing.T) {
	content := []byte("zebra apple mango cherry apple")
	res := &common.FileJob{
		Content: content,
		MatchLocations: map[string][][]int{
			"cherry": {{18, 24}},
			"apple":  {{25, 30}, {6, 11}},
			"mango":  {{12, 17}},
			"zebra":  {{0, 5}},
		},
	}

	for i := 0; i < 50; i++ {
		result := FindAllMatchingLines(res, -1, 0, 0)
		if len(result) != 1 {
			t.Fatalf("expected 1 line, got %d", len(result))
		}
		want := [][]int{{0, 5}, {6, 11}, {12, 17}, {18, 24}, {25, 30}}
		if !reflect.DeepEqual(result[0].Locs, want) {
			t.Fatalf("locs not sorted by start offset: want %v got %v", want, result[0].Locs)
		}
	}
}

// buildPathologicalFile mimics the real shape that surfaced this: generated
// CloudFormation, ~30k lines with ~17k matches spread across them.
func buildPathologicalFile() *common.FileJob {
	const numLines = 30000
	const matchEvery = 2 // ~15k matches

	var buf bytes.Buffer
	var positions [][]int
	for i := 0; i < numLines; i++ {
		start := buf.Len()
		buf.WriteString(`      "Description": "some resource description here",`)
		if i%matchEvery == 0 {
			positions = append(positions, []int{start + 7, start + 18})
		}
		buf.WriteString("\n")
	}

	return &common.FileJob{
		Content:        buf.Bytes(),
		MatchLocations: map[string][][]int{"Description": positions},
	}
}

func BenchmarkFindAllMatchingLinesPathological(b *testing.B) {
	res := buildPathologicalFile()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if r := FindAllMatchingLines(res, -1, 0, 0); len(r) == 0 {
			b.Fatal("expected results")
		}
	}
}

func BenchmarkFindAllMatchingLinesPathologicalSlow(b *testing.B) {
	res := buildPathologicalFile()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if r := findAllMatchingLinesSlow(res, -1, 0, 0); len(r) == 0 {
			b.Fatal("expected results")
		}
	}
}

func BenchmarkFindAllMatchingLinesSmall(b *testing.B) {
	var buf bytes.Buffer
	var positions [][]int
	for i := 0; i < 200; i++ {
		start := buf.Len()
		buf.WriteString("func something(name string) error {")
		if i%20 == 0 {
			positions = append(positions, []int{start, start + 4})
		}
		buf.WriteString("\n")
	}
	res := &common.FileJob{Content: buf.Bytes(), MatchLocations: map[string][][]int{"func": positions}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if r := FindAllMatchingLines(res, -1, 0, 0); len(r) == 0 {
			b.Fatal("expected results")
		}
	}
}

func BenchmarkFindAllMatchingLinesSmallSlow(b *testing.B) {
	var buf bytes.Buffer
	var positions [][]int
	for i := 0; i < 200; i++ {
		start := buf.Len()
		buf.WriteString("func something(name string) error {")
		if i%20 == 0 {
			positions = append(positions, []int{start, start + 4})
		}
		buf.WriteString("\n")
	}
	res := &common.FileJob{Content: buf.Bytes(), MatchLocations: map[string][][]int{"func": positions}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if r := findAllMatchingLinesSlow(res, -1, 0, 0); len(r) == 0 {
			b.Fatal("expected results")
		}
	}
}
