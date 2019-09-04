package processor

import (
	"testing"
)

func TestRanker(t *testing.T) {
	results := []*FileJob{}
	ranked := RankResults(results)

	if len(ranked) != 0 {
		t.Error("Should be zero results")
	}
}
