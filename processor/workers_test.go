package processor

import (
	"testing"
)

func TestProcessMatchesSingleMatch(t *testing.T) {
	StopProcessing = false
	ResultLimit = 100
	SearchString = []string{
		"match",
	}

	res := FileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, "this is a match")

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score != 1 {
		t.Errorf("Score should be 1")
	}
}

func TestProcessMatchesTwoMatch(t *testing.T) {
	StopProcessing = false
	ResultLimit = 100
	SearchString = []string{
		"match",
		"this",
	}

	res := FileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, "this is a match")

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score != 2 {
		t.Errorf("Score should be 1")
	}
}

func TestProcessMatchesTwoAndMatch(t *testing.T) {
	StopProcessing = false
	ResultLimit = 100
	SearchString = []string{
		"match",
		"AND",
		"this",
	}

	res := FileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, "this is a match")

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score != 2 {
		t.Errorf("Score should be 1")
	}
}

func TestProcessMatchesTwoNotMatch(t *testing.T) {
	StopProcessing = false
	ResultLimit = 100
	SearchString = []string{
		"match",
		"NOT",
		"this",
	}

	res := FileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, "this is a match")

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score != 0 {
		t.Error("Score should be 0 got", res.Score)
	}
}

func TestProcessMatchesFuzzyOne(t *testing.T) {
	StopProcessing = false
	ResultLimit = 100
	SearchString = []string{
		"this~1",
	}

	res := FileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, "this is a match")

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score != 4 {
		t.Error("Score should be 4 got", res.Score)
	}
}

func TestProcessMatchesFuzzyTwo(t *testing.T) {
	StopProcessing = false
	ResultLimit = 100
	SearchString = []string{
		"this~2",
	}

	res := FileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, "this is a match")

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score != 2 {
		t.Error("Score should be 2 got", res.Score)
	}
}
