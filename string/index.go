package string

import (
	"math"
	"strings"
)

// Extracts all of the locations of a string inside another string
// up-to the defined limit and does so without regular expressions
// which makes it considerably faster.
//
// Some benchmark results to illustrate the point (find more in index_benchmark_test.go)
//
// BenchmarkFindAllIndex-8                                           	 2458844	       480 ns/op
// BenchmarkIndexAll-8                                               	14819680	      79.6 ns/op
//
// Note that this method has a limit option allowing you to bail out
// at some threshold of matches which is useful in situations where
// additional matches are no longer useful. Similar to how FindAllIndex
// works. You can use -1 or math.MaxInt64
// to get what should hopefully be all possible matches although I suspect
// you may hit memory limits at that point.
//
// Note that this method is explicitly case sensitive in its matching
// A return value will be an empty slice if no match TODO correct?
func IndexAll(fulltext string, term string, limit int64) []int {
	locs := []int{}

	searchText := fulltext
	offSet := 0
	loc := strings.Index(searchText, term)

	if limit == -1 {
		// Similar to how regex FindAllString works
		// if we have -1 as the limit just try to get everything
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

		searchText = searchText[loc+len(term):]
		locs = append(locs, loc+offSet)

		// trim off the start, and look from there and keep trimming
		offSet += loc + len(term)
		loc = strings.Index(searchText, term)
	}

	return locs
}

func IndexAllIgnoreCase(term string, fulltext string, limit int64) []int {
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
	// Note if the term is over 5 characters long we want to get the first 5
	// characters which in Go means the first 5 runes as the input.
	// However this means you are not finding actual matches and as such
	// you the need to validate a potential match after you have found one
	var terms []string
	terms = PermuteCaseFolding(term)

	locs := []int{}
	// Now we have all the possible case situations we should search for our
	// potential matches
	for _, term := range terms {
		locs = append(locs, IndexAll(fulltext, term, limit)...)
		// TODO validate potential matches here
	}


	return locs
}
