package snippet

import (
	"math"
	"sort"
	"strings"
)

// Extracts all of the locations of a string inside another string
// upto the defined limit
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

	// If not words found show beginning of the text NB should not happen
	if len(locs) == 0 {
		locs = append(locs, 0)
	}

	return locs
}

// This method is about 3x more efficient then using regex
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
func GetPrevCount(relLength int) int {
	t := int(relLength / 6)

	if t <= 0 {
		t = 50
	}

	return t
}

func ExtractRelevant(words []string, fulltext string, locations []int, relLength int, prevCount int, indicator string) string {
	textLength := len(fulltext)

	if textLength <= relLength {
		return fulltext
	}

	if len(locations) == 0 {
		locations = ExtractLocations(words, fulltext)
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

	if startPos != 0 {
		relText = indicator + relText[strings.Index(relText, " ")+1:]
	}

	return relText
}
