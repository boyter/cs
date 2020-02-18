package string

import (
	"math"
	"sort"
	"strings"
)


type LocationType struct {
	Term     string
	Location int
}

type SnipLocation struct {
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
//
// NB makes the assumption that the locations are already sorted
func determineSnipLocations(locations []LocationType, previousCount int) (int, []SnipLocation) {
	startPos := locations[0].Location
	locCount := len(locations)
	smallestDiff := math.MaxInt32
	snipLocations := []SnipLocation{}

	var diff int
	if locCount > 2 {
		// We don't need to iterate the last value in this so chop off the last one
		// however note that we access the element anyway as that's how the inner loop works
		for i := 0; i < locCount-1; i++ {

			// We don't need to worry about the +1 out of bounds here
			// because we never loop the last term as the result
			// should be 100% the same as the previous iteration
			diff = locations[i+1].Location - locations[i].Location

			if i != locCount-1 {
				// If the term after this one is different reduce the weight so its considered more relevant
				// this is to make terms like "a" all next to each other worth less than "a" next to "apple"
				// so we should in theory choose that section of text to be the most relevant consider
				// this a boost based on diversity of the text we are snipping
				if locations[i].Term != locations[i+1].Term {
					diff = (diff / 2) - len(locations[i].Term) - len(locations[i+1].Term)
				}
			}

			snipLocations = append(snipLocations, SnipLocation{
				Location:  locations[i].Location,
				DiffScore: diff,
			})

			// If the terms are closer together and the previous set then
			// change this to be the location we want the snippet to be made from as its likely
			// to have better context as the terms should appear together or at least close
			if smallestDiff > diff {
				smallestDiff = diff
				startPos = locations[i].Location
			}
		}
	}

	if startPos > previousCount {
		startPos = startPos - previousCount
	} else {
		startPos = 0
	}

	// Sort the snip locations based firstly on their diffscore IE
	// which one we think is the best match and
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
func GetPrevCount(relLength int) int {
	return CalculatePrevCount(relLength, 6)
}

// This attempts to work out given the length of the text we want to display
// how much before we should cut. This is so we can land the highlighted text
// in the middle of the display rather than as the first part so there is
// context
func CalculatePrevCount(relLength int, divisor int) int {
	if divisor <= 0 {
		divisor = 6
	}

	t := relLength / divisor

	if t <= 0 {
		t = 50
	}

	return t
}

// Extracts out a relevant portion of text based on the supplied locations and text length
func ExtractRelevant(fulltext string, locations []LocationType, relLength int, prevCount int, indicator string) string {
	textLength := len(fulltext)

	if textLength <= relLength {
		return fulltext
	}

	startPos, _ := determineSnipLocations(locations, prevCount)

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

	relText := fulltext[startPos:endPos]

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

	return relText
}

