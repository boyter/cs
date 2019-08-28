package processor

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

// Returns the current time as a millisecond timestamp
func makeTimestampMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func tuiSearch(app *tview.Application, textView *tview.TextView) {
	// Kill off anything else that's processing
	processor.StopProcessing = true
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
		drawText(app, textView, "")
		return
	}

	processor.SearchString = strings.Split(strings.TrimSpace(searchTerm), " ")
	processor.CleanSearchString()
	processor.TotalCount = 0

	// Enable processing again
	processor.StopProcessing = false

	fileListQueue := make(chan *processor.FileJob, runtime.NumCPU())           // Files ready to be read from disk
	fileReadContentJobQueue := make(chan *processor.FileJob, runtime.NumCPU()) // Files ready to be processed
	fileSummaryJobQueue := make(chan *processor.FileJob, runtime.NumCPU())     // Files ready to be summarised

	go processor.WalkDirectoryParallel(filepath.Clean("."), fileListQueue)
	go processor.FileReaderWorker(fileListQueue, fileReadContentJobQueue)
	go processor.FileProcessorWorker(fileReadContentJobQueue, fileSummaryJobQueue)

	results := []*processor.FileJob{}
	reset := makeTimestampMilli()

	var spinLoc int
	update := true
	spinString := `\|/-`

	// NB this is not safe because results has no lock
	go func() {
		for update {
			// Every 100 ms redraw
			if makeTimestampMilli()-reset >= 100 {
				drawResults(app, results, textView, searchTerm, string(spinString[spinLoc]))
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
	processor.StopProcessing = true
	drawResults(app, results, textView, searchTerm, "")
}

func drawResults(app *tview.Application, results []*processor.FileJob, textView *tview.TextView, searchTerm string, inProgress string) {
	RankResults(results)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return strings.Compare(results[i].Location, results[j].Location) < 0
		}

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
		locations := GetResultLocations(res)
		rel := snippet.ExtractRelevant(string(res.Content), locations, int(SnippetLength), snippet.GetPrevCount(int(SnippetLength)), "…")
		resultText += rel + "\n\n"
	}

	drawText(app, textView, resultText)
}

func drawText(app *tview.Application, textView *tview.TextView, text string) {
	app.QueueUpdateDraw(func() {
		textView.Clear()
		_, err := fmt.Fprintf(textView, "%s", text)

		if err != nil {
			return
		}

		textView.ScrollToBeginning()
	})
}

var searchSlice = []string{}
var searchMutex sync.Mutex
var searchSliceMutex sync.Mutex

func ProcessTui() {
	app := tview.NewApplication()

	grid := tview.NewGrid().
		SetRows(3).
		SetColumns(1).
		SetBorders(false)

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		ScrollToBeginning()

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

	form := tview.NewForm()
	form.AddInputField("> ", "", 0, nil, func(text string) {
		searchSliceMutex.Lock()
		searchSlice = append(searchSlice, strings.TrimSpace(text))
		searchSliceMutex.Unlock()

		go tuiSearch(app, textView)
	}).
		AddDropDown("", []string{"100", "200", "300", "400", "500"}, 2, func(text string, index int) {
			v, _ := strconv.Atoi(text)
			processor.SnippetLength = int64(v)

			go tuiSearch(app, textView)
		})

	grid.AddItem(inputField, 0, 0, 1, 3, 0, 0, false)
	grid.AddItem(form, 1, 0, 1, 3, 0, 0, true)
	grid.AddItem(textView, 2, 0, 1, 3, 0, 0, false)

	app.Draw()

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
