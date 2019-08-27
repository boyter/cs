package processor

import (
	"time"
)

// Returns the current time as a millisecond timestamp
func makeTimestampMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// Returns the current time as a nanosecond timestamp as some things
// are far too fast to measure using nanoseconds
func makeTimestampNano() int64 {
	return time.Now().UnixNano()
}

const letterDigitFuzzyBytes = " abcdefghijklmnopqrstuvwxyz0123456789"

func makeFuzzyDistanceOne(term string) []string {
	vals := []string{}

	// Maybe they mistyped a single letter?
	// or added an additional one
	for i := 1; i < len(term); i++ {
		for _, b := range letterDigitFuzzyBytes {
			if string(b) == " " {
				vals = append(vals, term[:i] + term[i+1:])
			} else {
				vals = append(vals, term[:i] + string(b) + term[i+1:])
			}
		}
	}

	return vals
}

func makeFuzzyDistanceTwo(term string) []string {
	vals := makeFuzzyDistanceOne(term)

	// Maybe they forgot to type a letter? Try adding one
	for i := 0; i < len(term) + 1; i++ {
		for _, b := range letterDigitFuzzyBytes {
			if string(b) != " " {
				vals = append(vals, term[:i] + string(b) + term[i:])
			}
		}
	}

	return RemoveStringDuplicates(vals)
}
