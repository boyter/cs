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

	str "github.com/boyter/cs/string"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// Simple debounce function allowing us to wait on user input slightly
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

// TODO replace this with bool + mutex
var IsCollecting = NewBool(false) // The state indicating if we are collecting results

// Variables we need to keep around between searches, but are recreated on each new one
var tuiFileWalker file.FileWalker
var tuiFileReaderWorker FileReaderWorker2
var tuiSearcherWorker SearcherWorker

// If we are here that means we actually need to perform a search
func tuiSearch(app *tview.Application, textView *tview.TextView, searchTerm string) {
	// At this point we need to stop the background process that is running then wait for the
	// result collection to finish IE the part that collects results for display
	if tuiFileWalker.Walking() {
		tuiFileWalker.Terminate()
	}

	// The walker has stopped which means eventually the pipeline
	// should flush and the channel will be closed but until then
	// loop forever checking waiting for this to happen
	for {
		time.Sleep(time.Millisecond * 10)
		if IsCollecting.IsSet() == false {
			break
		}
	}

	// If the searchterm is empty then we draw out nothing and return
	if strings.TrimSpace(searchTerm) == "" {
		drawText(app, textView, "")
		return
	}

	SearchString = strings.Split(strings.TrimSpace(searchTerm), " ")
	TotalCount = 0

	fileQueue := make(chan *file.File)                      // NB unbuffered because we want the UI to respond and this is what causes affects
	toProcessQueue := make(chan *fileJob, runtime.NumCPU()) // Files to be read into memory for processing
	summaryQueue := make(chan *fileJob, runtime.NumCPU())   // Files that match and need to be displayed

	// If the user asks we should look back till we find the .git or .hg directory and start the search
	// or in case of SVN go back till we don't find it
	startDirectory := "."
	if FindRoot {
		startDirectory = file.FindRepositoryRoot(startDirectory)
	}

	tuiFileWalker = file.NewFileWalker(startDirectory, fileQueue)
	tuiFileWalker.EnableIgnoreFile = true
	tuiFileWalker.PathExclude = PathDenylist

	tuiFileReaderWorker = NewFileReaderWorker(fileQueue, toProcessQueue)

	tuiSearcherWorker = NewSearcherWorker(toProcessQueue, summaryQueue)
	tuiSearcherWorker.SearchString = SearchString
	tuiSearcherWorker.MatchLimit = 100

	go tuiFileWalker.Start()
	go tuiFileReaderWorker.Start()
	go tuiSearcherWorker.Start()

	// Updated with results as we get them NB this is
	// painted as we go TODO add lock for access to this
	results := []*fileJob{}

	// Counts when we last painted on the screen
	reset := makeTimestampMilli()

	// Used to display a spinner indicating a search is happening
	var spinLocation int
	update := true
	spinString := `\|/-`

	// NB this is not safe because results has no lock
	go func() {
		for update {
			// Every 50 ms redraw the current set of results
			if makeTimestampMilli()-reset >= 50 {
				drawResults(app, results, textView, searchTerm, string(spinString[spinLocation]))
				reset = makeTimestampMilli()
				spinLocation++

				if spinLocation >= len(spinString) {
					spinLocation = 0
				}
			}

			time.Sleep(5 * time.Millisecond)
		}
	}()

	IsCollecting.SetTo(true)
	defer IsCollecting.SetTo(false)
	for res := range summaryQueue {
		results = append(results, res)
	}
	update = false
	drawResults(app, results, textView, searchTerm, "")
}

func drawResults(app *tview.Application, results []*fileJob, textView *tview.TextView, searchTerm string, inProgress string) {
	rankResults2(100, results)

	//if int64(len(results)) >= TotalCount {
	//	results = results[:TotalCount]
	//}

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

		// Combine all the locations such that we can highlight correctly
		l := [][]int{}
		for _, value := range res.MatchLocations {
			l = append(l, value...)
		}

		// TODO need to escape the output https://godoc.org/github.com/rivo/tview#hdr-Colors
		coloredContent := str.HighlightString(string(res.Content), l, "[red]", "[white]")
		relevant, _, _ := str.ExtractRelevant(coloredContent, l, int(SnippetLength), snippet.GetPrevCount(int(SnippetLength)), "…")

		resultText += relevant + "\n\n"
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
