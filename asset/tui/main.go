// SPDX-License-Identifier: MIT OR Unlicense

package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"os"
	"strings"
	"sync"
	"time"
)

type DisplayResult struct {
	Title      *tview.TextView
	Body       *tview.TextView
	BodyHeight int
}

type Result struct {
	Title   string
	Content string
	Score   float64
}

func main() {
	app := tview.NewApplication()

	var overallFlex *tview.Flex
	var inputField *tview.InputField
	var queryFlex *tview.Flex
	var resultsFlex *tview.Flex
	var displayResults []DisplayResult

	//var codeResultMutex sync.Mutex
	var codeResults []Result

	// Sets up all of the UI components we need to actually display
	for i := 1; i < 50; i++ {
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

		displayResults = append(displayResults, DisplayResult{
			Title:      textViewTitle,
			Body:       textViewBody,
			BodyHeight: -1,
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
				// we want to work like fzf for piping into other things hence print out the selected version
				fmt.Println(codeResults[selected].Title)
				os.Exit(0)
			case tcell.KeyTab:
				//app.SetFocus(textView) need to change focus to the others but not the text itself
			case tcell.KeyUp:
				if selected != 0 {
					selected--
				}
				drawResultsState.SetChanged()
			case tcell.KeyDown:
				if selected != len(codeResults)-1 {
					selected++
				}
				drawResultsState.SetChanged()
			}
		})

	// setup the flex containers to have everything rendered neatly
	queryFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 8, false)

	resultsFlex = tview.NewFlex().SetDirection(tview.FlexRow)

	overallFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 1, 0, false).
		AddItem(resultsFlex, 0, 1, false)

	// Add all of the display codeResults into the container ready to be populated
	for _, t := range displayResults {
		resultsFlex.AddItem(nil, 1, 0, false)
		resultsFlex.AddItem(t.Title, 1, 0, false)
		resultsFlex.AddItem(nil, 1, 0, false)
		resultsFlex.AddItem(t.Body, t.BodyHeight, 1, false)
	}

	for i := 1; i < 21; i++ {
		codeResults = append(codeResults, Result{
			Title: fmt.Sprintf(`main.go`),
			Score: float64(i),
			Content: fmt.Sprintf(`%d func NewFlex%d() *Flex {
	f := &Flex{
		Box:       NewBox().SetBackgroundColor(tcell.ColorDefault),
		direction: [red]FlexColumn[white],
	}
	f.focus = f
	return f
}`, 1, i),
		})
	}

	// render loop running background is the only thing responsible for updating
	go func() {
		for {
			drawResults(displayResults, codeResults, selected, resultsFlex, app)
			time.Sleep(20 * time.Millisecond)
		}
	}()

	if err := app.SetRoot(overallFlex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}

type drawResultsStruct struct {
	DrawResultsCount int
	DrawResultsSync sync.Mutex
	DrawResultsChanged bool
}

func (srs *drawResultsStruct) SetChanged() {
	srs.DrawResultsSync.Lock()
	srs.DrawResultsChanged = true
	srs.DrawResultsSync.Unlock()
}

var drawResultsState = drawResultsStruct{}

// This is responsible for drawing all changes on the screen
func drawResults(displayResults []DisplayResult, codeResults []Result, selected int, resultsFlex *tview.Flex, app *tview.Application) {
	drawResultsState.DrawResultsSync.Lock()
	defer drawResultsState.DrawResultsSync.Unlock()
	if !drawResultsState.DrawResultsChanged {
		return
	}
	drawResultsState.DrawResultsCount++

	// reset the elements by clearing out every one
	for _, t := range displayResults {
		t.Title.SetText("")
		t.Body.SetText("")
	}

	// go and get the codeResults the user wants to see IE based on their up/down keypresses
	var p []Result
	for i, t := range codeResults {
		if i >= selected {
			p = append(p, t)
		}
	}

	// render out what the user wants to see based on the results that have been choser
	app.QueueUpdateDraw(func() {
		for i, t := range p {
			displayResults[i].Title.SetText(fmt.Sprintf("%d [fuchsia]%s (%f)[-:-:-]", drawResultsState.DrawResultsCount, t.Title, t.Score))
			displayResults[i].Body.SetText(t.Content)

			// we need to update the item so that it displays everything we have put in
			resultsFlex.ResizeItem(displayResults[i].Body, len(strings.Split(t.Content, "\n")), 0)
		}
	})

	drawResultsState.DrawResultsChanged = false
}
