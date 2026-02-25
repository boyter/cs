// SPDX-License-Identifier: MIT

package snippet

import (
	"bytes"
	"sort"
	"strings"

	"github.com/boyter/cs/v3/pkg/common"
)

// LineResult represents a single line from a search result,
// with match positions for highlighting.
type LineResult struct {
	LineNumber int     // 1-based line number
	Content    string  // plain text of the line
	Locs       [][]int // match positions within the line [start, end]
	Score      float64
}

const minTermLen = 2

// shouldFilterShortTerms returns true if MatchLocations contains at least
// one term that meets the minTermLen threshold, meaning short terms can
// safely be filtered out without losing all matches.
func shouldFilterShortTerms(matchLocations map[string][][]int) bool {
	for term := range matchLocations {
		if len(term) >= minTermLen {
			return true
		}
	}
	return false
}

// AddPhraseMatchLocations case-insensitively searches for the full query string
// in file content and adds any hits to the MatchLocations map. This gives a
// natural boost to lines containing the exact phrase since the key is long.
// Only operates on multi-word queries (contains a space).
func AddPhraseMatchLocations(content []byte, query string, matchLocations map[string][][]int) {
	if !strings.Contains(query, " ") {
		return
	}

	lowerContent := bytes.ToLower(content)
	lowerQuery := []byte(strings.ToLower(query))

	var positions [][]int
	idx := 0
	for {
		pos := bytes.Index(lowerContent[idx:], lowerQuery)
		if pos == -1 {
			break
		}
		start := idx + pos
		end := start + len(lowerQuery)
		positions = append(positions, []int{start, end})
		idx = start + len(lowerQuery)
	}

	if len(positions) > 0 {
		matchLocations[query] = positions
	}
}

// FindMatchingLines finds the most relevant matching lines in a file
// based on pre-computed match locations from the search pipeline,
// and returns them with surrounding context lines sorted by line number.
func FindMatchingLines(res *common.FileJob, surroundLines int) []LineResult {
	if len(res.MatchLocations) == 0 || len(res.Content) == 0 {
		return nil
	}

	const maxMatchingLines = 100
	const maxResultLines = 15

	// Split content into lines, tracking byte offsets
	rawLines := bytes.Split(res.Content, []byte("\n"))
	lineOffsets := make([]int, len(rawLines))
	offset := 0
	for i, line := range rawLines {
		lineOffsets[i] = offset
		offset += len(line) + 1 // +1 for the \n separator
	}

	// For each line, find match locations that overlap with it
	var matchingLines []LineResult
	for i, rawLine := range rawLines {
		lineStart := lineOffsets[i]
		lineEnd := lineStart + len(rawLine)

		var locs [][]int
		var score float64

		filterShort := shouldFilterShortTerms(res.MatchLocations)
		for term, positions := range res.MatchLocations {
			if filterShort && len(term) < minTermLen {
				continue
			}
			for _, pos := range positions {
				mStart, mEnd := pos[0], pos[1]

				// Check if the match overlaps with this line
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
					score += 4.0
				}
			}
		}

		if len(locs) > 0 {
			content := strings.TrimRight(string(rawLine), "\r")
			// Clamp locs if \r was trimmed
			for j := range locs {
				if locs[j][1] > len(content) {
					locs[j][1] = len(content)
				}
			}

			matchingLines = append(matchingLines, LineResult{
				LineNumber: i, // 0-based for now
				Content:    content,
				Locs:       locs,
				Score:      score,
			})

			if len(matchingLines) > maxMatchingLines {
				break
			}
		}
	}

	if len(matchingLines) == 0 {
		return nil
	}

	// Sort by score (best matches first)
	sort.Slice(matchingLines, func(i, j int) bool {
		return matchingLines[i].Score > matchingLines[j].Score
	})

	// Build result with surrounding context lines
	var resultLines []LineResult
	for _, ml := range matchingLines {
		if containsLineNumber(ml.LineNumber, resultLines) {
			continue
		}
		resultLines = append(resultLines, ml)

		// Add surrounding context lines
		for d := 1; d <= surroundLines; d++ {
			before := ml.LineNumber - d
			if before >= 0 && !containsLineNumber(before, resultLines) {
				if existing, found := findInLineResults(before, matchingLines); found {
					resultLines = append(resultLines, existing)
				} else {
					content := strings.TrimRight(string(rawLines[before]), "\r")
					resultLines = append(resultLines, LineResult{
						LineNumber: before,
						Content:    content,
					})
				}
			}

			after := ml.LineNumber + d
			if after < len(rawLines) && !containsLineNumber(after, resultLines) {
				if existing, found := findInLineResults(after, matchingLines); found {
					resultLines = append(resultLines, existing)
				} else {
					content := strings.TrimRight(string(rawLines[after]), "\r")
					resultLines = append(resultLines, LineResult{
						LineNumber: after,
						Content:    content,
					})
				}
			}
		}

		if len(resultLines) >= maxResultLines {
			break
		}
	}

	// Deduplicate
	seen := make(map[int]struct{})
	var clean []LineResult
	for _, lr := range resultLines {
		if _, ok := seen[lr.LineNumber]; !ok {
			seen[lr.LineNumber] = struct{}{}
			clean = append(clean, lr)
		}
	}

	// Sort by line number
	sort.Slice(clean, func(i, j int) bool {
		return clean[i].LineNumber < clean[j].LineNumber
	})

	// Convert to 1-based line numbers
	for i := range clean {
		clean[i].LineNumber++
	}

	return clean
}

// FindAllMatchingLines returns every line with at least one match, in file order.
// Unlike FindMatchingLines it has no artificial cap on the number of lines,
// does not add context lines, and does not sort by score.
// When limit > 0, at most that many matching lines are returned.
// When limit <= 0, all matching lines are returned.
// contextBefore/contextAfter add surrounding context lines (like grep -B/-A).
// Context lines have nil Locs. The limit applies to match count, not total lines.
func FindAllMatchingLines(res *common.FileJob, limit int, contextBefore, contextAfter int) []LineResult {
	if len(res.MatchLocations) == 0 || len(res.Content) == 0 {
		return nil
	}

	// Split content into lines, tracking byte offsets
	rawLines := bytes.Split(res.Content, []byte("\n"))
	lineOffsets := make([]int, len(rawLines))
	offset := 0
	for i, line := range rawLines {
		lineOffsets[i] = offset
		offset += len(line) + 1
	}

	filterShort := shouldFilterShortTerms(res.MatchLocations)

	// Build per-line match locs
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

	// No context requested — fast path
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

	// Build ranges and merge overlapping/adjacent ones
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

	// Merge overlapping/adjacent ranges
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

	// Build a lookup from line index to match locs
	matchByLine := make(map[int][][]int, len(matches))
	for _, m := range matches {
		matchByLine[m.index] = m.locs
	}

	// Emit results for each merged range
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

func containsLineNumber(lineNum int, results []LineResult) bool {
	for _, r := range results {
		if r.LineNumber == lineNum {
			return true
		}
	}
	return false
}

func findInLineResults(lineNum int, results []LineResult) (LineResult, bool) {
	for _, r := range results {
		if r.LineNumber == lineNum {
			return r, true
		}
	}
	return LineResult{}, false
}
