package processor

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

type snipLocation struct {
	Location  int
	DiffScore int
}

// Work out which is the most relevant portion to display
// This is done by looping over each match and finding the smallest distance between two found
// strings. The idea being that the closer the terms are the better match the snippet would be.
// When checking for matches we only change the location if there is a better match.
// The only exception is where we have only two matches in which case we just take the
// first as will be equally distant.
//
// Most algorithms such as those explored by A. Turpin, Y. Tsegay, D. Hawking and H. E. Williams
// Fast Generation of Result Snippets in Web Search tend to work using sentences.
//
// This is designed to work with source code to be fast.
func determineSnipLocations(locations [][]int, previousCount int) (int, []snipLocation) {

	// Need to assume we need to sort the locations as there may be multiple calls to
	// FindAll in the same slice
	sort.Slice(locations, func(i, j int) bool {
		// If equal then sort based on how long a match they are
		if locations[i][0] == locations[j][0] {
			return locations[i][1] > locations[j][1]
		}

		return locations[i][0] > locations[j][0]
	})

	startPos := locations[0][0]
	locCount := len(locations)
	smallestDiff := math.MaxInt32
	snipLocations := []snipLocation{}

	var diff int
	if locCount > 2 {
		// We don't need to iterate the last value in this so chop off the last one
		// however note that we access the element anyway as that's how the inner loop works
		for i := 0; i < locCount-1; i++ {

			// We don't need to worry about the +1 out of bounds here
			// because we never loop the last term as the result
			// should be 100% the same as the previous iteration
			diff = locations[i+1][0] - locations[i][0]

			if i != locCount-1 {
				// If the term after this one is different size reduce the diff so it is considered more relevant
				// this is to make terms like "a" all next to each other worth less than "a" next to "apple"
				// so we should in theory choose that section of text to be the most relevant consider
				// this a boost based on diversity of the text we are snipping
				// NB this would be better if it had the actual term
				if locations[i][1]-locations[i][0] != locations[i+1][1]-locations[i+1][0] {
					diff = (diff / 2) - (locations[i][1] - locations[i][0]) - (locations[i+1][1] - locations[i+1][0])
				}
			}

			snipLocations = append(snipLocations, snipLocation{
				Location:  locations[i][0],
				DiffScore: diff,
			})

			// If the terms are closer together and the previous set then
			// change this to be the location we want the snippet to be made from as its likely
			// to have better context as the terms should appear together or at least close
			if smallestDiff > diff {
				smallestDiff = diff
				startPos = locations[i][0]
			}
		}
	}

	if startPos > previousCount {
		startPos = startPos - previousCount
	} else {
		startPos = 0
	}

	// Sort the snip locations based firstly on their diffscore IE
	// which one we think is the best match
	sort.Slice(snipLocations, func(i, j int) bool {
		if snipLocations[i].DiffScore == snipLocations[j].DiffScore {
			return snipLocations[i].Location > snipLocations[j].Location
		}

		return snipLocations[i].DiffScore > snipLocations[j].DiffScore
	})

	return startPos, snipLocations
}

// A 1/6 ratio on tends to work pretty well and puts the terms
// in the middle of the extract hence this method is the default to
// use
func getPrevCount(relLength int) int {
	return calculatePrevCount(relLength, 6)
}

// This attempts to work out given the length of the text we want to display
// how much before we should cut. This is so we can land the highlighted text
// in the middle of the display rather than as the first part so there is
// context
func calculatePrevCount(relLength int, divisor int) int {
	if divisor <= 0 {
		divisor = 6
	}

	t := relLength / divisor

	if t <= 0 {
		t = 50
	}

	return t
}

// This is very loosely based on how snippet extraction works in http://www.sphider.eu/ which
// you can read about https://boyter.org/2013/04/building-a-search-result-extract-generator-in-php/ and
// http://stackoverflow.com/questions/1436582/how-to-generate-excerpt-with-most-searched-words-in-php
// This isn't a direct port because some functionality does not port directly but is a fairly faithful
// port
func extractRelevantV1(fulltext string, locations [][]int, relLength int, indicator string) Snippet {
	textLength := len(fulltext)

	if textLength <= relLength {
		return Snippet{
			Content:  fulltext,
			StartPos: 0,
			EndPos:   textLength,
		}
	}

	startPos, _ := determineSnipLocations(locations, getPrevCount(relLength))

	// If we are about to snip beyond the locations then dial it back
	// do we don't get a slice exception
	if textLength-startPos < relLength {
		startPos = startPos - (textLength-startPos)/2
	}

	endPos := startPos + relLength
	if endPos > textLength {
		endPos = textLength
	}

	if startPos >= endPos {
		startPos = endPos - relLength
	}

	if startPos < 0 {
		startPos = 0
	}

	relText := subString(fulltext, startPos, endPos)

	if startPos+relLength < textLength {
		t := strings.LastIndex(relText, " ")
		if t != -1 {
			relText = relText[0:]
		}
		relText += indicator
	}

	// If we didn't trim from the start its possible we trimmed in the middle of a word
	// this attempts to find a space to make that a cleaner break so we don't cut
	// in the middle of the word, even with the indicator as its not a good look
	if startPos != 0 {

		// Find the location of the first space
		indicatorPos := strings.Index(relText, " ")
		indicatorLen := 1

		// Its possible there was no close to the start, or that perhaps
		// its broken up by newlines or tabs, in which case we want to change
		// the cut to that position
		for _, c := range []string{"\n", "\r\n", "\t"} {
			tmp := strings.Index(relText, c)

			if tmp < indicatorPos {
				indicatorPos = tmp
				indicatorLen = len(c)
			}
		}

		relText = indicator + relText[indicatorPos+indicatorLen:]
	}

	return Snippet{
		Content:  relText,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

type bestMatch struct {
	StartPos int
	EndPos   int
	Score    float64
	Relevant []relevantV3
}

// Looks though the locations using a sliding window style algorithm
// where brute force the solution by iterating over every location we have
// and look for all matches that fall into the supplied length and ranking
// based on how many we have.
//
// Note that this does not have information about what the locations contain
// and as such is probably not the best algorithm, but should run in constant
// time which might be beneficial for VERY large amount of matches.
func extractRelevantV2(fulltext string, locations [][]int, relLength int, indicator string) Snippet {
	sort.Slice(locations, func(i, j int) bool {
		return locations[i][0] < locations[j][0]
	})

	wrapLength := relLength / 2
	bestMatches := []bestMatch{}

	// Slide around looking for matches that fit in the length
	for i := 0; i < len(locations); i++ {
		m := bestMatch{
			StartPos: locations[i][0],
			EndPos:   locations[i][1],
			Score:    1,
		}

		// Slide left
		j := i - 1
		for {
			if j < 0 {
				break
			}

			diff := locations[j][0] - locations[i][0]
			if diff > wrapLength {
				break
			}

			m.StartPos = locations[j][0]
			m.Score++
			j--
		}

		// Slide right
		j = i + 1
		for {
			if j >= len(locations) {
				break
			}

			diff := locations[j][1] - locations[i][0]
			if diff > wrapLength {
				break
			}

			m.EndPos = locations[j][1]
			m.Score++
			j++
		}

		bestMatches = append(bestMatches, m)
	}

	// Sort our matches by score
	sort.Slice(bestMatches, func(i, j int) bool {
		return bestMatches[i].Score > bestMatches[j].Score
	})

	startPos := bestMatches[0].StartPos
	endPos := bestMatches[0].EndPos

	if endPos-startPos < relLength {
		startPos -= relLength / 2
		endPos += relLength / 2

		if startPos < 0 {
			startPos = 0
		}

		if endPos > len(fulltext) {
			endPos = len(fulltext)
		}
	}

	return Snippet{
		Content:  indicator + subString(fulltext, startPos, endPos) + indicator,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

type relevantV3 struct {
	Word  string
	Start int
	End   int
}

// Looks though the locations using a sliding window style algorithm
// where brute force the solution by iterating over every location we have
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
//      it. Dear, dear Lizzy. A house in town! Every thing that is
//      charming! Three daughters married! Ten thousand a year! Oh, Lord!
//      What will become of me. I shall go distracted.”
//
//      This was enough to prove that her approbation need not be
//
// searchtext: poor nerves
// result:  your own children in such a way?
//      You take delight in vexing me. You have no compassion for my poor
//      nerves.”
//
//      “You mistake me, my dear. I have a high respect for your nerves.
//      They are my old friends. I have heard you mention them with
//      consideration these last
//
// Please note that testing this is... hard. This is because what is considered relevant also happens
// to differ between people. As such this is not tested as much as other methods and you should not
// rely on the results being static over time as the internals will be modified to produce better
// results where possible
func extractRelevantV3(res *fileJob, documentFrequencies map[string]int, relLength int, indicator string) Snippet {
	wrapLength := relLength / 2
	bestMatches := []bestMatch{}

	rv3 := []relevantV3{}
	// get all of the locations into a new data structure
	for k, v := range res.MatchLocations {
		for _, i := range v {
			rv3 = append(rv3, relevantV3{
				Word:  k,
				Start: i[0],
				End:   i[1],
			})
		}
	}

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
		space, b := findNearbySpace(res, m.StartPos, 10)
		if b {
			m.StartPos = space
		}
		space, b = findNearbySpace(res, m.EndPos, 10)
		if b {
			m.EndPos = space
		}

		// Now that we have the snippet start to rank it to produce a score indicating
		// how good a match it is and hopefully display to the user what they
		// were actually looking for
		m.Score += float64(len(m.Relevant))     // Factor in how many matches we have
		m.Score += float64(m.EndPos - m.EndPos) // Factor in how large the snippet is

		// Apply higher score where the words are near each other
		mid := rv3[i].Start + (rv3[i].End-rv3[i].End)/2 // match word midpoint
		for _, v := range m.Relevant {
			p := v.Start + (v.End-v.Start)/2 // comparison word midpoint

			// If the word is within a reasonable distance of this word boost the score
			// weighted by how common that word is so that matches like a impact the rank
			// less than something like cromulent
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

	startPos := bestMatches[0].StartPos
	endPos := bestMatches[0].EndPos

	return Snippet{
		Content:  string(res.Content[startPos:endPos]),
		StartPos: startPos,
		EndPos:   endPos,
	}
}

// Looks for a nearby whitespace character near this position
// and return the original position otherwise with a flag
// indicating if a space was actually found
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
	for i := pos; i < rightDistance; i++ {
		if unicode.IsSpace(rune(res.Content[i])) {
			return i, true
		}
	}

	return pos, false
}

type Snippet struct {
	Content  string
	StartPos int
	EndPos   int
}


// Gets a substring of a string rune aware without allocating additional memory at the expense
// of some additional CPU for a loop over the top which is probably worth it.
// Literally copy/pasted from below link
// https://stackoverflow.com/questions/28718682/how-to-get-a-substring-from-a-string-of-runes-in-golang
// TODO pretty sure this messes with the cuts but need to check
func subString(s string, start int, end int) string {
	start_str_idx := 0
	i := 0
	for j := range s {
		if i == start {
			start_str_idx = j
		}
		if i == end {
			return s[start_str_idx:j]
		}
		i++
	}
	return s[start_str_idx:]
}
