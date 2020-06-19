// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"strings"
)

type CodeResult struct {
	Title *tview.TextView
	Body *tview.TextView
	BodyHeight int
}

func main() {
	app := tview.NewApplication()

	var inputField *tview.InputField
	var queryFlex *tview.Flex
	var resultsFlex *tview.Flex
	var overallFlex *tview.Flex


	var codeResults []CodeResult

	for i:=1;i<100;i++ {
		var textViewTitle *tview.TextView
		var textViewBody *tview.TextView

		textViewTitle = tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			ScrollToBeginning()
		textViewTitle.SetText(fmt.Sprintf(`[purple]main%d.go (0.0%d)[white]`, i, i))

		textViewBody = tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			ScrollToBeginning()

		textViewBody.SetText(`import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type CodeResult struct {
	Title *tview.TextView
	Body *tview.TextView
}`)

		codeResults = append(codeResults, CodeResult{
			Title: textViewTitle,
			Body:  textViewBody,
			BodyHeight: len(strings.Split(textViewBody.GetText(false), "\n")),
		})
	}

	selected := 0

	inputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(0).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
			//app.SetFocus(textView) need to change focus to the others but not the text itself
			case tcell.KeyUp:
				if selected != 0 {
					selected--
				}

				resultsFlex.Clear()
				for i, t := range codeResults {
					if i >= selected {
						resultsFlex.AddItem(nil, 1, 0, false)
						resultsFlex.AddItem(t.Title, 1, 0, false)
						resultsFlex.AddItem(nil, 1, 0, false)
						resultsFlex.AddItem(t.Body, t.BodyHeight, 1, false)
					}
				}
			case tcell.KeyDown:
				if selected != len(codeResults) -1 {
					selected++
				}

				resultsFlex.Clear()
				for i, t := range codeResults {
					if i >= selected {
						resultsFlex.AddItem(nil, 1, 0, false)
						resultsFlex.AddItem(t.Title, 1, 0, false)
						resultsFlex.AddItem(nil, 1, 0, false)
						resultsFlex.AddItem(t.Body, t.BodyHeight, 1, false)
					}
				}
			}
		})


	queryFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 8, false)

	resultsFlex = tview.NewFlex().SetDirection(tview.FlexRow)

	overallFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 1, 0, false).
		AddItem(resultsFlex, 0, 1, false)


	for i, t := range codeResults {
		if i >= selected {
			resultsFlex.AddItem(nil, 1, 0, false)
			resultsFlex.AddItem(t.Title, 1, 0, false)
			resultsFlex.AddItem(nil, 1, 0, false)
			resultsFlex.AddItem(t.Body, t.BodyHeight, 1, false)
		}
	}

	if err := app.SetRoot(overallFlex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
