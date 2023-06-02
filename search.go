// SPDX-License-Identifier: MIT OR Unlicense

package main

import (
	"errors"
	str "github.com/boyter/go-string"
	"regexp"
	"strings"
)

const (
	Default = iota
	Quoted
	Regex
	Negated
	Fuzzy1
	Fuzzy2
)

type searchParams struct {
	Term string
	Type int64
}

// ParseQuery is a cheap and nasty parser. Needs to be reworked
// to provide real boolean logic with AND OR NOT
// but does enough for now
func ParseQuery(args []string) []searchParams {
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
					// If the term is // don't treat it as a regex treat it as a search for //
					if arg == "//" {
						params = append(params, searchParams{
							Term: "//",
							Type: Default,
						})
					} else {
						params = append(params, searchParams{
							Term: arg[1 : len(arg)-1],
							Type: Regex,
						})
					}
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

func regexSearch(needle searchParams, content string) (x [][]int, err error) {
	// Its possible the user supplies an invalid regex and if so we should not crash
	// but ignore it
	defer func() {
		if recover() != nil {
			err = errors.New("regex compile failure issue")
		}
	}()

	r := regexp.MustCompile("(?i)" + needle.Term)
	return r.FindAllIndex([]byte(content), 20), nil
}

const letterDigitFuzzyBytes = `abcdefghijklmnopqrstuvwxyz1234567890`

// Takes in a term and returns a slice of them which contains all the
// fuzzy versions of that str with things such as mis-spellings
// somewhat based on https://norvig.com/spell-correct.html
func makeFuzzyDistanceOne(term string) []string {
	vals := []string{term}

	if len(term) <= 2 {
		return vals
	}

	// Delete letters so turn "test" into "est" "tst" "tet"
	for i := 0; i < len(term); i++ {
		vals = append(vals, term[:i]+term[i+1:])
	}

	// Replace a letter or digit which effectively does transpose for us
	for i := 0; i < len(term); i++ {
		for _, b := range letterDigitFuzzyBytes {
			vals = append(vals, term[:i]+string(b)+term[i+1:])
		}
	}

	// Insert a letter or digit
	for i := 0; i < len(term); i++ {
		for _, b := range letterDigitFuzzyBytes {
			vals = append(vals, term[:i]+string(b)+term[i:])
		}
	}

	return str.RemoveStringDuplicates(vals)
}

// Similar to fuzzy 1 but in this case we add letters
// to make the distance larger
func makeFuzzyDistanceTwo(term string) []string {
	vals := makeFuzzyDistanceOne(term)

	// Maybe they forgot to type a letter? Try adding one
	for i := 0; i < len(term)+1; i++ {
		for _, b := range letterDigitFuzzyBytes {
			vals = append(vals, term[:i]+string(b)+term[i:])
		}
	}

	return str.RemoveStringDuplicates(vals)
}
