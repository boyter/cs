package printer

import (
	"strings"
)

func WriteColored(content []byte, locations map[string][]int, in string, out string) string {
	var str strings.Builder

	end := 0
	found := false

	for i, x := range content {

		found = false
		// Check if any of the locations match
		for key, value := range locations {
			for _, v := range value {
				if i == v {
					if !found && end == 0 {
						str.WriteString(in)
						found = true
					}

					// Go for the greatest value of end
					y := v + len(key) - 1
					if y > end {
						end = y
					}
				}
			}
		}

		str.WriteByte(x)

		if i == end {
			str.WriteString(out)
			end = 0
		}
	}

	return str.String()
}
