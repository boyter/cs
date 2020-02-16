package parser

import "strings"

func ParseArguments(args []string) {
	var str strings.Builder

	for _, arg := range args {
		str.WriteString(strings.TrimSpace(arg))
	}

	// With the arguments cleaned up parse out what we need
}