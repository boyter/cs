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

func extractLocation(word string, fulltext string) []int {
	locs := []int{}

	searchText := fulltext
	offSet := 0
	loc := strings.Index(searchText, word)

	for loc != -1 {
		searchText = searchText[loc+len(word):]
		locs = append(locs, loc+offSet)

		// trim off the start, and look from there and keep trimming
		offSet += loc + len(word)
		loc = strings.Index(searchText, word)
	}

	sort.Ints(locs)

	// If not words found show beginning of the text NB should not happen
	if len(locs) == 0 {
		locs = append(locs, 0)
	}

	return locs
}

// This method is about 3x more efficient then using regex
//BenchmarkExtractLocationsRegex-8     	   50000	     30159 ns/op
//BenchmarkExtractLocationsNoRegex-8   	  200000	     11915 ns/op
func extractLocationsNoRegex(words []string, fulltext string) []int {
	locs := []int{}

	fulltext = strings.ToLower(fulltext)

	for _, w := range words {
		for _, l := range extractLocation(w, fulltext) {
			locs = append(locs, l)
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
func extractRelevant(words []string, fulltext string, locations []int, relLength int, prevCount int, indicator string) string {
	textLength := len(fulltext)

	if textLength <= relLength {
		return fulltext
	}

	if len(locations) == 0 {
		locations = extractLocationsNoRegex(words, fulltext)
	}

	startPos := determineSnipLocations(locations, prevCount)

	// if we are going to snip too much...
	if textLength-startPos < relLength {
		startPos = startPos - (textLength-startPos)/2
	}

	endPos := startPos + relLength
	if endPos > textLength {
		endPos = textLength
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

	if startPos != 0 {
		relText = indicator + relText[strings.Index(relText, " ")+1:]
	}

	return relText
}
