// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	var dropdown *tview.DropDown
	var inputField *tview.InputField
	var extInputField *tview.InputField

	textView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true).
		ScrollToBeginning()

	list := NewCsList().
		AddItem("csperf/main.go (0.003)", "Some explanatory text", '1', func() {
			app.Stop()
			fmt.Println("csperf/main.go")
		}).
		AddItem("corpus/1080-0.txt", `var textView *tview.TextView
	var dropdown *tview.DropDown
	var inputField *tview.InputField
	var extInputField *tview.InputField

	textView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true).
		ScrollToBeginning()`, '2', nil).
		AddItem("corpus/pg16328.txt", "Some explanatory text", '3', nil).
		AddItem("corpus/84-0.txt", "Some explanatory text", '4', nil).
		AddItem("corpus/pg2542.txt", "Some explanatory text", '5', nil).
		AddItem("corpus/snippet/chinese_war_and_peace.txt", "Some explanatory text", '6', nil).
		AddItem("corpus/844.txt.utf-8", "Some explanatory text", '7', nil).
		AddItem("corpus/46-0.txt", "Some explanatory text Some explanatory\n text Some explanatory text Some explanatory text\n Some explanatory text Some explanatory text Some explanatory text Some\n explanatory text Some explanatory text ", 'h', nil).
		AddItem("corpus/prideandprejudice.txt", "Some explanatory text", '8', nil).
		AddItem("corpus/2701-0.txt", "Some explanatory text", '9', nil).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})

	dropdown = tview.NewDropDown().
		SetOptions([]string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900"}, nil).
		SetCurrentOption(3).
		SetLabelColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.Color16).
		SetSelectedFunc(func(text string, index int) {
			app.SetFocus(inputField)
		}).
		SetDoneFunc(func(key tcell.Key) {
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
		SetChangedFunc(func(text string) {

		}).
		SetDoneFunc(func(key tcell.Key) {
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
		SetChangedFunc(func(text string) {

		}).
		SetDoneFunc(func(key tcell.Key) {
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
		//AddItem(textView, 0, 3, false)
		AddItem(list, 0, 3, false)


	if err := app.SetRoot(flex, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}
