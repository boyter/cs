package processor

func removeStringDuplicates(elements []string) []string {
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
