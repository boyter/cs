package processor

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

type searchParams struct {
	Term string
	Type int64
}

// Cheap and nasty parser. Needs to be reworked
// to provide real boolean logic with AND OR NOT
// but does enough for now
func parseArguments(args []string) []searchParams {
	cleanArgs := []string{}

	// Clean the arguments to avoid redundant spaces and the like
	for _, arg := range args {
		cleanArgs = append(cleanArgs, strings.TrimSpace(arg))
	}

	params := []searchParams{}
	startIndex := 0
	mode := Default

	// With the arguments cleaned up parse out what we need
	// note that this is very ugly
	for ind, arg := range cleanArgs {
		if strings.HasPrefix(arg, `"`) {
			if len(arg) != 1 {
				if strings.HasSuffix(arg, `"`) {
					params = append(params, searchParams{
						Term: arg[1 : len(arg)-1],
						Type: Quoted,
					})
				} else {
					mode = Quoted
					startIndex = ind
				}
			}
		} else if mode == Quoted && strings.HasSuffix(arg, `"`) {

			t := strings.Join(cleanArgs[startIndex:ind+1], " ")
			params = append(params, searchParams{
				Term: t[1 : len(t)-1],
				Type: Quoted,
			})
			mode = Default
		} else if strings.HasPrefix(arg, `/`) {
			if len(arg) != 1 {
				// If we end with / not prefixed with a \ we are done
				if strings.HasSuffix(arg, `/`) {
					params = append(params, searchParams{
						Term: arg[1 : len(arg)-1],
						Type: Regex,
					})
				} else {
					mode = Regex
					startIndex = ind
				}
			}
		} else if mode == Regex && strings.HasSuffix(arg, `/`) {
			t := strings.Join(cleanArgs[startIndex:ind+1], " ")
			params = append(params, searchParams{
				Term: t[1 : len(t)-1],
				Type: Regex,
			})
			mode = Default
		} else if arg == "NOT" {
			// If we start with NOT we cannot negate so ignore
			if ind != 0 {
				params = append(params, searchParams{
					Term: arg,
					Type: Negated,
				})
			}
		} else if strings.HasSuffix(arg, "~1") {
			params = append(params, searchParams{
				Term: strings.TrimRight(arg, "~1"),
				Type: Fuzzy1,
			})
		} else if strings.HasSuffix(arg, "~2") {
			params = append(params, searchParams{
				Term: strings.TrimRight(arg, "~2"),
				Type: Fuzzy2,
			})
		} else {
			params = append(params, searchParams{
				Term: arg,
				Type: Default,
			})
		}
	}

	// If the user didn't end properly that's ok lets do it for them
	if mode == Regex {
		t := strings.Join(cleanArgs[startIndex:], " ")
		params = append(params, searchParams{
			Term: t[1:],
			Type: Regex,
		})
	}
	if mode == Quoted {
		t := strings.Join(cleanArgs[startIndex:], " ")
		params = append(params, searchParams{
			Term: t[1:],
			Type: Quoted,
		})
	}

	return params
}
