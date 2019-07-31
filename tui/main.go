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
	"time"
)

func updateText(textView *tview.TextView) {
	time.Sleep(100 * time.Millisecond)
	searchMutex.Lock()
	defer searchMutex.Unlock()

	searchSliceMutex.Lock()
	var searchTerm string
	if len(searchSlice) != 0 {
		searchTerm = searchSlice[len(searchSlice)-1]
		searchSlice = []string{}
	}
	searchSliceMutex.Unlock()

	if strings.TrimSpace(searchTerm) == "" {
		textView.Clear()
		return
	}

	processor.SearchString = strings.Split(strings.TrimSpace(searchTerm), " ")
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
			drawResults(results, textView, searchTerm)
		}
	}

	drawResults(results, textView, searchTerm)
}

func drawResults(results []*processor.FileJob, textView *tview.TextView, searchTerm string) {
	processor.RankResults(results)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	pResults := results
	if len(results) > 10 {
		pResults = results[:10]
	}

	var resultText string
	resultText += strconv.Itoa(len(results)) + " result(s) for '" + searchTerm + "'\n\n"
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

	textView.Clear()
	_, _ = fmt.Fprintf(textView, "%s", resultText)
	textView.ScrollToBeginning()
}

var searchSlice = []string{}
var searchMutex sync.Mutex
var searchSliceMutex sync.Mutex

func main() {
	app := tview.NewApplication()

	grid := tview.NewGrid().
		SetRows(2).
		SetColumns(0).
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
		SetLabelColor(tcell.ColorBlue).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			searchSliceMutex.Lock()
			searchSlice = append(searchSlice, strings.TrimSpace(text))
			searchSliceMutex.Unlock()

			go updateText(textView)
		})

	grid.AddItem(inputField, 0, 0, 1, 3, 0, 0, false)
	grid.AddItem(textView, 1, 0, 1, 3, 0, 0, false)

	if err := app.SetRoot(grid, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
