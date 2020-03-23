package processor

import (
	"testing"
)

func TestProcessTui(t *testing.T) {
	ProcessTui(false)
}

//func TestTuiSearch(t *testing.T) {
//	app := tview.NewApplication()
//	textview := tview.NewTextView()
//
//	app.SetRoot(textview, true)
//
//	tuiSearch(app, textview, "something")
//}
