package processor

import (
	"github.com/boyter/sc/processor/snippet"
	"sort"
)

func RemoveStringDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if !encountered[elements[v]] == true {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

func RemoveIntDuplicates(elements []int) []int {
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

func getResultLocations(res *FileJob) []snippet.LocationType {
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
