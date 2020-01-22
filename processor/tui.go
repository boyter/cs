package processor

import (
	"fmt"
	"github.com/boyter/cs/processor/printer"
	"github.com/boyter/cs/processor/snippet"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)


func debounce(interval time.Duration, input chan string, app *tview.Application, textView *tview.TextView, cb func(app *tview.Application, textView *tview.TextView, arg string)) {
	var item string
	timer := time.NewTimer(interval)
	for {
		select {
		case item = <-input:
			timer.Reset(interval)
		case <-timer.C:
			if item != "" {
				go cb(app, textView, item)
			}
		}
	}
}

func tuiSearch(app *tview.Application, textView *tview.TextView, searchTerm string) {

	if strings.TrimSpace(searchTerm) == "" {
		drawText(app, textView, "")
		return
	}

	SearchString = strings.Split(strings.TrimSpace(searchTerm), " ")
	CleanSearchString()
	TotalCount = 0

	fileListQueue := make(chan *FileJob, runtime.NumCPU())           // Files ready to be read from disk
	fileReadContentJobQueue := make(chan *FileJob, runtime.NumCPU()) // Files ready to be processed
	fileSummaryJobQueue := make(chan *FileJob, runtime.NumCPU())     // Files ready to be summarised

	go walkDirectorySimple(fileListQueue)
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
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	for res := range fileSummaryJobQueue {
		results = append(results, res)
	}

	update = false
	drawResults(app, results, textView, searchTerm, "")
}

func drawResults(app *tview.Application, results []*FileJob, textView *tview.TextView, searchTerm string, inProgress string) {
	RankResults(results)
	SortResults(results)

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
		coloredContent := printer.WriteColored(res.Content, res.Locations, "[red]", "[white]")
		rel := snippet.ExtractRelevant(coloredContent, locations, int(SnippetLength), snippet.GetPrevCount(int(SnippetLength)), "…")

		resultText += rel + "\n\n"
	}

	drawText(app, textView, resultText)
}

func drawText(app *tview.Application, textView *tview.TextView, text string) {
	app.QueueUpdateDraw(func() {
		textView.Clear()
		// for debugging
		//stack := debug.Stack()
		//_, err := fmt.Fprintf(textView, strconv.Itoa(runtime.NumGoroutine())+" %s", string(stack))

		// usual happy path
		_, err := fmt.Fprintf(textView, "%s", text)

		if err != nil {
			return
		}

		textView.ScrollToBeginning()
	})
}

var textMutex sync.Mutex

// Param actually runs things which is only used for getting test coverage
func ProcessTui(run bool) {
	app := tview.NewApplication()

	var textView *tview.TextView
	var inputField *tview.InputField
	var extInputField *tview.InputField
	var snippetInputField *tview.InputField
	var lastSearch string

	eventChan := make(chan string)

	textView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		ScrollToBeginning()

	snippetInputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetAcceptanceFunc(tview.InputFieldInteger).
		SetText(strconv.Itoa(int(SnippetLength))).
		SetFieldWidth(4).
		SetChangedFunc(func(text string) {
			if strings.TrimSpace(text) == "" {
				SnippetLength = 300 // default
			} else {
				t, _ := strconv.Atoi(text)
				if t == 0 {
					SnippetLength = 300
				} else {
					SnippetLength = int64(t)
				}
			}

			eventChan <- lastSearch
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(inputField)
			case tcell.KeyBacktab:
				app.SetFocus(extInputField)
			case tcell.KeyEnter:
				eventChan <- lastSearch
			case tcell.KeyUp:
				SnippetLength = min(SnippetLength+50, 2000)
				snippetInputField.SetText(strconv.Itoa(int(SnippetLength)))
				eventChan <- lastSearch
			case tcell.KeyPgUp:
				SnippetLength = min(SnippetLength+200, 2000)
				snippetInputField.SetText(strconv.Itoa(int(SnippetLength)))
				eventChan <- lastSearch
			case tcell.KeyDown:
				SnippetLength = max(50, SnippetLength-50)
				snippetInputField.SetText(strconv.Itoa(int(SnippetLength)))
				eventChan <- lastSearch
			case tcell.KeyPgDn:
				SnippetLength = max(50, SnippetLength-200)
				snippetInputField.SetText(strconv.Itoa(int(SnippetLength)))
				eventChan <- lastSearch
			}
		})

	extInputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabelColor(tcell.ColorWhite).
		SetText(strings.Join(WhiteListExtensions, ",")).
		SetFieldWidth(10).
		SetAcceptanceFunc(func(text string, c rune) bool {
			if c == ' ' {
				return false
			}

			return true
		}).
		SetChangedFunc(func(text string) {
			if strings.TrimSpace(text) == "" {
				WhiteListExtensions = []string{}
			} else {
				WhiteListExtensions = strings.Split(text, ",")
			}

			eventChan <- lastSearch
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(snippetInputField)
			case tcell.KeyBacktab:
				app.SetFocus(inputField)
			case tcell.KeyEnter:
				eventChan <- lastSearch
			}
		})

	inputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			textMutex.Lock()
			lastSearch = text
			textMutex.Unlock()
			eventChan <- text

			if strings.TrimSpace(text) == "" {
				drawText(app, textView, "")
			}
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(extInputField)
			case tcell.KeyBacktab:
				app.SetFocus(snippetInputField)
			case tcell.KeyEnter:
				eventChan <- lastSearch
			}
		})

	queryFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 8, false).
		AddItem(extInputField, 10, 0, false).
		AddItem(snippetInputField, 5, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 2, 0, false).
		AddItem(textView, 0, 3, false)

	// Start the debounce after everything else is setup and leave it running
	// forever in the background
	go debounce(time.Millisecond*100, eventChan, app, textView, tuiSearch)

	if run {
		if err := app.SetRoot(flex, true).SetFocus(inputField).Run(); err != nil {
			panic(err)
		}
	}
}
