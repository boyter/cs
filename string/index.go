package string

import (
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

// IndexAll extracts all of the locations of a string inside another string
// up-to the defined limit and does so without regular expressions
// which makes it considerably faster than FindAllIndex.
//
// Some benchmark results to illustrate the point (find more in index_benchmark_test.go)
//
// BenchmarkFindAllIndex-8                         2458844	       480.0 ns/op
// BenchmarkIndexAll-8                            14819680	        79.6 ns/op
//
// For pure literal searches IE no regular expression logic this method
// is a drop in replacement for FindAllIndex.
//
// Similar to how FindAllIndex the limit option can be passed -1
// to get all matches.
//
// Note that this method is explicitly case sensitive in its matching.
// A return value of nil indicates no match.
func IndexAll(haystack string, needle string, limit int64) [][]int {
	// Return contains a slice of slices where index 0 is the location of the match in bytes
	// and index 1 contains the end location in bytes of the match
	locs := [][]int{}

	// Perform the first search outside the main loop to make the method
	// easier to understand
	searchText := haystack
	offSet := 0
	loc := strings.Index(searchText, needle)

	if limit <= -1 {
		// Similar to how regex FindAllString works
		// if we have -1 as the limit set to max to
		//  try to get everything
		limit = math.MaxInt64
	} else {
		// Increment by one because we do count++ at the start of the loop
		// and as such there is a off by 1 error in the return otherwise
		limit++
	}

	var count int64
	for loc != -1 {
		count++

		if count == limit {
			break
		}

		// trim off the portion we already searched, and look from there
		searchText = searchText[loc+len(needle):]
		locs = append(locs, []int{loc + offSet, loc + len(needle)})

		// We need to keep the offset of the match so we continue searching
		offSet += loc + len(needle)

		// strings.Index does checks of if the string is empty so we don't need
		// to explicitly do it ourselves
		loc = strings.Index(searchText, needle)
	}

	// Retain compatibility with FindAllIndex method
	if len(locs) == 0 {
		return nil
	}

	return locs
}

func IndexAllIgnoreCaseUnicode(haystack string, needle string, limit int64) [][]int {
	// One of the problems with finding locations ignoring case is that
	// the different case representations can have different byte counts
	// which means the locations using strings or bytes Index can be off
	// if you apply strings.ToLower to your haystack then use strings.Index.
	//
	// This can be overcome using regular expressions but suffers the penalty
	// of hitting the regex engine and paying the price of case
	// insensitive match there.
	//
	// This method tries something else which is used by some regex engines
	// such as the one in rust where given a string literal if you get
	// all the case options of that such as turning foo into foo Foo fOo FOo foO FoO fOO FOO
	// and then searching over those.
	//
	// Note if the needle is over 2 characters long we want to get the first 2
	// characters which in Go means the first 2 runes as the input.
	// However this means you are not finding actual matches and as such
	// you the need to validate a potential match after you have found one

	// TODO If the needle is hilariously long it probably makes sense to fall back into regex

	locs := [][]int{}
	var charLimit = 3

	var searchTerms []string
	if utf8.RuneCountInString(needle) <= charLimit {
		searchTerms = PermuteCaseFolding(needle)

		for _, term := range searchTerms {
			locs = append(locs, IndexAll(haystack, term, limit)...)
		}
	} else {
		// Look for potential matchs and only then find real ones
		s := []rune(needle)
		searchTerms = PermuteCaseFolding(string(s[:charLimit]))
		regexIgnore := regexp.MustCompile(`(?i)` + needle)

		for _, term := range searchTerms {
			potentialMatches := IndexAll(haystack, term, limit)

			for _, match := range potentialMatches {
				// We have a potential match, so now see if it actually matches
				toMatch := haystack[match[0] : match[0]+len(needle)]

				// Use a regular expression here to match because we already cut down the time
				if regexIgnore.Match([]byte(toMatch)) {
					locs = append(locs, []int{match[0], match[0] + len(needle)})
				}
			}
		}
	}

	if len(locs) == 0 {
		return nil
	}

	return locs
}
