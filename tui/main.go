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

func main() {
	app := tview.NewApplication()

	grid := tview.NewGrid().
		SetRows(0, 0).
		SetColumns(2, 0).
		SetBorders(true)

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	var mutex sync.Mutex
	inputField := tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorBlue).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			if strings.TrimSpace(text) == "" {
				return
			}

			mutex.Lock()
			defer mutex.Unlock()

			// TODO hook into search here
			textView.Clear()

			processor.SearchString = strings.Split(text, " ")
			fileListQueue := make(chan *processor.FileJob, 100)                     // Files ready to be read from disk
			fileReadContentJobQueue := make(chan *processor.FileJob, 100) // Files ready to be processed
			fileSummaryJobQueue := make(chan *processor.FileJob, 100)         // Files ready to be summarised

			go processor.WalkDirectoryParallel(filepath.Clean("."), fileListQueue)
			go processor.FileReaderWorker(fileListQueue, fileReadContentJobQueue)
			go processor.FileProcessorWorker(fileReadContentJobQueue, fileSummaryJobQueue)

			results := []*processor.FileJob{}
			for res := range fileSummaryJobQueue {
				results = append(results, res)
			}

			processor.RankResults(results)
			sort.Slice(results, func(i, j int) bool {
				return results[i].Score > results[j].Score
			})

			resultText := ""
			for _, res := range results {
				resultText += fmt.Sprintf("%s (%.3f)", res.Location, res.Score) + "\n"

				locs := []int{}
				for k := range res.Locations {
					locs = append(locs, res.Locations[k]...)
				}
				locs = processor.RemoveIntDuplicates(locs)

				rel := processor.ExtractRelevant(processor.SearchString, string(res.Content), locs, 300, 50, "…")
				resultText += rel + "\n"
			}

			resultText = ""
			for i := 0; i< 100; i++ {
				resultText += "" + strconv.Itoa(i) + "\n"
			}

			_, _ = fmt.Fprintf(textView, "%s", resultText)
		})

	grid.AddItem(inputField, 0, 0, 1, 3, 0, 100, false)
	// Layout for screens wider than 100 cells.
	grid.AddItem(textView, 1, 0, 1, 3, 0, 100, false)

	if err := app.SetRoot(grid, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
