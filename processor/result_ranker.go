// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package processor

import (
	"math"
	"sort"
	"strings"
)

// Takes in the search terms and results and applies chained
// ranking over them to produce a score
// Note that this method will evolve over time
// and as such you should never rely on the returned results being
// the same
func rankResults2(results []*fileJob) []*fileJob {
	results = rankResultsTFIDF2(results)
	results = rankResultsLocation2(results)
	sortResults2(results)
	return results
}

// Base value used to determine how much location matches
// should be boosted by
const (
	LocationBoostValue2 = 0.05
)

// Given the results will boost the rank of them based on matches in the
// file location field.
// This is not using TF-IDF or any fancy algorithm just basic checks
// and boosts
func rankResultsLocation2(results []*fileJob) []*fileJob {
	for i := 0; i < len(results); i++ {
		//loc := bytes.ToLower([]byte(results[i].Location))
		foundTerms := 0
		//for _, s := range searchTerms {
		//	t := snippet.ExtractLocations(searchTerms, loc)
		//
		//	// Boost the rank slightly based on number of matches and on
		//	// how long a match it is as we should reward longer matches
		//	if len(t) != 0 && t[0] != 0 {
		//		foundTerms++
		//
		//		// If the rank is ever 0 than nothing will change, so set it
		//		// to a small value to at least introduce some ranking here
		//		if results[i].Score == 0 {
		//			results[i].Score = 0.1
		//		}
		//
		//		// Set the score to be itself * 1.something where something
		//		// is 0.05 times the number of matches * the length of the match
		//		// so if the user searches for test a file in the location
		//		// /test/test.go
		//		// will be boosted and have a higher rank than
		//		// /test/other.go
		//		//
		//		// Of course this assumes that they have the text test in the
		//		// content otherwise the match is discarded
		//		results[i].Score = results[i].Score * (1.0 +
		//			(LocationBoostValue2 * float64(len(t)) * float64(len(s))))
		//	}
		//}

		// If we found multiple terms (assuming we have multiple), boost yet again to
		// reward matches which have multiple matches
		if foundTerms > 1 {
			results[i].Score = results[i].Score * (1 + LocationBoostValue2*float64(foundTerms))
		}
	}

	return results
}

// TF-IDF implementation which ranks the results
// Technically this is not a real TF-IDF because we don't
// have counts of terms for documents that don't match
// so the IDF value is not correctly calculated
// https://en.wikipedia.org/wiki/Tf-idf
func rankResultsTFIDF2(results []*fileJob) []*fileJob {
	idf := map[string]int{}
	for _, r := range results {
		for k := range r.Locations {
			idf[k] = idf[k] + len(r.Locations[k])
		}
	}

	// Increment for loop to avoid duffcopy
	// https://stackoverflow.com/questions/45786687/runtime-duffcopy-is-called-a-lot
	for i := 0; i < len(results); i++ {
		var weight float64
		for k := range results[i].Locations {
			tf := float64(len(results)) / float64(len(results[i].Locations[k]))
			weight += math.Log(1+tf) * float64(len(results[i].Locations[k]))
		}

		results[i].Score = weight
	}

	return results
}

// Sort a slice of filejob results based on their score for displaying
// and then sort based on location to stop any undeterministic ordering happening
// as since the location includes the filename we should never have two matches
// that are 100% equal based on the two criteria we use.
func sortResults2(results []*fileJob) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return strings.Compare(results[i].Location, results[j].Location) < 0
		}

		return results[i].Score > results[j].Score
	})
}
