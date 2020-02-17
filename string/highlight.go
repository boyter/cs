package string

import (
	"strings"
)

// HighlightString takes in some content and locations and then inserts in/out
// strings which can be used for highlighting around matching terms. For example
// you could pass in "test" and have it return "<strong>te</strong>st"
// locations accepts output from regex.FindAllIndex IndexAllIgnoreCaseUnicode or IndexAll
func HighlightString(content string, locations [][]int, in string, out string) string {
	var str strings.Builder

	end := -1
	found := false

	// Range over string which is rune aware so even if we get invalid
	// locations we should hopefully ignore them as the byte offset wont
	// match
	for i, x := range content {
		found = false

		// TODO be sure to profile the below, because previously needed
		// a small lookup cache to cut down on the work done here

		// Find which of the locations match
		// and if so write the start string
		for _, val := range locations {

			// We have a match where the outer index matches
			// against the val[0] which contains the location of the match
			if i == val[0] {
				// We only write the found string once per match and
				// only if we are not in the middle of one
				if !found && end <= 0 {
					str.WriteString(in)
					found = true
				}

				// Determine the expected end location for this match
				// and only if its further than the expected end do we
				// change to deal with overlaps if say we are trying to match
				// on t and tes against test where we want tes as the longest
				// match to be the end that's written
				y := val[1] - 1 // val[1] in this case is the length of the match
				if y > end {
					end = y
				}
			}
		}

		str.WriteString(string(x))

		// If at the end, and its not -1 meaning the first char
		// which should never happen (I hope!) then write the end string
		if i == end && end != -1 {
			str.WriteString(out)
			end = 0
		}
	}

	return str.String()
}
