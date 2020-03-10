// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package processor

import (
	str "github.com/boyter/cs/string"
	"math"
	"sort"
	"strings"
)

// Takes in the search terms and results and applies chained
// ranking over them to produce a score
// Note that this method will evolve over time
// and as such you should never rely on the returned results being
// the same
func rankResults(corpusCount int, results []*fileJob) []*fileJob {
	results = rankResultsTFIDF(corpusCount, results)
	results = rankResultsLocation(results)
	sortResults(results)
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
func rankResultsLocation(results []*fileJob) []*fileJob {
	for i := 0; i < len(results); i++ {
		foundTerms := 0
		for key := range results[i].MatchLocations {
			locs := str.IndexAllIgnoreCaseUnicode(results[i].Location, key, -1)

			// Boost the rank slightly based on number of matches and on
			// how long a match it is as we should reward longer matches
			if len(locs) != 0 {
				foundTerms++

				// If the rank is ever 0 than nothing will change, so set it
				// to a small value to at least introduce some ranking here
				if results[i].Score == 0 {
					results[i].Score = 0.1
				}

				// Set the score to be itself * 1.something where something
				// is 0.05 times the number of matches * the length of the match
				// so if the user searches for test a file in the location
				// /test/test.go
				// will be boosted and have a higher rank than
				// /test/other.go
				//
				// Of course this assumes that they have the text test in the
				// content otherwise the match is discarded
				results[i].Score = results[i].Score * (1.0 +
					(LocationBoostValue2 * float64(len(locs)) * float64(len(key))))
			}
		}

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
//
// NB loops in here use increment to avoid duffcopy
// https://stackoverflow.com/questions/45786687/runtime-duffcopy-is-called-a-lot
// due to how often it is called by things like the TUI mode
func rankResultsTFIDF(corpusCount int, results []*fileJob) []*fileJob {
	documentFrequencies := calculateDocumentFrequency(results)

	// Get the number of docs with each word in it, which is just the number of results because we are AND only
	// and as such each document must contain all the words although they may have different counts
	var weight float64
	for i := 0; i < len(results); i++ {
		weight = 0

		for key, value := range results[i].MatchLocations {
			// Technically the IDF for this is wrong because we only
			// have the count for the matches of the document not all the terms
			// that are actually required IE
			// its likely that a search for "a b" is missing the counts
			// for documents that have a but not b and as such
			// the document frequencies are off with respect to the total
			// corpus... although we could get that if needed since we do calculate it... TODO investigate
			// Anyway this isn't a huge issue in practice because TF/IDF
			// still works for a search of a single term such as a or if multiple terms
			// happen to match every document in the corpus which while unlikely
			// is still something that could happen

			// TF  = number of this words in this document
			// IDF = number of documents that contain this word
			// N   = total number of documents

			tf := float64(len(value))
			idf := float64(documentFrequencies[key])
			n := corpusCount

			weight += tf * math.Log(float64(n)/idf)
		}

		// For filename matches we have potentially no tf so apply a simple weight that
		// allows the location boost to do its thing
		if math.IsNaN(weight) {
			weight = 1
		}

		results[i].Score = weight
	}

	return results
}

// TODO add test for this
func calculateDocumentFrequency(results []*fileJob) map[string]int {
	// Calculate the document frequency for all words
	documentFrequencies := map[string]int{}
	for i := 0; i < len(results); i++ {
		for k := range results[i].MatchLocations {
			documentFrequencies[k] = documentFrequencies[k] + len(results[i].MatchLocations[k])
		}
	}

	return documentFrequencies
}

// Sort a slice of filejob results based on their score for displaying
// and then sort based on location to stop any undeterministic ordering happening
// as since the location includes the filename we should never have two matches
// that are 100% equal based on the two criteria we use.
func sortResults(results []*fileJob) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return strings.Compare(results[i].Location, results[j].Location) < 0
		}

		return results[i].Score > results[j].Score
	})
}
