// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package processor

import (
	"fmt"
	"github.com/boyter/cs/file"
	"github.com/boyter/cs/processor/snippet"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
			go cb(app, textView, item)
		}
	}
}

var IsCollecting = NewBool(false) // The state indicating if we are collecting results

func tuiSearch(app *tview.Application, textView *tview.TextView, searchTerm string) {
	// At this point we need to stop the background process that is running then wait for the
	// result collection to finish IE the part that collects results for display
	if IsWalking.IsSet() == true {
		TerminateWalking.SetTo(true)
	}

	for {
		time.Sleep(time.Millisecond * 50)
		if IsCollecting.IsSet() == false {
			break
		}
	}

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

	// If the user asks we should look back till we find the .git or .hg directory and start the search
	// or in case of SVN go back till we don't find it
	startDirectory := "."
	if FindRoot {
		startDirectory = file.FindRepositoryRoot(startDirectory)
	}

	go walkDirectory(startDirectory, fileListQueue)
	go FileReaderWorker(fileListQueue, fileReadContentJobQueue)
	go FileProcessorWorker(fileReadContentJobQueue, fileSummaryJobQueue)

	results := []*FileJob{}
	reset := makeTimestampMilli()

	var spinLocation int
	update := true
	spinString := `\|/-`

	// NB this is not safe because results has no lock
	go func() {
		for update {
			// Every 50 ms redraw
			if makeTimestampMilli()-reset >= 50 {
				drawResults(app, results, textView, searchTerm, string(spinString[spinLocation]))
				reset = makeTimestampMilli()
				spinLocation++

				if spinLocation >= len(spinString) {
					spinLocation = 0
				}
			}

			if update {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	IsCollecting.SetTo(true)
	defer IsCollecting.SetTo(false)
	for res := range fileSummaryJobQueue {
		results = append(results, res)
	}
	update = false
	drawResults(app, results, textView, searchTerm, "")
}

func drawResults(app *tview.Application, results []*FileJob, textView *tview.TextView, searchTerm string, inProgress string) {
	rankResults(SearchBytes, results)
	sortResults(results)

	if int64(len(results)) >= TotalCount {
		results = results[:TotalCount]
	}

	pResults := results
	if len(results) > 20 {
		pResults = results[:20]
	}

	var resultText string
	resultText += fmt.Sprintf("%d results(s) for '%s' from %d files %s\n\n", len(results), searchTerm, atomic.LoadInt64(&TotalCount), inProgress)

	for i, res := range pResults {
		resultText += fmt.Sprintf("[purple]%d. %s (%.3f)", i+1, res.Location, res.Score) + "[white]\n\n"

		// For debugging seeing the locations can be helpful
		//for key, val := range res.Locations {
		//	resultText += fmt.Sprintf("%s %d\n", key, val)
		//}
		//resultText += "\n"

		// TODO need to escape the output https://godoc.org/github.com/rivo/tview#hdr-Colors
		locations := GetResultLocations(res)
		coloredContent := snippet.WriteHighlights(res.Content, res.Locations, "[red]", "[white]")
		rel := snippet.ExtractRelevant(coloredContent, locations, int(SnippetLength), snippet.GetPrevCount(int(SnippetLength)), "…")

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

var textMutex sync.Mutex

const (
	SearchMode          string = " > search box"
	LocationExcludeMode string = " > location exclusion"
	ExtensionMode       string = " > extension filter ('go' 'go,java')"
	SnippetMode         string = " > snippet size selector"
	FuzzyMode           string = " > fuzzy search toggle"
	CaseSensitiveMode   string = " > case sensitive toggle"
	TextMode            string = " > text scroll"
)

// Param actually runs things which is only used for getting test coverage
func ProcessTui(run bool) {
	app := tview.NewApplication()

	var textView *tview.TextView
	var statusView *tview.InputField
	var searchInputField *tview.InputField
	var extensionInputField *tview.InputField
	var snippetInputField *tview.InputField
	var excludeInputField *tview.InputField
	var fuzzyCheckbox *tview.Checkbox
	var casesensitiveCheckbox *tview.Checkbox
	var lastSearch string

	eventChan := make(chan string)

	// For displaying status of where you are
	statusView = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorGreen).
		SetFieldTextColor(tcell.ColorBlack).
		SetText(SearchMode)

	textView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true).
		ScrollToBeginning().
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(searchInputField)
				statusView.SetText(SearchMode)
			case tcell.KeyBacktab:
				app.SetFocus(fuzzyCheckbox)
				statusView.SetText(FuzzyMode)
			}
		})

	snippetInputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault).
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
				app.SetFocus(casesensitiveCheckbox)
				statusView.SetText(CaseSensitiveMode)
			case tcell.KeyBacktab:
				app.SetFocus(extensionInputField)
				statusView.SetText(ExtensionMode)
			case tcell.KeyEnter:
				eventChan <- lastSearch
			case tcell.KeyUp:
				SnippetLength = min(SnippetLength+50, 8000)
				snippetInputField.SetText(strconv.Itoa(int(SnippetLength)))
				eventChan <- lastSearch
			case tcell.KeyPgUp:
				SnippetLength = min(SnippetLength+200, 8000)
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

	excludeInputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetText(strings.Join(LocationExcludePattern, ",")).
		SetFieldWidth(10).
		SetChangedFunc(func(text string) {
			text = strings.TrimSpace(text)

			t := []string{}
			for _, s := range strings.Split(text, ",") {
				if strings.TrimSpace(s) != "" {
					t = append(t, strings.TrimSpace(s))
				}
			}
			LocationExcludePattern = t

			eventChan <- lastSearch
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(extensionInputField)
				statusView.SetText(ExtensionMode)
			case tcell.KeyBacktab:
				app.SetFocus(searchInputField)
				statusView.SetText(SearchMode)
			}
		})

	extensionInputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabel(" ").
		SetLabelColor(tcell.ColorWhite).
		SetText(strings.Join(AllowListExtensions, ",")).
		SetFieldWidth(10).
		SetAcceptanceFunc(func(text string, c rune) bool {
			if c == ' ' {
				return false
			}

			return true
		}).
		SetChangedFunc(func(text string) {
			if strings.TrimSpace(text) == "" {
				AllowListExtensions = []string{}
			} else {
				AllowListExtensions = strings.Split(text, ",")
			}

			eventChan <- lastSearch
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(snippetInputField)
				statusView.SetText(SnippetMode)
			case tcell.KeyBacktab:
				app.SetFocus(excludeInputField)
				statusView.SetText(LocationExcludeMode)
			case tcell.KeyEnter:
				eventChan <- lastSearch
			}
		})

	searchInputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			textMutex.Lock()
			lastSearch = text
			textMutex.Unlock()
			eventChan <- lastSearch
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(excludeInputField)
				statusView.SetText(LocationExcludeMode)
			case tcell.KeyBacktab:
				app.SetFocus(textView)
				statusView.SetText(TextMode)
			case tcell.KeyEnter:
				eventChan <- lastSearch
			}
		})

	casesensitiveCheckbox = tview.NewCheckbox().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabel("").
		SetChecked(Fuzzy).
		SetChangedFunc(func(checked bool) {
			CaseSensitive = checked
			eventChan <- lastSearch
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(fuzzyCheckbox)
				statusView.SetText(FuzzyMode)
			case tcell.KeyBacktab:
				app.SetFocus(snippetInputField)
				statusView.SetText(SnippetMode)
			}
		})

	fuzzyCheckbox = tview.NewCheckbox().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabel("").
		SetChecked(Fuzzy).
		SetChangedFunc(func(checked bool) {
			Fuzzy = checked
			eventChan <- lastSearch
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(textView)
				statusView.SetText(TextMode)
			case tcell.KeyBacktab:
				app.SetFocus(casesensitiveCheckbox)
				statusView.SetText(CaseSensitiveMode)
			}
		})

	queryFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(searchInputField, 0, 8, false).
		AddItem(excludeInputField, 10, 0, false).
		AddItem(extensionInputField, 10, 0, false).
		AddItem(snippetInputField, 5, 1, false).
		AddItem(casesensitiveCheckbox, 1, 1, false).
		AddItem(fuzzyCheckbox, 1, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 2, 0, false).
		AddItem(textView, 0, 3, false).
		AddItem(statusView, 1, 0, false).
		AddItem(nil, 1, 0, false)

	// Start the debounce after everything else is setup and leave it running
	// forever in the background
	go debounce(time.Millisecond*50, eventChan, app, textView, tuiSearch)

	if run {
		if err := app.SetRoot(flex, true).SetFocus(searchInputField).Run(); err != nil {
			panic(err)
		}
	}
}
