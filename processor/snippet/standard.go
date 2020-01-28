package snippet

import (
	"strings"
)

// WriteColoured takes in some content and locations and then inserts in/out
// strings which can be used for highlighting around matching terms. For example
// you could pass in "test" and have it return "<strong>te</strong>st"
func WriteHighlights(content []byte, locations map[string][]int, in string, out string) string {
	var str strings.Builder

	end := -1
	found := false

	for i, x := range content {
		found = false
		// Check if any of the locations match
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

		str.WriteByte(x)

		// If at the end, and its not -1 meaning the first char
		// which should never happen then write the end string
		if i == end && end != -1 {
			str.WriteString(out)
			end = 0
		}
	}

	return str.String()
}
