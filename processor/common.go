// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package processor

import (
	"github.com/boyter/cs/processor/snippet"
	"sort"
)

// Simple helper method that removes duplicates from
// any given int slice and then returns a nice
// duplicate free int slice
func removeIntDuplicates(elements []int) []int {
	encountered := map[int]bool{}
	result := []int{}

	for v := range elements {
		if !encountered[elements[v]] == true {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

func GetResultLocations(res *fileJob) []snippet.LocationType {
	locations := []snippet.LocationType{}
	for k, v := range res.Locations {
		for _, i := range v {
			locations = append(locations, snippet.LocationType{
				Term:     k,
				Location: i,
			})
		}
	}

	sort.Slice(locations, func(i, j int) bool {
		return locations[i].Location < locations[j].Location
	})

	return locations
}
