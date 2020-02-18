package processor

import (
	"testing"
)

func TestRemoveIntDuplicates(t *testing.T) {
	clean := RemoveIntDuplicates([]int{1, 1})

	if len(clean) != 1 {
		t.Error("Should have no duplicates")
	}
}

func TestGetResultLocationsZeroLocations(t *testing.T) {

	locations := GetResultLocations(&fileJob{
		Filename:  "test",
		Extension: "test",
		Location:  "test",
		Content:   []byte("test"),
		Bytes:     0,
		Hash:      nil,
		Binary:    false,
		Score:     0,
		Locations: map[string][]int{},
	})

	if len(locations) != 0 {
		t.Error("Should get no results got", len(locations))
	}
}

func TestGetResultLocationsThreeResults(t *testing.T) {
	loc := map[string][]int{}
	loc["test"] = []int{1, 2, 3}

	locations := GetResultLocations(&fileJob{
		Filename:  "test",
		Extension: "test",
		Location:  "test",
		Content:   []byte("test"),
		Bytes:     0,
		Hash:      nil,
		Binary:    false,
		Score:     0,
		Locations: loc,
	})

	if len(locations) != 3 {
		t.Error("Should get 3 results got", len(locations))
	}
}
