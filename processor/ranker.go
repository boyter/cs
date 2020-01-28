package processor

import (
	"math"
	"sort"
	"strings"
)

func RankResults(searchTerms []string, results []*FileJob) []*FileJob {
	// TODO blend rankers if possible
	results = RankResultsTFIDF(searchTerms, results)
	results = RankResultsTitle(searchTerms, results)
	return results
}

func RankResultsVectorSpace(searchTerms []string, results []*FileJob) []*FileJob {
	return results
}

func RankResultsTitle(searchTerms []string, results []*FileJob) []*FileJob {

	for i := 0; i < len(results); i++ {
		for _, s := range searchTerms {
			if strings.Contains(results[i].Location, s) {
				// Boost the rank slightly
				results[i].Score = results[i].Score * 1.1
			}
		}
	}

	return results
}

// TF-IDF ranking of results
func RankResultsTFIDF(searchTerms []string, results []*FileJob) []*FileJob {
	idf := map[string]int{}
	for _, r := range results {
		for k := range r.Locations {
			idf[k] = idf[k] + len(r.Locations[k])
		}
	}

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

// Sort a slice of filejob results based on their score and then location to stop
// any undeterministic ordering happening
func SortResults(results []*FileJob) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return strings.Compare(results[i].Location, results[j].Location) < 0
		}

		return results[i].Score > results[j].Score
	})
}
