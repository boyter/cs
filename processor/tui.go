package processor

import (
	"fmt"
	"github.com/boyter/sc/processor/snippet"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
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
		drawText(app, textView, "")
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

	directoryWalker := NewDirectoryWalker(fileListQueue)
	_ = directoryWalker.Walk(".")
	go directoryWalker.Run()
	go FileReaderWorker(fileListQueue, fileReadContentJobQueue)
	go FileProcessorWorker(fileReadContentJobQueue, fileSummaryJobQueue)

	results := []*FileJob{}
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
	StopProcessing = true
	drawResults(app, results, textView, searchTerm, "")
}

func drawResults(app *tview.Application, results []*FileJob, textView *tview.TextView, searchTerm string, inProgress string) {
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

	var textView *tview.TextView
	var dropdown *tview.DropDown
	var inputField *tview.InputField
	var lastSearch string

	textView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		ScrollToBeginning()

	dropdown = tview.NewDropDown().
		SetOptions([]string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900"}, nil).
		SetCurrentOption(3).
		SetLabelColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.Color16).
		SetSelectedFunc(func(text string, index int) {
			app.SetFocus(inputField)
			t, _ := strconv.Atoi(text)
			SnippetLength = int64(t)

			searchSliceMutex.Lock()
			searchSlice = append(searchSlice, strings.TrimSpace(lastSearch))
			searchSliceMutex.Unlock()

			go tuiSearch(app, textView)
		}).
		SetDoneFunc(func(key tcell.Key){
			switch key {
			case tcell.KeyTab:
				app.SetFocus(inputField)
			}
		})

	inputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			searchSliceMutex.Lock()
			searchSlice = append(searchSlice, strings.TrimSpace(text))
			lastSearch = text
			searchSliceMutex.Unlock()

			go tuiSearch(app, textView)
		}).
		SetDoneFunc(func(key tcell.Key){
			switch key {
			case tcell.KeyTab:
				app.SetFocus(dropdown)
			case tcell.KeyEnter:
				searchSliceMutex.Lock()
				searchSlice = append(searchSlice, strings.TrimSpace(lastSearch))
				searchSliceMutex.Unlock()

				go tuiSearch(app, textView)
			}
		})

	queryFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 8, false).
		AddItem(dropdown, 4, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 2, 0, false).
		AddItem(textView, 0, 3, false)

	if err := app.SetRoot(flex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
