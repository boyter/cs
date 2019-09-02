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
	var extInputField *tview.InputField

	textView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		ScrollToBeginning().
		SetBorder(true).
	    SetTitle("something")

	dropdown = tview.NewDropDown().
		SetOptions([]string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900"}, nil).
		SetCurrentOption(3).
		SetLabelColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.Color16).
		SetSelectedFunc(func(text string, index int) {
			app.SetFocus(inputField)
		}).
		SetDoneFunc(func(key tcell.Key){
			switch key {
			case tcell.KeyTab:
				app.SetFocus(inputField)
			}
		})

	extInputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(10).
		SetChangedFunc(func(text string){

		}).
		SetDoneFunc(func(key tcell.Key){
			switch key {
			case tcell.KeyTab:
				app.SetFocus(dropdown)
			}
		})

	inputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(0).
		SetChangedFunc(func(text string){

		}).
		SetDoneFunc(func(key tcell.Key){
			switch key {
			case tcell.KeyTab:
				app.SetFocus(extInputField)
			}
		})

	queryFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 8, false).
		AddItem(extInputField, 10, 0, false).
		AddItem(dropdown, 4, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 1, 0, false).
		AddItem(textView, 0, 3, false)

	if err := app.SetRoot(flex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
