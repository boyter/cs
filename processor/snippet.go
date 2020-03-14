package processor

import (
	"fmt"
	"sort"
	"unicode"

	str "github.com/boyter/cs/string"
)

// Defines the maximum bytes either side of the match
// we are willing to return
var SNIP_SIDE_MAX int = 10

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
	Content  string
	StartPos int
	EndPos   int
}

// Looks though the locations using a sliding window style algorithm
// where it "brute forces" the solution by iterating over every location we have
// and look for all matches that fall into the supplied length and ranking
// based on how many we have.
//
// This algorithm ranks using document frequencies that are kept for
// TF/IDF ranking with various other checks. Look though the source
// to see how it actually works as it is a constant work in progress.
// Some examples of what it can produce which I consider good results,
//
// corpus: Jane Austens Pride and Prejudice
// searchtext: ten thousand a year
// result:  before. I hope he will overlook
//	it. Dear, dear Lizzy. A house in town! Every thing that is
//	charming! Three daughters married! Ten thousand a year! Oh, Lord!
//	What will become of me. I shall go distracted.”
//
//	This was enough to prove that her approbation need not be
//
// searchtext: poor nerves
// result:  your own children in such a way?
//	You take delight in vexing me. You have no compassion for my poor
//	nerves.”
//
//	“You mistake me, my dear. I have a high respect for your nerves.
//	They are my old friends. I have heard you mention them with
//	consideration these last
//
// Please note that testing this is... hard. This is because what is considered relevant also happens
// to differ between people. As such this is not tested as much as other methods and you should not
// rely on the results being static over time as the internals will be modified to produce better
// results where possible
func extractRelevantV3(res *fileJob, documentFrequencies map[string]int, relLength int, indicator string) []Snippet {
	wrapLength := relLength / 2
	bestMatches := []bestMatch{}

	rv3 := []relevantV3{}
	// Get all of the locations into a new data structure
	// which makes things easy to sort and deal with
	for k, v := range res.MatchLocations {
		for _, i := range v {
			rv3 = append(rv3, relevantV3{
				Word:  k,
				Start: i[0],
				End:   i[1],
			})
		}
	}

	// If we have a single result and its a filename match which has
	// no real start or end position
	// it means we have no content to look through so just display the first
	// chunk of the file
	if len(rv3) == 1 && rv3[0].Start == 0 && rv3[0].End == 0 {
		endPos := 300
		if len(res.Content) < 300 {
			endPos = len(res.Content)
		}

		// TODO check for if we cut in the middle of a multibyte character
		return []Snippet{
			{
				Content:  string(res.Content[:endPos]),
				StartPos: 0,
				EndPos:   0,
			},
		}
	}

	// Sort the results so when we slide around everything is in order
	sort.Slice(rv3, func(i, j int) bool {
		return rv3[i].Start < rv3[j].Start
	})

	// Slide around looking for matches that fit in the length
	for i := 0; i < len(rv3); i++ {
		m := bestMatch{
			StartPos: rv3[i].Start,
			EndPos:   rv3[i].End,
			Relevant: []relevantV3{rv3[i]},
		}

		// Slide left
		j := i - 1
		//for j:= 1 - 1; j--; j >= 0 {	// Does this replace breaks ?
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

		// TIM to check
		// Now we see if there are any nearby spaces to avoid us cutting in the
		// middle of a word if we can avoid it
		//space, b := findNearbySpace(res, m.StartPos, SNIP_SIDE_MAX)
		space, b := findSpaceRight(res, m.StartPos, SNIP_SIDE_MAX)
		if b {
			m.StartPos = space
		}
		space, b = findSpaceRight(res, m.EndPos, SNIP_SIDE_MAX)
		if b {
			m.EndPos = space
		}

		// If we are very close to the start, just push it out so we get the actual start
		if m.StartPos <= SNIP_SIDE_MAX {
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
		//m.Score += float64(m.EndPos - m.StartPos) // Factor in how large the snippet is NB weight this as it makes things worse sometimes

		// Apply higher score where the words are near each other
		mid := rv3[i].Start + (rv3[i].End-rv3[i].End)/2 // match word midpoint
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
			// Use 2 here because we want to avoid punctuation
			if abs(rv3[i].Start-v.End) <= 2 || abs(rv3[i].End-v.Start) <= 2 {
				m.Score += 20
			}
		}

		// If the match is bounded by a space boost it slightly
		// because its likely to be a better match
		if rv3[i].Start >= 1 && unicode.IsSpace(rune(res.Content[rv3[i].Start-1])) {
			m.Score += 5
		}
		if rv3[i].End < len(res.Content)-1 && unicode.IsSpace(rune(res.Content[rv3[i].End+1])) {
			m.Score += 5
		}

		// If the word is an exact match to what the user typed boost it
		// such that while we the search may be insensitive the ranking of
		// the snippet does consider case
		if string(res.Content[rv3[i].Start:rv3[i].End]) == rv3[i].Word {
			m.Score += 5
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

	// Limit to the 10 best matches
	if len(bestMatches) > 10 {
		bestMatches = bestMatches[:10]
	}

	snippets := []Snippet{}
	for _, b := range bestMatches {
		snippets = append(snippets, Snippet{
			Content:  string(res.Content[b.StartPos:b.EndPos]),
			StartPos: b.StartPos,
			EndPos:   b.EndPos,
		})
	}

	return snippets
}

// Looks for a nearby whitespace character near this position (`pos`)
// up to `distance` away.  Returns index of space if a space was found
// otherwise the the position is distance to the next complete
// code-point past the `distance`.  In this case flag `false`.
func findNearbySpace(res *fileJob, pos int, distance int) (int, bool) {
	leftDistance := pos - distance
	if leftDistance < 0 {
		leftDistance = 0
	}

	// Avoid possible overflow if we need to check the last byte
	if pos == len(res.Content) {
		pos--
	}

	// look left
	// TODO should get a slice, and iterate the runes in it
	// TODO check but I think this is acceptable for whitespace chars...
	for i := pos; i >= leftDistance; i-- {
		if unicode.IsSpace(rune(res.Content[i])) {
			return i, true
		}
	}

	rightDistance := pos + distance
	if rightDistance >= len(res.Content) {
		rightDistance = len(res.Content)
	}

	// look right
	// TODO should get a slice, and iterate the runes in it
	for i := pos; i < rightDistance; i++ {
		if unicode.IsSpace(rune(res.Content[i])) {
			return i, true
		}
	}

	return pos, false
}

//func findNearbySpace(res *fileJob, pos int, distance int) (int, bool) {
// TODO: call search left and search right
//}

func findSpaceLeft(res *fileJob, pos int, distance int) (idx int, found bool) {
	// Deal with misuse
	if len(res.Content) < 1 {
		return 0, false
	}

	// Avoid overflow from invalid pos
	if (pos >= len(res.Content)) || (pos < 0) {
		pos = len(res.Content) - 1
	}

	// Avoid overflows from invalid distance
	if distance > pos {
		distance = pos
	}

	// Set default return values
	idx, found = pos, false

	// Look for spaces
	for i := 0; i <= distance; i++ { // Does this need to be <= or can it be < ?
		idx = pos - i
		idx2 := idx + 1
		if i == 0 { // We'll have an index error for the 2nd byte
			idx2 = idx
		}
		found = str.IsSpace(res.Content[idx], res.Content[idx2])
		if found {
			return
		}
	}

	// No space found.  Count back up to make sure we don't split mid rune.
	// Only count back to the largest possible index.
	for idx < (len(res.Content)-1) && !str.StartOfRune(res.Content[idx]) {
		idx++
	}

	//fmt.Println("yoyo")
	return
}

func findSpaceRight(res *fileJob, pos int, distance int) (idx int, found bool) {
	fmt.Println("len(res.Content)", len(res.Content))
	// Deal with misuse
	if len(res.Content) < 1 {
		return 0, false
	}

	// Avoid overflow from invalid pos
	if (pos >= len(res.Content)) || (pos < 0) {
		pos = 0
	}

	// Avoid overflows from invalid distance
	if distance > (len(res.Content) - pos) { // Up to here
		distance = len(res.Content) - 1 - pos
	}

	// Set default return values
	idx, found = pos, false

	// Look for spaces // TODO could set i to idx for count up
	for i := 0; i <= distance; i++ {
		idx = pos + i
		idx2 := idx + 1
		if idx2 == len(res.Content) { // ... we'll have an index error.
			// ∴ We need an index that is valid.
			// All that matters here is that we don't accidentally
			// get a match by choosing at random.  No 2 byte
			// 'space' is a repeat of the first byte so putting
			// the same byte in twice won't break anything.
			idx2 = idx
		}
		fmt.Printf("i: %d - res.Content[%d] = '%c', res.Content[%d] = '%c'\n",
			i, idx, res.Content[idx], idx2, res.Content[idx2])
		found = str.IsSpace(res.Content[idx], res.Content[idx2])
		if found {
			return
		}
	}
	fmt.Println("Did not find a space, within given distance", distance)

	// No space found.  Count back up to make sure we don't split mid rune.
	// Only count back to the largest possible index.
	for idx > 0 && !str.StartOfRune(res.Content[idx]) {
		idx--
	}

	return
}
