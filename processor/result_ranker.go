// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package processor

import (
	"github.com/boyter/cs/str"
	"math"
	"sort"
	"strings"
)

// Takes in the search terms and results and applies chained
// ranking over them to produce a score and then sort those results
// and return them
// Note that this method will evolve over time
// and as such you should never rely on the returned results being
// the same
func rankResults(corpusCount int, results []*fileJob) []*fileJob {
	documentFrequencies := calculateDocumentFrequency(results)

	results = rankResultsTFIDF(corpusCount, results, documentFrequencies) // needs to come first because it resets the scores
	results = rankResultsPhrase(results, documentFrequencies)
	results = rankResultsLocation(results)
	sortResults(results)
	return results
}

// Base value used to determine how much location matches
// should be boosted by
const (
	LocationBoostValue = 0.05
	PhraseBoostValue   = 1.00
)

// Given the results boost based on how close the phrases are to each other IE make it slightly phrase
// heavy. This is fairly similar to how the snippet extraction works but with less work because it does
// not need to deal with cutting between unicode endpoints
// NB this is one of the more expensive parts of the ranking
func rankResultsPhrase(results []*fileJob, documentFrequencies map[string]int) []*fileJob {
	for i := 0; i < len(results); i++ {
		rv3 := convertToRelevant(results[i])

		for j := 0; j < len(rv3); j++ {
			if j == 0 {
				continue
			}

			// If the word is within a reasonable distance of this word boost the score
			// weighted by how common that word is so that matches like 'a' impact the rank
			// less than something like 'cromulent' which in theory should not occur as much
			if rv3[j].Start-rv3[j-1].End < 5 {
				// Set to 1 which seems to produce reasonable results by only boosting a little per term
				results[i].Score += PhraseBoostValue / float64(documentFrequencies[rv3[j].Word])
			}
		}
	}

	return results
}

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
				if results[i].Score == 0 || math.IsNaN(results[i].Score) {
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
					(LocationBoostValue * float64(len(locs)) * float64(len(key))))

				// If the location is closer to the start boost or rather don't
				// affect negatively as much because we reduce the score slightly based on
				// how far away from the start it is
				low := math.MaxInt32
				for _, l := range locs {
					if l[0] < low {
						low = l[0]
					}
				}

				results[i].Score = results[i].Score*1.0 - (float64(low) * 0.02)
			}
		}

		// If we found multiple terms (assuming we have multiple), boost yet again to
		// reward matches which have multiple matches
		if foundTerms > 1 {
			results[i].Score = results[i].Score * (1 + LocationBoostValue*float64(foundTerms))
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
func rankResultsTFIDF(corpusCount int, results []*fileJob, documentFrequencies map[string]int) []*fileJob {
	// Get the number of docs with each word in it, which is just the number of results because we are AND only
	// and as such each document must contain all the words although they may have different counts
	var weight float64
	for i := 0; i < len(results); i++ {
		weight = 0

		// We don't know how many words are actually in this document... and I don't want to check
		// because its going to slow things down. Keep in mind that this works inside the words themselves
		// I.E. partial matches are the norm so it makes sense to base it on the number of bytes
		// where we assume about 50 "words" per 1000 bytes of text.
		// Also ensure that it is at least 1 to avoid divide by zero errors later on.
		words := float64(maxInt(1, results[i].Bytes/20))

		for key, value := range results[i].MatchLocations {
			// Technically the IDF for this is wrong because we only
			// have the count for the matches of the document not all the terms
			// that are actually required I.E.
			// its likely that a search for "a b" is missing the counts
			// for documents that have a but not b and as such
			// the document frequencies are off with respect to the total
			// corpus... although we could get that if needed since we do calculate it...
			// Anyway this isn't a huge issue in practice because TF/IDF
			// still works for a search of a single term such as a or if multiple terms
			// happen to match every document in the corpus which while unlikely
			// is still something that could happen
			// Its also slightly off due to the fact that we don't know the number of words
			// in the document anyway but it should be close enough

			// TF  = number of this words in this document / words in entire document
			// IDF = number of documents that contain this word

			tf := float64(len(value)) / words
			idf := float64(corpusCount) / float64(documentFrequencies[key])

			weight += tf * math.Log2(idf)
		}

		// Override the score here because we don't want whatever we got originally
		// which is just based on the number of keyword matches... of course this assumes
		// that
		results[i].Score = weight
	}

	return results
}

func calculateDocumentFrequency(results []*fileJob) map[string]int {
	// Calculate the document frequency for all words across all documents
	// that we have to get the term frequency for each allowing us to determine
	// how rare or common a word is across the corpus
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
