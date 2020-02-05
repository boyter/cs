package string

// Simple helper method that removes duplicates from
// any given string slice and then returns a nice
// duplicate free string slice
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
