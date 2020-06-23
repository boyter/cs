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

type displayResult struct {
	Title      *tview.TextView
	Body       *tview.TextView
	BodyHeight int
}

type codeResult struct {
	Title   string
	Content string
	Score   float64
}

type drawResultsStruct struct {
	Query string
	Count   int
	Sync    sync.Mutex
	Changed bool
	Running bool
}

func (drs *drawResultsStruct) SetChanged(b bool) {
	drs.Sync.Lock()
	defer drs.Sync.Unlock()
	drs.Changed = b
}

func (drs *drawResultsStruct) GetChanged() bool {
	drs.Sync.Lock()
	defer drs.Sync.Unlock()
	return drs.Changed
}

func (drs *drawResultsStruct) SetRunning(b bool) {
	drs.Sync.Lock()
	defer drs.Sync.Unlock()
	drs.Running = b
}

func (drs *drawResultsStruct) GetRunning() bool {
	drs.Sync.Lock()
	defer drs.Sync.Unlock()
	return drs.Running
}

// This is responsible for drawing all changes on the screen
func (drs *drawResultsStruct) drawResults(displayResults []displayResult, codeResults []codeResult, selected int, status string, resultsFlex *tview.Flex, statusView *tview.TextView, app *tview.Application) {
	drs.Sync.Lock()
	defer drs.Sync.Unlock()

	if !drs.Changed {
		return
	}

	// NB this is just here so we can see updates in this test
	drs.Count++

	// reset the elements by clearing out every one
	for _, t := range displayResults {
		t.Title.SetText("")
		t.Body.SetText("")
	}

	// go and get the codeResults the user wants to see using selected as the offset to display from
	var p []codeResult
	for i, t := range codeResults {
		if i >= selected {
			p = append(p, t)
		}
	}

	// render out what the user wants to see based on the results that have been chosen
	app.QueueUpdateDraw(func() {
		for i, t := range p {
			displayResults[i].Title.SetText(fmt.Sprintf("%d [fuchsia]%s (%f)[-:-:-]", drs.Count, t.Title, t.Score))
			displayResults[i].Body.SetText(t.Content)

			// we need to update the item so that it displays everything we have put in
			resultsFlex.ResizeItem(displayResults[i].Body, len(strings.Split(t.Content, "\n")), 0)
		}

		statusView.SetText(status)
	})

	// we can only set that nothing
	drs.Changed = false
}

func main() {
	app := tview.NewApplication()
	drawResultsState := drawResultsStruct{}

	var overallFlex *tview.Flex
	var inputField *tview.InputField
	var queryFlex *tview.Flex
	var resultsFlex *tview.Flex
	var statusView *tview.TextView
	var displayResults []displayResult

	var codeResults []codeResult

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

		displayResults = append(displayResults, displayResult{
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
				drawResultsState.SetRunning(true)
				drawResultsState.SetChanged(true)
			case tcell.KeyDown:
				if selected != len(codeResults)-1 {
					selected++
				}
				drawResultsState.SetRunning(false)
				drawResultsState.SetChanged(true)
			}
		}).
		SetChangedFunc(func(text string) {
			text = strings.TrimSpace(text)
			drawResultsState.Query = text
		})

	statusView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		ScrollToBeginning()

	// setup the flex containers to have everything rendered neatly
	queryFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 8, false)

	resultsFlex = tview.NewFlex().SetDirection(tview.FlexRow)

	overallFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(statusView, 1, 0, false).
		AddItem(resultsFlex, 0, 1, false)

	// Add all of the display codeResults into the container ready to be populated
	for _, t := range displayResults {
		resultsFlex.AddItem(nil, 1, 0, false)
		resultsFlex.AddItem(t.Title, 1, 0, false)
		resultsFlex.AddItem(nil, 1, 0, false)
		resultsFlex.AddItem(t.Body, t.BodyHeight, 1, false)
	}

	// add in a few results just to get things going
	for i := 1; i < 21; i++ {
		codeResults = append(codeResults, codeResult{
			Title: fmt.Sprintf(`main.go`),
			Score: float64(i),
			Content: fmt.Sprintf(`func NewFlex%d() *Flex {
	f := &Flex{
		Box:       NewBox().SetBackgroundColor(tcell.ColorDefault),
		direction: [red]FlexColumn[white],
	}
	f.focus = f
	return f
}`, i),
		})
	}

	// trigger the first render without user action
	drawResultsState.SetChanged(true)

	// render loop running background is the only thing responsible for updating the results
	go func() {
		// Used to show what is happening on the page
		var spinString = `\|/-`
		var spinLocation = 0
		var spinRun = 0

		for {
			status := fmt.Sprintf("%d results(s) for '%s' from %d files", len(codeResults), drawResultsState.Query, 87)
			if drawResultsState.GetRunning() {
				status = fmt.Sprintf("%d results(s) for '%s' from %d files %s", len(codeResults), drawResultsState.Query, 87, string(spinString[spinLocation]))

				spinRun++
				if spinRun == 4 {
					spinLocation++
					if spinLocation >= len(spinString) {
						spinLocation = 0
					}
					spinRun = 0
					drawResultsState.SetChanged(true)
				}
			}

			drawResultsState.drawResults(displayResults, codeResults, selected, status, resultsFlex, statusView, app)
			time.Sleep(30 * time.Millisecond)
		}
	}()

	if err := app.SetRoot(overallFlex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}

