package processor

import (
	"math"
	"regexp"
	"sort"
	"strings"
)

// find the locations of each of the words
// Nothing exciting here. The array_unique is required
// unless you decide to make the words unique before passing in
func extractLocations(words []string, fulltext string) []int {
	locs := []int{}

	fulltext = strings.ToLower(fulltext)

	for _, w := range words {
		t := regexp.MustCompile(w).FindAllIndex([]byte(fulltext), -1)

		for _, x := range t {
			locs = append(locs, x[0])
		}
	}

	sort.Ints(locs)

	// If not words found show beginning of the text NB should not happen
	if len(locs) == 0 {
		locs = append(locs, 0)
	}

	return locs
}

func extractLocationsNoRegex(words []string, fulltext string) []int {
	locs := []int{}

	fulltext = strings.ToLower(fulltext)


	for _, w := range words {
		searchText := fulltext
		offSet := 0
		loc := strings.Index(searchText, w)

		for loc != -1 {
			searchText = searchText[loc+len(w):]
			locs = append(locs, loc+offSet)
			// trim off the start, and look from there and keep trimming
			offSet += loc+len(w)
			loc = strings.Index(searchText, w)
		}
	}

	sort.Ints(locs)

	// If not words found show beginning of the text NB should not happen
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
func determineSnipLocations(locations []int, previousCount int) int {
	startPos := locations[0]
	locCount := len(locations)
	smallestDiff := math.MaxInt32

	var diff int
	if locCount > 2 {
		for i := 0; i < locCount; i++ {
			if i == locCount-1 { // at the end
				diff = locations[i] - locations[i-1]
			} else {
				diff = locations[i+1] - locations[i]
			}

			if smallestDiff > diff {
				smallestDiff = diff
				startPos = locations[i]
			}
		}
	}

	if startPos > previousCount {
		startPos = startPos - previousCount
	} else {
		startPos = 0
	}

	return startPos
}

// 1/6 ratio on prevcount tends to work pretty well and puts the terms
// in the middle of the extract
// indicator is usually ellipsis or some such
func extractRelevant(words []string, fulltext string, relLength int, prevCount int, indicator string) string {
	textLength := len(fulltext)

	if textLength <= relLength {
		return fulltext
	}

	locations := extractLocationsNoRegex(words, fulltext)
	startPos := determineSnipLocations(locations, prevCount)

	// if we are going to snip too much...
	if textLength-startPos < relLength {
		startPos = startPos - (textLength-startPos)/2
	}

	endPos := startPos + relLength
	if endPos > textLength {
		endPos = textLength
	}

	relText := fulltext[startPos:endPos]

	if startPos+relLength < textLength {
		relText = relText[0:strings.LastIndex(relText, " ")] + indicator
	}

	if startPos != 0 {
		relText = indicator + relText[strings.Index(relText, " ")+1:]
	}

	return relText
}
