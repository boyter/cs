// SPDX-License-Identifier: MIT OR Unlicense

package snippet

import (
	"bytes"
	"sort"
	"unicode"

	str "github.com/boyter/go-string"

	"github.com/boyter/cs/pkg/common"
)

const (
	SnipSideMax int = 10 // Defines the maximum bytes either side of the match we are willing to return
	// The below are used for adding boosts to match conditions of snippets to hopefully produce the best match
	PhraseHeavyBoost = 20
	SpaceBoundBoost  = 5
	ExactMatchBoost  = 5
	// Below is used to control CPU burn time trying to find the most relevant snippet
	RelevanceCutoff = 10_000
)

type bestMatch struct {
	StartPos int
	EndPos   int
	Score    float64
	Relevant []relevantV3
}

// Internal structure used just for matching things together
type relevantV3 struct {
	Word  string
	Start int
	End   int
}

type Snippet struct {
	Content   string
	StartPos  int
	EndPos    int
	Score     float64
	LineStart int
	LineEnd   int
}

// ExtractRelevant looks through the locations using a sliding window style algorithm
// where it "brute forces" the solution by iterating over every location we have
// and look for all matches that fall into the supplied length and ranking
// based on how many we have.
//
// This algorithm ranks using document frequencies that are kept for
// TF/IDF ranking with various other checks.
func ExtractRelevant(res *common.FileJob, documentFrequencies map[string]int, relLength int) []Snippet {
	wrapLength := relLength / 2
	var bestMatches []bestMatch

	rv3 := convertToRelevant(res)

	// if we have a huge amount of matches we want to reduce it because otherwise it takes forever
	// to return something if the search has many matches.
	if len(rv3) > RelevanceCutoff {
		rv3 = rv3[:RelevanceCutoff]
	}

	// Slide around looking for matches that fit in the length
	for i := 0; i < len(rv3); i++ {
		m := bestMatch{
			StartPos: rv3[i].Start,
			EndPos:   rv3[i].End,
			Relevant: []relevantV3{rv3[i]},
		}

		// Slide left
		j := i - 1
		for {
			// Ensure we never step outside the bounds of our slice
			if j < 0 {
				break
			}

			// How close is the matches start to our end?
			diff := rv3[i].End - rv3[j].Start

			// If the diff is greater than the target then break out as there is no
			// more reason to keep looking as the slice is sorted
			if diff > wrapLength {
				break
			}

			// If we didn't break this is considered a larger match
			m.StartPos = rv3[j].Start
			m.Relevant = append(m.Relevant, rv3[j])
			j--
		}

		// Slide right
		j = i + 1
		for {
			// Ensure we never step outside the bounds of our slice
			if j >= len(rv3) {
				break
			}

			// How close is the matches end to our start?
			diff := rv3[j].End - rv3[i].Start

			// If the diff is greater than the target then break out as there is no
			// more reason to keep looking as the slice is sorted
			if diff > wrapLength {
				break
			}

			m.EndPos = rv3[j].End
			m.Relevant = append(m.Relevant, rv3[j])
			j++
		}

		// If the match around this isn't long enough expand it out
		// roughly based on how large a context we need to add
		l := m.EndPos - m.StartPos
		if l < relLength {
			add := (relLength - l) / 2
			m.StartPos -= add
			m.EndPos += add

			if m.StartPos < 0 {
				m.StartPos = 0
			}

			if m.EndPos > len(res.Content) {
				m.EndPos = len(res.Content)
			}
		}

		// Now we see if there are any nearby spaces to avoid us cutting in the
		// middle of a word if we can avoid it
		sf := false
		ef := false
		m.StartPos, sf = findSpaceLeft(res, m.StartPos, SnipSideMax)
		m.EndPos, ef = findSpaceRight(res, m.EndPos, SnipSideMax)

		// Check if we are cutting in the middle of a multibyte char and if so
		// go looking till we find the start. We only do so if we didn't find a space,
		// and if we aren't at the start or very end of the content
		for !sf && m.StartPos != 0 && m.StartPos != len(res.Content) && !str.StartOfRune(res.Content[m.StartPos]) {
			m.StartPos--
		}
		for !ef && m.EndPos != 0 && m.EndPos != len(res.Content) && !str.StartOfRune(res.Content[m.EndPos]) {
			m.EndPos--
		}

		// If we are very close to the start, just push it out so we get the actual start
		if m.StartPos <= SnipSideMax {
			m.StartPos = 0
		}
		// As above, but against the end so we just include the rest if we are close
		if len(res.Content)-m.EndPos <= 10 {
			m.EndPos = len(res.Content)
		}

		// Now that we have the snippet start to rank it to produce a score indicating
		// how good a match it is and hopefully display to the user what they
		// were actually looking for
		m.Score += float64(len(m.Relevant)) // Factor in how many matches we have

		// Apply higher score where the words are near each other
		mid := rv3[i].Start
		for _, v := range m.Relevant {
			p := v.Start + (v.End-v.Start)/2 // comparison word midpoint

			// If the word is within a reasonable distance of this word boost the score
			// weighted by how common that word is so that matches like 'a' impact the rank
			// less than something like 'cromulent' which in theory should not occur as much
			if abs(mid-p) < (relLength / 3) {
				m.Score += 100 / float64(documentFrequencies[v.Word])
			}
		}

		// Try to make it phrase heavy such that if words line up next to each other
		// it is given a much higher weight
		for _, v := range m.Relevant {
			// Use 2 here because we want to avoid punctuation such that a search for
			// cat dog will still be boosted if we find cat. dog
			if abs(rv3[i].Start-v.End) <= 2 || abs(rv3[i].End-v.Start) <= 2 {
				m.Score += PhraseHeavyBoost
			}
		}

		// If the match is bounded by a space boost it slightly
		// because its likely to be a better match
		if rv3[i].Start >= 1 && unicode.IsSpace(rune(res.Content[rv3[i].Start-1])) {
			m.Score += SpaceBoundBoost
		}
		if rv3[i].End < len(res.Content)-1 && unicode.IsSpace(rune(res.Content[rv3[i].End+1])) {
			m.Score += SpaceBoundBoost
		}

		// If the word is an exact match to what the user typed boost it
		// So while the search may be case insensitive the ranking of
		// the snippet does consider case when boosting ever so slightly
		if string(res.Content[rv3[i].Start:rv3[i].End]) == rv3[i].Word {
			m.Score += ExactMatchBoost
		}

		// This mod applies over the whole score because we want to most unique words to appear in the middle
		// of the snippet over those where it is on the edge which this should achieve even if it means
		// we may miss out on a slightly better match
		m.Score = m.Score / float64(documentFrequencies[rv3[i].Word]) // Factor in how unique the word is
		bestMatches = append(bestMatches, m)
	}

	// Sort our matches by score such that tbe best snippets are at the top
	sort.Slice(bestMatches, func(i, j int) bool {
		return bestMatches[i].Score > bestMatches[j].Score
	})

	// Now what we have it sorted lets get just the ones that don't overlap so we have all the unique snippets
	var bestMatchesClean []bestMatch
	var ranges [][]int
	for _, b := range bestMatches {
		isOverlap := false
		for _, r := range ranges {
			if b.StartPos >= r[0] && b.StartPos <= r[1] {
				isOverlap = true
			}

			if b.EndPos >= r[0] && b.EndPos <= r[1] {
				isOverlap = true
			}
		}

		if !isOverlap {
			ranges = append(ranges, []int{b.StartPos, b.EndPos})
			bestMatchesClean = append(bestMatchesClean, b)
		}
	}

	// Limit to the 20 best matches
	if len(bestMatchesClean) > 20 {
		bestMatchesClean = bestMatchesClean[:20]
	}

	var snippets []Snippet
	for _, b := range bestMatchesClean {

		index := bytes.Index(res.Content, res.Content[b.StartPos:b.EndPos])
		startLineOffset := 1
		for i := 0; i < index; i++ {
			if res.Content[i] == '\n' {
				startLineOffset++
			}
		}

		contentLineOffset := startLineOffset
		for _, i := range res.Content[b.StartPos:b.EndPos] {
			if i == '\n' {
				contentLineOffset++
			}
		}

		snippets = append(snippets, Snippet{
			Content:   string(res.Content[b.StartPos:b.EndPos]),
			StartPos:  b.StartPos,
			EndPos:    b.EndPos,
			Score:     b.Score,
			LineStart: startLineOffset,
			LineEnd:   contentLineOffset,
		})
	}

	return snippets
}

// Get all of the locations into a new data structure
// which makes things easy to sort and deal with
func convertToRelevant(res *common.FileJob) []relevantV3 {
	var rv3 []relevantV3

	for k, v := range res.MatchLocations {
		if len(k) < minTermLen {
			continue
		}
		for _, i := range v {
			rv3 = append(rv3, relevantV3{
				Word:  k,
				Start: i[0],
				End:   i[1],
			})
		}
	}

	// Sort the results so when we slide around everything is in order
	sort.Slice(rv3, func(i, j int) bool {
		return rv3[i].Start < rv3[j].Start
	})

	return rv3
}

// Looks for a nearby whitespace character near this position (`pos`)
// up to `distance` away.  Returns index of space if a space was found and
// true, otherwise returns the original index and false
func findSpaceRight(res *common.FileJob, pos int, distance int) (int, bool) {
	if len(res.Content) == 0 {
		return pos, false
	}

	end := pos + distance
	if end > len(res.Content)-1 {
		end = len(res.Content) - 1
	}

	// Look for spaces
	for i := pos; i <= end; i++ {
		if str.StartOfRune(res.Content[i]) && unicode.IsSpace(rune(res.Content[i])) {
			return i, true
		}
	}

	return pos, false
}

// Looks for nearby whitespace character near this position
// up to distance away. Returns index of space if a space was found and tru
// otherwise the original index is return and false
func findSpaceLeft(res *common.FileJob, pos int, distance int) (int, bool) {
	if len(res.Content) == 0 {
		return pos, false
	}

	if pos >= len(res.Content) {
		return pos, false
	}

	end := pos - distance
	if end < 0 {
		end = 0
	}

	// Look for spaces
	for i := pos; i >= end; i-- {
		if str.StartOfRune(res.Content[i]) && unicode.IsSpace(rune(res.Content[i])) {
			return i, true
		}
	}

	return pos, false
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
