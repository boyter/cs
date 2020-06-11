// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package processor

import (
	"testing"
)

// This is based roughly the example provided by https://en.wikipedia.org/wiki/Tf%E2%80%93idf
// with the output for it compared to ensure the results are fairly similar
func TestRankResultsTFIDFTraditional(t *testing.T) {
	ml1 := map[string][][]int{}

	ml2 := map[string][][]int{}
	ml2["example"] = [][]int{{1}, {2}, {3}}

	s := []*fileJob{
		{
			MatchLocations: ml1,
			Location:       "/test/other.go",
			Bytes:          12,
		},
		{
			MatchLocations: ml2,
			Location:       "/test/test.go",
			Bytes:          12,
		},
	}

	s = rankResultsTFIDF(2, s, calculateDocumentFrequency(s), true)

	if s[0].Score > s[1].Score {
		t.Error("index 0 should have lower score than 1")
	}

	if s[1].Score < 0.13 || s[1].Score > 0.16 {
		t.Error("score should be in this range")
	}
}

func TestRankResultsTFIDFComparison(t *testing.T) {
	ml1 := map[string][][]int{}
	ml1["example"] = [][]int{{1}, {2}, {3}}

	s := []*fileJob{
		{
			MatchLocations: ml1,
			Location:       "/test/other.go",
			Bytes:          12,
		},
	}

	s = rankResultsTFIDF(2, s, calculateDocumentFrequency(s), true)
	score1 := s[0].Score

	s = rankResultsTFIDF(2, s, calculateDocumentFrequency(s), false)
	score2 := s[0].Score

	if score1 == score2 {
		t.Error("expected scores to be slightly different")
	}
}

func TestRankResultsRankerComparison(t *testing.T) {
	ml1 := map[string][][]int{}
	ml1["example"] = [][]int{{1}, {2}, {3}}

	s := []*fileJob{
		{
			MatchLocations: ml1,
			Location:       "/test/other.go",
			Bytes:          12,
		},
	}

	s = rankResultsTFIDF(2, s, calculateDocumentFrequency(s), true)
	score1 := s[0].Score

	s = rankResultsTFIDF(2, s, calculateDocumentFrequency(s), false)
	score2 := s[0].Score

	s = rankResultsBM25(2, s, calculateDocumentFrequency(s))
	score3 := s[0].Score

	if score1 == score2 || score1 == score3 || score2 == score3 {
		t.Error("expected scores to be slightly different")
	}
}

func TestRankResultsLocation(t *testing.T) {
	ml := map[string][][]int{}
	ml["test"] = [][]int{{1}, {2}, {3}}

	s := []*fileJob{
		{
			MatchLocations: ml,
			Location:       "/test/other.go",
		},
		{
			MatchLocations: ml,
			Location:       "/test/test.go",
		},
	}

	s = rankResultsLocation(s)

	if s[0].Score > s[1].Score {
		t.Error("index 0 should have lower score than 1")
	}
}

func TestCalculateDocumentFrequency(t *testing.T) {
	ml := map[string][][]int{}
	ml["test"] = [][]int{{1}, {2}, {3}}

	s := []*fileJob{
		{
			MatchLocations: ml,
		},
		{
			MatchLocations: ml,
		},
	}

	freq := calculateDocumentTermFrequency(s)

	if len(freq) != 1 || freq["test"] != 6 {
		t.Error("did not work as expected")
	}
}

func TestSortResults(t *testing.T) {
	s := []*fileJob{
		{
			Filename: "1",
			Location: "",
			Score:    10,
		},
		{
			Filename: "2",
			Location: "",
			Score:    20,
		},
	}
	sortResults(s)

	if s[0].Filename != "2" {
		t.Error("expected 2 first")
	}
}

func TestSortResultsEqualScore(t *testing.T) {
	s := []*fileJob{
		{
			Filename: "1",
			Location: "2",
			Score:    10,
		},
		{
			Filename: "2",
			Location: "1",
			Score:    10,
		},
	}
	sortResults(s)

	if s[0].Filename != "2" {
		t.Error("expected 2 first")
	}
}
