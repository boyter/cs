package processor

import (
	"testing"
)

func TestRanker(t *testing.T) {
	locations := map[string][]int{}
	locations["test"] = []int{1, 2, 3}

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
	ranked := rankResultsTFIDF([][]byte{}, results)

	if len(ranked) != 1 {
		t.Error("Should be one results")
	}

	if ranked[0].Score <= 0 {
		t.Error("Score should be greater than 0")
	}
}

func TestRankResultsLocation(t *testing.T) {
	results := []*FileJob{
		{
			Filename: "test.go",
			Location: "/this/matches/something/test.go",
			Score:    0,
		},
	}
	ranked := rankResultsLocation([][]byte{[]byte("something")}, results)

	if ranked[0].Score == 0 {
		t.Error("Expect rank to be > 0 got", ranked[0].Score)
	}
}

func TestRankResultsLocationScoreCheck(t *testing.T) {
	results := []*FileJob{
		{
			Filename: "test1.go",
			Location: "/this/matches/something/test1.go",
			Score:    0,
		},
		{
			Filename: "test2.go",
			Location: "/this/matches/something/test2.go",
			Score:    0,
		},
	}
	ranked := rankResultsLocation([][]byte{[]byte("something"), []byte("test1")}, results)

	if ranked[0].Score <= ranked[1].Score {
		t.Error("Expect first to get higher match", ranked[0].Score, ranked[1].Score)
	}
}
