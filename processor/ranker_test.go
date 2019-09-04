package processor

import (
	"testing"
)

func TestRanker(t *testing.T) {

	locations := map[string][]int{}
	locations["test"] = []int{1,2,3}

	results := []*FileJob{
		{
			Filename:  "test.go",
			Extension: "go",
			Location:  "/",
			Content:   []byte("content"),
			Bytes:     10,
			Hash:      nil,
			Binary:    false,
			Score:     0,
			Locations: locations,
		},
	}
	ranked := RankResults(results)

	if len(ranked) != 1 {
		t.Error("Should be one results")
	}

	if ranked[0].Score <= 0 {
		t.Error("Score should be greater than 0")
	}
}
