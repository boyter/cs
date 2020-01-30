package snippet

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

// Extracts all of the locations of a string inside another string
// upto the defined limit and does so without regular expressions
// which  makes it about 3x more efficient
func ExtractLocation(word string, fulltext string, limit int) []int {
	locs := []int{}

	searchText := fulltext
	offSet := 0
	loc := strings.Index(searchText, word)

	count := 0
	for loc != -1 {
		count++

		if count == limit {
			break
		}

		searchText = searchText[loc+len(word):]
		locs = append(locs, loc+offSet)

		// trim off the start, and look from there and keep trimming
		offSet += loc + len(word)
		loc = strings.Index(searchText, word)
	}

	sort.Ints(locs)

	return locs
}

// This method is about 3x more efficient then using regex to find
// all the locations assuming you are matching plain strings
//
// BenchmarkExtractLocationsRegex-8     	   50000	     30159 ns/op
// BenchmarkExtractLocationsNoRegex-8   	  200000	     11915 ns/op
func ExtractLocations(words []string, fulltext string) []int {
	locs := []int{}

	fulltext = strings.ToLower(fulltext)

	for _, w := range words {
		for _, l := range ExtractLocation(w, fulltext, 20) {
			locs = append(locs, l)
		}
	}

	sort.Ints(locs)

	// If no words were found then we show the beginning of the text
	// by setting the location to 0
	// Note this should not happen generally as you would only try to snip
	// things for which you know a match is made
	if len(locs) == 0 {
		locs = append(locs, 0)
	}

	return locs
}

// Work out which is the most relevant portion to display
// This is done by looping over each match and finding the smallest distance between two found
// strings. The idea being that the closer the terms are the better match the snippet would be.
// When checking for matches we only change the location if there is a better match.
// The only exception is where we have only two matches in which case we just take the
// first as will be equally distant.
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

// When we use the ExtractRelevant method it is possible that it cut on things that were inserted
// using the WriteHighlights method which might make the output look odd. This method attempts to fix
// the issue by removing partial matches on the edges
func PatchRelevant(relevant, in, out, indicator string) string {
	if in == "" && out == "" {
		return relevant
	}

	//// if the indicator
	//start := len(indicator)
	//
	//for i := range relevant {
	//
	//}

	return relevant
}

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
