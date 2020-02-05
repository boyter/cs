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
func FindLocations(term string, fulltext string, limit int64) []int {
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
func FindTermLocations(terms []string, fulltext string, limit int64) []int {
	locs := []int{}

	for _, w := range terms {
		for _, l := range FindLocations(w, fulltext, limit) {
			locs = append(locs, l)
		}
	}

	return locs
}

func FindLocationsCase(term string, fulltext string, limit int64) []int {
	return FindLocations(term, fulltext, limit)
}

func FindLocationsIgnoreCase(term string, fulltext string, limit int64) []int {
	// One of the problems with finding locations ignoring case is that
	// the different case representations can have different byte counts
	// which means the locations using strings or bytes Index can be off
	// if you blindly Lower everything then use Index.
	// This is easy to overcome using regex but suffers the penalty
	// of hitting the regex engine and then paying the price of case
	// insensitive match there.
	// This method tries something else which is used by some regex engines
	// such as the one in rust where given a string literal if you get
	// all the case options of that E.G turn foo into foo Foo fOo FOo foO FoO fOO FOO
	// and then search over those you can find potential matches very quickly
	// and then only on finding a potential match look for an actual one.

	// Note if the term is over 4 characters long we want to get the first 4
	// characters which means in reality the first 4 runes as the input
	// may not be ASCII although this also has issues which can be overcome
	// by https://github.com/rivo/uniseg
	var terms []string
	if len(term) > 4 {
		terms = PermuteCase(term[:4])
	} else {
		terms = PermuteCase(term)
	}
	// Now we have all the possible case situations we should search for our
	// potential matches
	FindTermLocations(terms, fulltext, limit)

	return FindLocations(term, fulltext, limit)
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
