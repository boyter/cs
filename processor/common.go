// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package processor

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

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}