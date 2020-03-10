// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package string

import (
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

// IndexAll extracts all of the locations of a string inside another string
// up-to the defined limit and does so without regular expressions
// which makes it faster than FindAllIndex in most situations while
// not being any slower. It performs worst when working against random
// data.
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
func IndexAll(haystack string, needle string, limit int) [][]int {
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
		limit = math.MaxInt32
	} else {
		// Increment by one because we do count++ at the start of the loop
		// and as such there is a off by 1 error in the return otherwise
		limit++
	}

	var count int
	for loc != -1 {
		count++

		if count == limit {
			break
		}

		// trim off the portion we already searched, and look from there
		searchText = searchText[loc+len(needle):]
		locs = append(locs, []int{loc + offSet, loc + offSet + len(needle)})

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

var __permuteCache = map[string][]string{}

// IndexAllIgnoreCaseUnicode extracts all of the locations of a string inside another string
// up-to the defined limit. It is designed to be faster than uses of FindAllIndex with
// case insenstive matching enabled, by looking for string literals first and then
// checking for exact matches. It also does so in a unicode aware way such that a search
// for S will search for S s and ſ which a simple strings.ToLower over the haystack
// and the needle will not.
func IndexAllIgnoreCaseUnicode(haystack string, needle string, limit int) [][]int {
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
	// such as the one in Rust where given a string literal if you get
	// all the case options of that such as turning foo into foo Foo fOo FOo foO FoO fOO FOO
	// and then use Boyer-Moore or some such for those. Of course using something
	// like Aho-Corasick or Rabin-Karp to get multi match would be a better idea so you
	// can match all of the input in one pass.
	//
	// If the needle is over some amount of characters long you chop off the first few
	// and then search for those. However this means you are not finding actual matches and as such
	// you the need to validate a potential match after you have found one.
	// In this case the confirmation match is done using regular expressions
	// because its faster than checking for all case options for longer needles.

	locs := [][]int{}
	// Char limit is the cut-off where we switch from all case permutations
	// to just the first 3 and then check for an actual match
	// in my tests 3 speeds things up the most against test data
	// of many famous books concatenated together and large
	// amounts of data from /dev/urandom
	var charLimit = 3

	//var searchTerms []string
	if utf8.RuneCountInString(needle) <= charLimit {
		// We are below the limit we set, so get all the search
		// terms and search for that
		searchTerms, ok := __permuteCache[needle]
		if !ok {
			if len(__permuteCache) > 10 {
				__permuteCache = map[string][]string{}
			}
			searchTerms = PermuteCaseFolding(needle)
			__permuteCache[needle] = searchTerms
		}

		// TODO - Investigate
		// This is using IndexAll in a loop which was faster than
		// any implementation of Aho-Corasick or Boyer-Moore I tried
		// but in theory Aho-Corasick / Rabin-Karp or even a modified
		// version of Boyer-Moore should be faster than this
		for _, term := range searchTerms {
			locs = append(locs, IndexAll(haystack, term, limit)...)

			if limit > 0 && len(locs) > limit {
				return locs[:limit]
			}
		}
	} else {
		// Over the character limit so look for potential matches and only then check to find real ones

		// Note that we have to use runes here to avoid cutting bytes off so
		// cast things around to ensure it works
		s := []rune(needle)

		searchTerms, ok := __permuteCache[string(s[:charLimit])]
		if !ok {
			if len(__permuteCache) > 10 {
				__permuteCache = map[string][]string{}
			}
			searchTerms = PermuteCaseFolding(string(s[:charLimit]))
			__permuteCache[string(s[:charLimit])] = searchTerms
		}

		// We create a regular expression which is used for validating the match
		// after we have identified a potential one
		regexIgnore := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(needle))

		// TODO - Investigate
		// This is using IndexAll in a loop which was faster than
		// any implementation of Aho-Corasick or Boyer-Moore I tried
		// but in theory Aho-Corasick / Rabin-Karp or even a modified
		// version of Boyer-Moore should be faster than this
		for _, term := range searchTerms {
			potentialMatches := IndexAll(haystack, term, -1)

			for _, match := range potentialMatches {
				// We have a potential match, so now see if it actually matches
				// by getting the actual value out of our haystack
				if len(haystack) < match[0]+len(needle) {
					continue
				}

				// Because the length of the needle might be different to what we just found as a match
				// based on byte size we add enough extra on the end to deal with the difference
				e := len(needle) + len(needle) - 1
				for match[0]+e > len(haystack) {
					e--
				}
				toMatch := haystack[match[0] : match[0]+e]

				// Use a regular expression to match because we already cut down the time
				// needed and its faster than CaseFolding large needles and then iterating
				// over that list, and for especially long needles it will produce billions
				// of results we need to check.
				// NB have to use findAllHere and not Match because we need to know the
				// length of the match such that we can produce the correct offset.
				i := regexIgnore.FindAllIndex([]byte(toMatch), -1)

				if len(i) != 0 {
					// When we have confirmed a match we add it to our total
					// but adjust the positions to the match and the length of the
					// needle to ensure the byte count lines up
					locs = append(locs, []int{match[0], match[0] + (i[0][1] - i[0][0])})

					if limit > 0 && len(locs) > limit {
						return locs[:limit]
					}
				}
			}
		}
	}

	// Retain compatibility with FindAllIndex method
	if len(locs) == 0 {
		return nil
	}

	return locs
}
