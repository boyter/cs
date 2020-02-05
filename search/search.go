package search

import (
	"strings"
)

// Extracts all of the locations of a string inside another string
// upto the defined limit and does so without regular expressions
// which  makes it about 3x more efficient than the regex way of doing this.
//
// A sample benchmark result to illustrate the point
//
// BenchmarkExtractLocationsRegex-8     	   50000	     30159 ns/op
// BenchmarkExtractLocationsNoRegex-8   	  200000	     11915 ns/op
//
// Also note that this method has a limit option allowing you to bail out
// at some threshold of matches which is useful in situations where
// additional matches are no longer useful. Otherwise set to to math.MaxInt64
// to get what should hopefully be all possible matches although I suspect
// you may hit memory limits at that point.
//
// Note that this method is explicitly case sensitive in its matching
func ExtractLocations(term string, fulltext string, limit int64) []int {
	locs := []int{}

	searchText := fulltext
	offSet := 0
	loc := strings.Index(searchText, term)

	limit++

	var count int64
	for loc != -1 {
		count++

		if count == limit {
			break
		}

		searchText = searchText[loc+len(term):]
		locs = append(locs, loc+offSet)

		// trim off the start, and look from there and keep trimming
		offSet += loc + len(term)
		loc = strings.Index(searchText, term)
	}

	return locs
}

// This method is about 3x more efficient then using regex to find
// all the locations assuming you are using literal strings.
//
// Also note that this method has a limit option allowing you to bail out
// at some threshold of matches which is useful in situations where
// additional matches are no longer useful. Otherwise set to to math.MaxInt64
// to get what should hopefully be all possible matches although I suspect
// you may hit memory limits at that point.
//
// One subtle thing about this method is that it work
func ExtractTermLocations(terms []string, fulltext string, limit int64) []int {
	locs := []int{}

	for _, w := range terms {
		for _, l := range ExtractLocations(w, fulltext, limit) {
			locs = append(locs, l)
		}
	}

	return locs
}

// Given a string returns a slice containing all possible case permutations
// of that string such that input of foo will return
// foo Foo fOo FOo foO FoO fOO FOO
// Note that very long inputs can produce an enormous amount of
// results in the returned slice
func PermuteCase(input string) []string {
	l := len(input)
	max := 1 << l

	combinations := []string{}

	for i := 0; i < max; i++ {
		s := ""
		for idx, ch := range input {
			if (i & (1 << idx)) == 0 {
				s += strings.ToUpper(string(ch))
			} else {
				s += strings.ToLower(string(ch))
			}
		}

		combinations = append(combinations, s)
	}

	return combinations
}
