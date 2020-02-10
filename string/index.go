package string

import (
	"math"
	"strings"
)

// Extracts all of the locations of a string inside another string
// up-to the defined limit and does so without regular expressions
// which makes it considerably faster.
//
// Some benchmark results to illustrate the point (find in index_benchmark_test.go)
//
// BenchmarkFindAllIndex-8                                           	 2458844	       480 ns/op
// BenchmarkIndexAll-8                                               	14819680	      79.6 ns/op
// BenchmarkFindAllIndexLarge-8                                      	 1415024	       767 ns/op
// BenchmarkIndexAllLarge-8                                          	 4332188	       273 ns/op
// BenchmarkFindAllIndexUnicode-8                                    	 2614605	       453 ns/op
// BenchmarkIndexAllUnicode-8                                        	11995201	      98.0 ns/op
// BenchmarkFindAllIndexUnicodeLarge-8                               	  995239	      1362 ns/op
// BenchmarkIndexAllUnicodeLarge-8                                   	 2327736	       508 ns/op
// BenchmarkFindAllIndexManyPartialMatches-8                         	  921036	      1365 ns/op
// BenchmarkIndexAllManyPartialMatches-8                             	 1237137	       959 ns/op
// BenchmarkFindAllIndexUnicodeManyPartialMatches-8                  	 1564449	       763 ns/op
// BenchmarkIndexAllUnicodeManyPartialMatches-8                      	 3305750	       367 ns/op
// BenchmarkFindAllIndexUnicodeManyPartialMatchesVeryLarge-8         	      27	  40394119 ns/op
// BenchmarkIndexAllUnicodeManyPartialMatchesVeryLarge-8             	    2802	    430293 ns/op
// BenchmarkFindAllIndexUnicodeManyPartialMatchesSuperLarge-8        	       1	1568026700 ns/op
// BenchmarkIndexAllUnicodeManyPartialMatchesSuperLarge-8            	      12	 100250200 ns/op
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
func IndexesAll(fulltext string, terms []string, limit int64) []int {
	locs := []int{}

	for _, w := range terms {
		locs = append(locs, IndexAll(fulltext, w, limit)...)
	}

	return locs
}

func FindLocationsCase(term string, fulltext string, limit int64) []int {
	return IndexAll(fulltext, term, limit)
}

func IndexesAllIgnoreCase(term string, fulltext string, limit int64) []int {
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
	terms = PermuteCase(term)

	// Now we have all the possible case situations we should search for our
	// potential matches
	IndexesAll(fulltext, terms, limit)

	return IndexAll(fulltext, term, limit)
}
