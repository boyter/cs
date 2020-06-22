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

type Result struct {
	Title string
	Content string
	Score float64
}

func main() {
	app := tview.NewApplication()

	var inputField *tview.InputField
	var queryFlex *tview.Flex
	var resultsFlex *tview.Flex
	var overallFlex *tview.Flex

	var codeResults []CodeResult
	var results []Result

	for i:=1;i<21;i++ {
		results = append(results, Result{
			Title: fmt.Sprintf(`main.go`),
			Score: float64(i),
			Content:  fmt.Sprintf(`func NewFlex%d() *Flex {
	f := &Flex{
		Box:       NewBox().SetBackgroundColor(tcell.ColorDefault),
		direction: FlexColumn,
	}
	f.focus = f
	return f
}`, i),
		})
	}

	// setup all of the display results
	for i:=1;i<21;i++ {
		var textViewTitle *tview.TextView
		var textViewBody *tview.TextView

		textViewTitle = tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			ScrollToBeginning()

		textViewBody = tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			ScrollToBeginning()

		codeResults = append(codeResults, CodeResult{
			Title: textViewTitle,
			Body:  textViewBody,
			BodyHeight: 1,
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
			case tcell.KeyEnter:
				app.Stop()
				fmt.Println(results[selected].Title)
			case tcell.KeyTab:
			//app.SetFocus(textView) need to change focus to the others but not the text itself
			case tcell.KeyUp:
				if selected != 0 {
					selected--
				}

				for _, t := range codeResults {
					t.Title.SetText("")
					t.Body.SetText("")
				}

				var p []Result
				for i, t := range results {
					if i >= selected {
						p = append(p, t)
					}
				}

				for i, t := range p {
					codeResults[i].Title.SetText(fmt.Sprintf("[purple]%s (%f)[white]", t.Title, t.Score))
					codeResults[i].Body.SetText(t.Content)
					resultsFlex.ResizeItem(codeResults[i].Body, len(strings.Split(t.Content, "\n")), 0)
				}

			case tcell.KeyDown:
				if selected != len(codeResults) -1 {
					selected++
				}

				for _, t := range codeResults {
					t.Title.SetText("")
					t.Body.SetText("")
				}

				var p []Result
				for i, t := range results {
					if i >= selected {
						p = append(p, t)
					}
				}

				for i, t := range p {
					codeResults[i].Title.SetText(fmt.Sprintf("[purple]%s (%f)[white]", t.Title, t.Score))
					codeResults[i].Body.SetText(t.Content)
					resultsFlex.ResizeItem(codeResults[i].Body, len(strings.Split(t.Content, "\n")), 0)
				}

			}
		})

	queryFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 8, false)

	resultsFlex = tview.NewFlex().SetDirection(tview.FlexRow)

	overallFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 1, 0, false).
		AddItem(resultsFlex, 0, 1, false)

	// Add all of the display results into the container ready to be populated
	for _, t := range codeResults {
		resultsFlex.AddItem(nil, 1, 0, false)
		resultsFlex.AddItem(t.Title, 1, 0, false)
		resultsFlex.AddItem(nil, 1, 0, false)
		resultsFlex.AddItem(t.Body, t.BodyHeight, 1, false)
	}

	if err := app.SetRoot(overallFlex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
