package string


import (
	"strings"
)

// WriteColoured takes in some content and locations and then inserts in/out
// strings which can be used for highlighting around matching terms. For example
// you could pass in "test" and have it return "<strong>te</strong>st"
// TODO locations should be modified to take in the output from IndexAll which is [][]int
func WriteHighlights(content string, locations map[string][]int, in string, out string) string {
	var str strings.Builder

	end := -1
	found := false

	// Cheap and nasty cache to avoid looping the map too much when we range over content
	locationLookup := map[int]int{}
	for _, value := range locations {
		for _, v := range value {
			locationLookup[v] = 0
		}
	}

	// Range over string which is rune aware so even if we get invalid
	// locations we should hopefully ignore them as the byte offset wont
	// match
	for i, x := range content {
		found = false

		_, ok := locationLookup[i]

		if ok {
			// Find which of the locations match
			// and if so write the start string
			for key, value := range locations {
				for _, v := range value {
					if i == v {
						// We only write the found string once per match and
						// only if we are not in the middle of one
						if !found && end <= 0 {
							str.WriteString(in)
							found = true
						}

						// Go for the greatest value of end
						// and always check if it should be pushed out
						// so we can cover cases where overlaps occur
						y := v + len(key) - 1
						if y > end {
							end = y
						}
					}
				}
			}
		}

		str.WriteString(string(x))

		// If at the end, and its not -1 meaning the first char
		// which should never happen then write the end string
		if i == end && end != -1 {
			str.WriteString(out)
			end = 0
		}
	}

	return str.String()
}

