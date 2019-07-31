package main

import (
	"fmt"
	"github.com/boyter/sc/processor"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func updateText(view *tview.TextView) {
	searchMutex.Lock()
	defer searchMutex.Unlock()

	searchSliceMutex.Lock()
	s := ""
	if len(searchSlice) != 0 {
		s = searchSlice[len(searchSlice)-1]
		searchSlice = []string{}
	}
	searchSliceMutex.Unlock()


	if strings.TrimSpace(s) == "" {
		view.Clear()
		return
	}

	processor.SearchString = strings.Split(strings.TrimSpace(s), " ")
	fileListQueue := make(chan *processor.FileJob, 100)                     // Files ready to be read from disk
	fileReadContentJobQueue := make(chan *processor.FileJob, 100) // Files ready to be processed
	fileSummaryJobQueue := make(chan *processor.FileJob, 100)         // Files ready to be summarised

	go processor.WalkDirectoryParallel(filepath.Clean("."), fileListQueue)
	go processor.FileReaderWorker(fileListQueue, fileReadContentJobQueue)
	go processor.FileProcessorWorker(fileReadContentJobQueue, fileSummaryJobQueue)


	results := []*processor.FileJob{}
	count := 0
	for res := range fileSummaryJobQueue {
		count++
		results = append(results, res)

		if count % 20 == 0 {
			drawResults(results, view)
		}
	}

	drawResults(results, view)
}

func drawResults(results []*processor.FileJob, view *tview.TextView) {
	processor.RankResults(results)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	pResults := results
	if len(results) > 10 {
		pResults = results[:10]
	}

	resultText := strconv.Itoa(len(results)) + " result(s)\n\n"
	for _, res := range pResults {
		resultText += fmt.Sprintf("%s (%.3f)", res.Location, res.Score) + "\n"

		locs := []int{}
		for k := range res.Locations {
			locs = append(locs, res.Locations[k]...)
		}
		locs = processor.RemoveIntDuplicates(locs)

		rel := processor.ExtractRelevant(processor.SearchString, string(res.Content), locs, 200, 50, "…")
		resultText += rel + "\n\n"
	}

	view.Clear()
	_, _ = fmt.Fprintf(view, "%s", resultText)
}

var searchSlice = []string{}
var searchMutex sync.Mutex
var searchSliceMutex sync.Mutex

func main() {
	app := tview.NewApplication()

	grid := tview.NewGrid().
		SetRows(0).
		SetColumns(2).
		SetBorders(true)

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

			// when this guy runs we want to add the value onto a slice
			searchSliceMutex.Lock()
			searchSlice = append(searchSlice, text)
			defer searchSliceMutex.Unlock()

			go updateText(textView)
			return

			//if strings.TrimSpace(text) == "" || len(strings.TrimSpace(text)) < 3 {
			//	textView.Clear()
			//	return
			//}
			//
			//time.Sleep(1 * time.Second)
			//return
			//
		})

	grid.AddItem(inputField, 0, 0, 1, 3, 1, 100, false)
	grid.AddItem(textView, 1, 0, 1, 3, 0, 100, false)

	if err := app.SetRoot(grid, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
