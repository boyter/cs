package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	var textView *tview.Box
	var dropdown *tview.DropDown
	var inputField *tview.InputField

	textView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		ScrollToBeginning().
		SetBorder(true).
	    SetTitle("something")

	dropdown = tview.NewDropDown().
		SetLabel("").
		SetOptions([]string{"100", "200", "300", "400", "500"}, nil).
		SetCurrentOption(2).
		SetSelectedFunc(func(text string, index int) {
			app.SetFocus(inputField)
		})

	inputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(0).
		SetChangedFunc(func(text string){
			app.SetFocus(dropdown)
		})

	queryFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 1, false).
		AddItem(dropdown, 0, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 0, 2, false).
		AddItem(textView, 0, 3, false)

	if err := app.SetRoot(flex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
