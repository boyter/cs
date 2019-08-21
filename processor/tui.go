package processor

import (
	"fmt"
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

func tuiSearch(app *tview.Application, textView *tview.TextView) {
	// Kill off anything else that's processing
	StopProcessing = true
	// Wait a bit for everything to die and so the user can type a little more
	time.Sleep(100 * time.Millisecond)

	searchMutex.Lock()
	defer searchMutex.Unlock()

	searchSliceMutex.Lock()
	var searchTerm string
	if len(searchSlice) != 0 {
		searchTerm = searchSlice[len(searchSlice)-1]
		searchSlice = []string{}
	} else {
		// If the slice is empty we want to bail out
		searchSliceMutex.Unlock()
		return
	}
	searchSliceMutex.Unlock()

	if strings.TrimSpace(searchTerm) == "" {
		drawText(textView, "")
		return
	}

	SearchString = strings.Split(strings.TrimSpace(searchTerm), " ")
	CleanSearchString()
	TotalCount = 0

	// Enable processing again
	StopProcessing = false

	fileListQueue := make(chan *FileJob, runtime.NumCPU())           // Files ready to be read from disk
	fileReadContentJobQueue := make(chan *FileJob, runtime.NumCPU()) // Files ready to be processed
	fileSummaryJobQueue := make(chan *FileJob, runtime.NumCPU())     // Files ready to be summarised

	go WalkDirectoryParallel(filepath.Clean("."), fileListQueue)
	go FileReaderWorker(fileListQueue, fileReadContentJobQueue)
	go FileProcessorWorker(fileReadContentJobQueue, fileSummaryJobQueue)

	results := []*FileJob{}
	reset := makeTimestampMilli()

	var spinLoc int
	update := true
	// NB this is not safe because results has no lock
	go func() {
		for update {
			// Every 100 ms redraw
			if makeTimestampMilli()-reset >= 100 {
				drawResults(results, textView, searchTerm, string(spinString[spinLoc]))
				reset = makeTimestampMilli()
				spinLoc++

				if spinLoc >= len(spinString) {
					spinLoc = 0
				}
			}

			if update {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	for res := range fileSummaryJobQueue {
		results = append(results, res)
	}

	update = false
	StopProcessing = true
	drawResults(results, textView, searchTerm, "")
}

var spinString = `\|/-`

func drawResults(results []*FileJob, textView *tview.TextView, searchTerm string, inProgress string) {
	RankResults(results)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if int64(len(results)) >= TotalCount {
		results = results[:TotalCount]
	}

	pResults := results
	if len(results) > 20 {
		pResults = results[:20]
	}

	var resultText string
	resultText += strconv.Itoa(len(results)) + " result(s) for '" + searchTerm + "' " + inProgress + "\n\n"

	for i, res := range pResults {
		resultText += fmt.Sprintf("[purple]%d. %s (%.3f)", i+1, res.Location, res.Score) + "[white]\n\n"

		// TODO need to escape the output https://godoc.org/github.com/rivo/tview#hdr-Colors
		var rel string
		if SnippetVersion == 1 {
			locations := []snippet.LocationType{}
			for k, v := range res.Locations {
				for _, i := range v {
					locations = append(locations, snippet.LocationType{
						Term:     k,
						Location: i,
					})
				}
			}

			sort.Slice(locations, func(i, j int) bool {
				return locations[i].Location < locations[j].Location
			})

			rel = snippet.ExtractRelevant2(string(res.Content), locations, int(SnippetLength), snippet.GetPrevCount(int(SnippetLength)), "…")
		} else {
			locs := []int{}
			for k := range res.Locations {
				locs = append(locs, res.Locations[k]...)
			}
			locs = RemoveIntDuplicates(locs)

			rel = snippet.ExtractRelevant(SearchString, string(res.Content), locs, int(SnippetLength), snippet.GetPrevCount(int(SnippetLength)), "…")
		}

		resultText += rel + "\n\n"
	}

	drawText(textView, resultText)
}

func drawText(textView *tview.TextView, text string) {
	textView.Clear()
	_, _ = fmt.Fprintf(textView, "%s", text)
	textView.ScrollToBeginning()
}

var searchSlice = []string{}
var searchMutex sync.Mutex
var searchSliceMutex sync.Mutex

func ProcessTui() {
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
			app.QueueUpdateDraw(func() {})
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

			go tuiSearch(app, textView)
		})

	grid.AddItem(inputField, 0, 0, 1, 3, 0, 1, false)
	grid.AddItem(textView, 1, 0, 1, 3, 0, 0, false)

	app.Draw()
	if err := app.SetRoot(grid, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
