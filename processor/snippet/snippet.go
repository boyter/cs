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

	// If no words found show beginning of the text NB should not happen
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
func determineSnipLocations(locations []LocationType, previousCount int) int {
	startPos := locations[0].Location
	locCount := len(locations)
	smallestDiff := math.MaxInt32

	var diff int
	if locCount > 2 {
		for i := 0; i < locCount; i++ {
			if i == locCount-1 { // at the end
				diff = locations[i].Location - locations[i-1].Location
			} else {
				diff = locations[i+1].Location - locations[i].Location
			}

			if i != locCount-1 {
				// If the term after this one is different reduce the weight so its considered more relevant
				if locations[i].Term != locations[i+1].Term {
					diff = (diff / 2) - len(locations[i].Term) - len(locations[i+1].Term)
				}
			}

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

	return startPos
}

// 1/6 ratio on prevcount tends to work pretty well and puts the terms
// in the middle of the extract
func GetPrevCount(relLength int) int {
	return CalculatePrevCount(relLength, 6)
}

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

func ExtractRelevant(fulltext string, locations []LocationType, relLength int, prevCount int, indicator string) string {
	textLength := len(fulltext)

	if textLength <= relLength {
		return fulltext
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
		indicatorPos := strings.Index(relText, " ")
		indicatorLen := 1

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
