// SPDX-License-Identifier: MIT

package ranker

import (
	"math"
	"sort"
	"strings"

	str "github.com/boyter/go-string"

	"github.com/boyter/cs/pkg/common"
)

// Base value used to determine how much location matches
// should be boosted by
const (
	LocationBoostValue = 0.05
	DefaultScoreValue  = 0.01
	BytesWordDivisor   = 2
)

// RankResults takes in the search results and applies chained
// ranking over them to produce a score and then sort those results
// and return them.
// The rankerName parameter selects the algorithm: "simple", "bm25", "tfidf2",
// or anything else for classic TF-IDF.
func RankResults(rankerName string, corpusCount int, results []*common.FileJob) []*common.FileJob {
	// needs to come first because it resets the scores
	switch rankerName {
	case "simple":
		// in this case the results are already ranked by the number of matches
	case "bm25":
		results = rankResultsBM25(corpusCount, results, CalculateDocumentFrequency(results))
		results = rankResultsLocation(results)
	case "tfidf2":
		results = rankResultsTFIDF(corpusCount, results, CalculateDocumentFrequency(results), false)
		results = rankResultsLocation(results)
	default:
		results = rankResultsTFIDF(corpusCount, results, CalculateDocumentFrequency(results), true)
		results = rankResultsLocation(results)
	}

	sortResults(results)
	return results
}

// Given the results will boost the rank of them based on matches in the
// file location field.
func rankResultsLocation(results []*common.FileJob) []*common.FileJob {
	for i := 0; i < len(results); i++ {
		foundTerms := 0
		for key := range results[i].MatchLocations {
			l := str.IndexAllIgnoreCase(results[i].Location, key, -1)

			if len(l) != 0 {
				foundTerms++

				if results[i].Score == 0 || math.IsNaN(results[i].Score) {
					results[i].Score = DefaultScoreValue
				}

				results[i].Score = results[i].Score * (1.0 +
					(LocationBoostValue * float64(len(l)) * float64(len(key))))

				low := math.MaxInt32
				for _, l := range l {
					if l[0] < low {
						low = l[0]
					}
				}

				results[i].Score = results[i].Score * (1.0 / (1.0 + float64(low)*0.02))
			}
		}

		if foundTerms > 1 {
			results[i].Score = results[i].Score * (1 + LocationBoostValue*float64(foundTerms))
		}
	}

	return results
}

// TF-IDF implementation which ranks the results
func rankResultsTFIDF(corpusCount int, results []*common.FileJob, documentFrequencies map[string]int, classic bool) []*common.FileJob {
	var weight float64
	for i := 0; i < len(results); i++ {
		weight = 0

		words := float64(maxInt(1, results[i].Bytes/BytesWordDivisor))

		for word, wordCount := range results[i].MatchLocations {
			tf := float64(len(wordCount)) / words
			idf := math.Log10(float64(corpusCount) / float64(documentFrequencies[word]))

			if classic {
				weight += tf * idf
			} else {
				weight += math.Sqrt(tf) * idf * (1 / math.Sqrt(words))
			}
		}

		results[i].Score = weight
	}

	return results
}

// BM25 implementation which ranks the results
//
//	IDF * TF * (k1 + 1)
//
// BM25 = sum ----------------------------
//
//	TF + k1 * (1 - b + b * D / L)
func rankResultsBM25(corpusCount int, results []*common.FileJob, documentFrequencies map[string]int) []*common.FileJob {
	if len(results) == 0 {
		return results
	}

	var weight float64

	var averageDocumentWords float64
	for i := 0; i < len(results); i++ {
		averageDocumentWords += float64(maxInt(1, results[i].Bytes/BytesWordDivisor))
	}
	averageDocumentWords = averageDocumentWords / float64(len(results))

	k1 := 1.2
	b := 0.75

	for i := 0; i < len(results); i++ {
		weight = 0

		words := float64(maxInt(1, results[i].Bytes/BytesWordDivisor))

		for word, wordCount := range results[i].MatchLocations {
			rawCount := float64(len(wordCount))
			idf := math.Log10(1 + float64(corpusCount)/float64(documentFrequencies[word]))

			step1 := idf * rawCount * (k1 + 1)
			step2 := rawCount + k1*(1-b+(b*words/averageDocumentWords))

			weight += step1 / step2
		}

		results[i].Score = weight
	}

	return results
}

// CalculateDocumentTermFrequency calculates the document term frequency for all words
// across all documents, letting us know how many times a term appears across the corpus.
// This is mostly used for snippet extraction.
func CalculateDocumentTermFrequency(results []*common.FileJob) map[string]int {
	documentFrequencies := map[string]int{}
	for i := 0; i < len(results); i++ {
		for k := range results[i].MatchLocations {
			documentFrequencies[k] = documentFrequencies[k] + len(results[i].MatchLocations[k])
		}
	}

	return documentFrequencies
}

// CalculateDocumentFrequency calculates the document frequency for all words
// across all documents, allowing us to know the number of documents for which a term appears.
// This is mostly used for TF-IDF calculation.
func CalculateDocumentFrequency(results []*common.FileJob) map[string]int {
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
func sortResults(results []*common.FileJob) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return strings.Compare(results[i].Location, results[j].Location) < 0
		}

		return results[i].Score > results[j].Score
	})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
