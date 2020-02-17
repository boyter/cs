package parser

import (
	"strings"
)

const (
	Default int64 = 0
	Quoted  int64 = 1
	Regex   int64 = 2
	Negated int64 = 3
	Fuzzy1  int64 = 4
	Fuzzy2  int64 = 5
)

type SearchParams struct {
	Term string
	Type int64
}

// Cheap and nasty parser. Needs to be reworked
// to provide real boolean logic with AND OR NOT
// but does enough for this
func ParseArguments(args []string) []SearchParams {
	cleanArgs := []string{}

	// Clean the arguments to avoid redundant spaces and the like
	for _, arg := range args {
		cleanArgs = append(cleanArgs, strings.TrimSpace(arg))
	}

	searchParams := []SearchParams{}
	startIndex := 0
	mode := Default

	// With the arguments cleaned up parse out what we need
	for ind, arg := range cleanArgs {
		if strings.HasPrefix(arg, `"`) {
			if strings.HasSuffix(arg, `"`) {
				searchParams = append(searchParams, SearchParams{
					Term: arg,
					Type: Quoted,
				})
			} else {
				mode = Quoted
				startIndex = ind
			}
		} else if mode == Quoted && strings.HasSuffix(arg, `"`) {
			// quote
			searchParams = append(searchParams, SearchParams{
				Term: strings.Join(cleanArgs[startIndex:ind+1], " "),
				Type: Quoted,
			})
			mode = Default
		} else if strings.HasPrefix(arg, `/`) {
			// If we end with / not prefixed with a \ we are done
			if strings.HasSuffix(arg, `/`) {
				searchParams = append(searchParams, SearchParams{
					Term: arg,
					Type: Regex,
				})
			} else {
				mode = Regex
				startIndex = ind
			}
		} else if mode == Regex && strings.HasSuffix(arg, `/`) {
			// quote
			searchParams = append(searchParams, SearchParams{
				Term: strings.Join(cleanArgs[startIndex:ind+1], " "),
				Type: Regex,
			})
			mode = Default
		} else if arg == "NOT" {
			// If we start with NOT we cannot negate so ignore
			if ind != 0 {
				searchParams = append(searchParams, SearchParams{
					Term: arg,
					Type: Negated,
				})
			}
		} else if strings.HasSuffix(arg, "~1") {
			searchParams = append(searchParams, SearchParams{
				Term: arg,
				Type: Fuzzy1,
			})
		} else if strings.HasSuffix(arg, "~2") {
			searchParams = append(searchParams, SearchParams{
				Term: arg,
				Type: Fuzzy2,
			})
		} else {
			searchParams = append(searchParams, SearchParams{
				Term: arg,
				Type: Default,
			})
		}
	}

	return searchParams
}
