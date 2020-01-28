package processor

import (
	"github.com/boyter/cs/processor/snippet"
	"math"
	"sort"
	"strings"
)

func RankResults(searchTerms []string, results []*FileJob) []*FileJob {
	// TODO blend rankers if possible
	results = RankResultsTFIDF(searchTerms, results)
	results = RankResultsLocation(searchTerms, results)
	return results
}

func RankResultsVectorSpace(searchTerms []string, results []*FileJob) []*FileJob {
	return results
}

func RankResultsLocation(searchTerms []string, results []*FileJob) []*FileJob {
	for i := 0; i < len(results); i++ {
		loc := strings.ToLower(results[i].Location)
		foundTerms := 0
		for _, s := range searchTerms {
			t := snippet.ExtractLocations([]string{strings.ToLower(s)}, loc)

			// Boost the rank slightly based on number of matches and on
			// how long a match it is as we should reward longer matches
			if len(t) != 0 && t[0] != 0 {
				foundTerms++

				// If the rank is ever 0 than nothing will change, so set it
				// to a small value to at least introduce some ranking here
				if results[i].Score == 0 {
					results[i].Score = 0.1
				}

				results[i].Score = results[i].Score * (
					1.0 +
						(0.05 * float64(len(t)) * float64(len(s))))
			}
		}

		// if we found multiple types, boost yet again to reward better matches
		if foundTerms > 1 {
			results[i].Score = results[i].Score * (1 + 0.05 * float64(foundTerms))
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
