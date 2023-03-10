// SPDX-License-Identifier: MIT OR Unlicense

package processor

import (
	"strconv"
)

// Simple helper method that removes duplicates from
// any given int slice and then returns a nice
// duplicate free int slice
func removeIntDuplicates(elements []int) []int {
	encountered := map[int]bool{}
	var result []int

	for v := range elements {
		if !encountered[elements[v]] {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func tryParseInt(x string, def int) int {
	t, err := strconv.Atoi(x)

	if err != nil {
		return def
	}

	return t
}
