// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package processor

import "testing"

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

	freq := calculateDocumentFrequency(s)

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
