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

func TestGetResultLocations(t *testing.T) {
	locations := GetResultLocations(&FileJob{
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
		t.Error("Should get no results")
	}
}
