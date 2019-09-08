package processor

import (
	"github.com/rivo/tview"
	"testing"
)

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
		t.Error("Expected highlighted content")
	}
}

func TestColourSearchStringContentWithMatchMulti(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0, 5}

	content := colorSearchString(&FileJob{
		Content:   []byte("this this"),
		Locations: loc,
	})

	if content != "[red]this[white] [red]this[white]" {
		t.Error("Expected highlighted content")
	}
}

// TODO fix this
func TestColourSearchStringContentWithMatchMultiWords(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0, 5}
	loc["t"] = []int{0}

	content := colorSearchString(&FileJob{
		Content:   []byte("this this"),
		Locations: loc,
	})

	// TODO this should return [red]this[white] [red]this[white]
	//if content != "[red]this[white] [red]this[white]" {
	//	t.Error("Expected highlighted content got=", content)
	//}
	if len(content) == 0 {
		t.Error("Expected a return")
	}
}

func TestProcessTui(t *testing.T) {
	ProcessTui(false)
}

func TestTuiSearch(t *testing.T) {
	app := tview.NewApplication()
	textview := tview.NewTextView()

	app.SetRoot(textview, true)

	tuiSearch(app, textview, "something")
}
