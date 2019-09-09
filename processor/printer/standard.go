package printer

import (
	"strings"
)

func WriteColored(content []byte, locations map[string][]int, in string, out string) string {
	var str strings.Builder

	end := 0

	for i, x := range content {

		// Check if any of the locations match
		for key, value := range locations {
			for _, v := range value {
				if i == v {
					str.WriteString(in)
					end = v + len(key) -1
				}
			}
		}

		str.WriteByte(x)

		if i == end {
			str.WriteString(out)
		}
	}

	return str.String()
}
