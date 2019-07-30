package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	grid := tview.NewGrid().
		SetRows(0, 0).
		SetColumns(2, 0).
		SetBorders(false)

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	inputField := tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorBlue).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			// TODO hook into search here
			textView.Clear()

			processor

			_, _ = fmt.Fprintf(textView, "%s ", text)
		})

	grid.AddItem(inputField, 0, 0, 1, 3, 0, 100, false)
	// Layout for screens wider than 100 cells.
	grid.AddItem(textView, 1, 0, 1, 3, 0, 100, false)

	if err := app.SetRoot(grid, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
