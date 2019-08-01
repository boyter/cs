package main

import (
	"fmt"
	"github.com/boyter/sc/processor"
	"github.com/boyter/sc/processor/snippet"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func search(textView *tview.TextView) {
	// Kill off anything else that's processing
	processor.StopProcessing = true
	// Wait a bit for everything to die and so the user can type a little more
	time.Sleep(100 * time.Millisecond)

	searchMutex.Lock()
	defer searchMutex.Unlock()
	shouldSpin = true
	// Enable processing again
	processor.StopProcessing = false

	searchSliceMutex.Lock()
	var searchTerm string
	if len(searchSlice) != 0 {
		searchTerm = searchSlice[len(searchSlice)-1]
		searchSlice = []string{}
	} else {
		// If the slice is empty we want to bail out
		searchSliceMutex.Unlock()
		shouldSpin = false
		return
	}
	searchSliceMutex.Unlock()

	if strings.TrimSpace(searchTerm) == "" {
		drawText(textView, "")
		shouldSpin = false
		return
	}

	processor.SearchString = strings.Split(strings.TrimSpace(searchTerm), " ")
	fileListQueue := make(chan *processor.FileJob, runtime.NumCPU())           // Files ready to be read from disk
	fileReadContentJobQueue := make(chan *processor.FileJob, runtime.NumCPU()) // Files ready to be processed
	fileSummaryJobQueue := make(chan *processor.FileJob, runtime.NumCPU())     // Files ready to be summarised

	processor.TotalCount = 0
	go processor.WalkDirectoryParallel(filepath.Clean("."), fileListQueue)
	go processor.FileReaderWorker(fileListQueue, fileReadContentJobQueue)
	go processor.FileProcessorWorker(fileReadContentJobQueue, fileSummaryJobQueue)

	results := []*processor.FileJob{}
	count := 0
	for res := range fileSummaryJobQueue {
		count++
		results = append(results, res)

		if count%20 == 0 {
			drawResults(results, textView, searchTerm)
		}
	}

	drawResults(results, textView, searchTerm)
	shouldSpin = false
}

func drawResults(results []*processor.FileJob, textView *tview.TextView, searchTerm string) {
	processor.RankResults(results)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if int64(len(results)) > processor.TotalCount {
		results = results[:processor.TotalCount]
	}

	pResults := results
	if len(results) > 20 {
		pResults = results[:20]
	}

	var resultText string
	resultText += strconv.Itoa(len(results)) + " result(s) for '" + searchTerm + "'\n\n"
	for i, res := range pResults {
		resultText += fmt.Sprintf("%d. %s (%.3f)", i + 1, res.Location, res.Score) + "\n\n"

		locs := []int{}
		for k := range res.Locations {
			locs = append(locs, res.Locations[k]...)
		}
		locs = processor.RemoveIntDuplicates(locs)

		rel := snippet.ExtractRelevant(processor.SearchString, string(res.Content), locs, int(processor.SnippetLength), 50, "…")
		resultText += rel + "\n\n"
	}

	drawText(textView, resultText)
	shouldSpin = false
}

func drawText(textView *tview.TextView, text string) {
	textView.Clear()
	_, _ = fmt.Fprintf(textView, "%s", text)
	textView.ScrollToBeginning()
}

var searchSlice = []string{}
var searchMutex sync.Mutex
var searchSliceMutex sync.Mutex
var spinnerString = `\|/-`
var shouldSpin = false

func runningIndicator(app *tview.Application, inputField *tview.InputField) {
	for {
		time.Sleep(100 * time.Millisecond)
		var i int
		for shouldSpin {
			inputField.SetLabel(string(spinnerString[i]) + " ")
			app.Draw()
			time.Sleep(100 * time.Millisecond)

			i++
			if i >= len(spinnerString) {
				i = 0
			}
		}

		inputField.SetLabel("> ")
		app.Draw()
	}
}

func main() {
	app := tview.NewApplication()

	grid := tview.NewGrid().
		SetRows(2).
		SetColumns(1).
		SetBorders(false)

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		ScrollToBeginning().
		SetChangedFunc(func() {
			app.Draw()
		})

	inputField := tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			searchSliceMutex.Lock()
			searchSlice = append(searchSlice, strings.TrimSpace(text))
			searchSliceMutex.Unlock()

			go search(textView)
		})

	grid.AddItem(inputField, 0, 0, 1, 3, 0, 1, false)
	grid.AddItem(textView, 1, 0, 1, 3, 0, 0, false)

	go runningIndicator(app, inputField)

	if err := app.SetRoot(grid, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
