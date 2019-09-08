package processor

import "testing"

func TestColourSearchStringEmpty(t *testing.T) {
	content := colorSearchString(&FileJob{})

	if content != "" {
		t.Error("Expected empty content")
	}
}

func TestColourSearchStringContent(t *testing.T) {
	content := colorSearchString(&FileJob{
		Content:   []byte("this is some content"),
		Locations: nil,
	})

	if content != "" {
		t.Error("Expected empty content")
	}
}

func TestColourSearchStringContentWithMatch(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}

	content := colorSearchString(&FileJob{
		Content:   []byte("this is some content"),
		Locations: loc,
	})

	if content != "[red]this[white] is some content" {
		t.Error("Expected empty content")
	}
}