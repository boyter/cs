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
	if Ranker == "bm25" {
		results = rankResultsBM25(corpusCount, results, calculateDocumentFrequency(results))
	} else {
		results = rankResultsTFIDF(corpusCount, results, calculateDocumentFrequency(results)) // needs to come first because it resets the scores
	}
	results = rankResultsLocation(results)
	// TODO maybe need to add something here to reward phrases
	sortResults(results)
	return results
}

// Base value used to determine how much location matches
// should be boosted by
const (
	LocationBoostValue = 0.05
	DefaultScoreValue  = 0.01
	PhraseBoostValue   = 1.00
	BytesWordDivisor = 2
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
			l := str.IndexAllIgnoreCaseUnicode(results[i].Location, key, -1)

			// Boost the rank slightly based on number of matches and on
			// how long a match it is as we should reward longer matches
			if len(l) != 0 {
				foundTerms++

				// If the rank is ever 0 than nothing will change, so set it
				// to a small value to at least introduce some ranking here
				if results[i].Score == 0 || math.IsNaN(results[i].Score) {
					results[i].Score = DefaultScoreValue
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
					(LocationBoostValue * float64(len(l)) * float64(len(key))))

				// If the location is closer to the start boost or rather don't
				// affect negatively as much because we reduce the score slightly based on
				// how far away from the start it is
				low := math.MaxInt32
				for _, l := range l {
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
	var weight float64
	for i := 0; i < len(results); i++ {
		weight = 0

		// We don't know how many words are actually in this document... and I don't want to check
		// because its going to slow things down. Keep in mind that this works inside the words themselves
		// I.E. partial matches are the norm so it makes sense to base it on the number of bytes
		// Also ensure that it is at least 1 to avoid divide by zero errors later on.
		words := float64(maxInt(1, results[i].Bytes/BytesWordDivisor))

		// word in the case is the word we are dealing with IE what the user actually searched for
		// and wordCount is the locations of those words allowing us to know the number of words matching
		for word, wordCount := range results[i].MatchLocations {
			// Technically the IDF for this is wrong because we only
			// have the count for the matches of the document not all the terms
			// that are actually required
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

			tf := float64(len(wordCount)) / words
			idf := math.Log10(float64(corpusCount) / float64(documentFrequencies[word]))

			if Ranker == "tfidfl" {
				// Lucene modification to improve results https://opensourceconnections.com/blog/2015/10/16/bm25-the-next-generation-of-lucene-relevation/
				weight += math.Sqrt(tf) * idf * (1/math.Sqrt(words))
			} else {
				weight += tf * idf
			}
		}

		// Override the score here because we don't want whatever we got originally
		// which is just based on the number of keyword matches... of course this assumes
		// that
		results[i].Score = weight
	}

	return results
}

// BM25 implementation which ranks the results
// Technically this is not a real BM25 because we don't
// have counts of terms for documents that don't match
// so the IDF value is not correctly calculated
// https://en.wikipedia.org/wiki/Okapi_BM25
//
// NB loops in here use increment to avoid duffcopy
// https://stackoverflow.com/questions/45786687/runtime-duffcopy-is-called-a-lot
// due to how often it is called by things like the TUI mode
//
//                 IDF * TF * (k1 + 1)
// BM25 = sum ----------------------------
//            TF + k1 * (1 - b + b * D / L)
func rankResultsBM25(corpusCount int, results []*fileJob, documentFrequencies map[string]int) []*fileJob {
	var weight float64

	// Get the average number of words across all documents because we need that in BM25 to calculate correctly
	var averageDocumentWords float64
	for i := 0; i < len(results); i++ {
		averageDocumentWords += float64(maxInt(1, results[i].Bytes/BytesWordDivisor))
	}
	averageDocumentWords = averageDocumentWords/float64(len(results))

	k1 := 1.2
	b := 0.75

	for i := 0; i < len(results); i++ {
		weight = 0

		// We don't know how many words are actually in this document... and I don't want to check
		// because its going to slow things down. Keep in mind that this works inside the words themselves
		// I.E. partial matches are the norm so it makes sense to base it on the number of bytes
		// Also ensure that it is at least 1 to avoid divide by zero errors later on.
		words := float64(maxInt(1, results[i].Bytes/BytesWordDivisor))

		// word in the case is the word we are dealing with IE what the user actually searched for
		// and wordCount is the locations of those words allowing us to know the number of words matching
		for word, wordCount := range results[i].MatchLocations {

			// TF  = number of this words in this document / words in entire document
			// IDF = number of documents that contain this word
			tf := float64(len(wordCount)) / words
			idf := math.Log10(float64(corpusCount) / float64(documentFrequencies[word]))

			step1 := idf * tf * (k1 + 1)
			step2 := tf + k1 * (1 - b + (b * words / averageDocumentWords))

			weight += step1 / step2
		}

		// Override the score here because we don't want whatever we got originally
		// which is just based on the number of keyword matches... of course this assumes
		// that
		results[i].Score = weight
	}

	return results
}

// Calculate the document term frequency for all words across all documents
// letting us know how many times a term appears across the corpus
// This is mostly used for snippet extraction
func calculateDocumentTermFrequency(results []*fileJob) map[string]int {
	documentFrequencies := map[string]int{}
	for i := 0; i < len(results); i++ {
		for k := range results[i].MatchLocations {
			documentFrequencies[k] = documentFrequencies[k] + len(results[i].MatchLocations[k])
		}
	}

	return documentFrequencies
}

// Calculate the document frequency for all words across all documents
// allowing us to know the number of documents for which a term appears
// This is mostly used for TF-IDF calculation
func calculateDocumentFrequency(results []*fileJob) map[string]int {
	documentFrequencies := map[string]int{}
	for i := 0; i < len(results); i++ {
		for k := range results[i].MatchLocations {
			documentFrequencies[k] = documentFrequencies[k] + 1
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
