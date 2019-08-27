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

const letterDigitFuzzyBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

func makeFuzzy(term string) []string {
	vals := []string{}

	for i := 1; i < len(term); i++ {
		for _, b := range letterDigitFuzzyBytes {
			vals = append(vals, term[:i] + string(b) + term[i+1:])
		}
	}

	return vals
}
