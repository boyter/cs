// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package processor

import (
	"fmt"
	"github.com/boyter/cs/file"
	str "github.com/boyter/cs/str"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var isRunningMutex sync.Mutex
var resultsMutex sync.Mutex

// Variables we need to keep around between searches, but are recreated on each new one
var tuiFileWalker *file.FileWalker
var tuiFileReaderWorker *FileReaderWorker
var tuiSearcherWorker *SearcherWorker
var instanceCount int

// Used to show that things are happening, and can be modified to whatever is required
var spinString = `\|/-`

var debugCount = 0

type searchTermStruct struct {
	SearchTerm string
	TimeStamp  int64
}

var queuedSearch []searchTermStruct
var queuedSearchMutex sync.Mutex
var queuedSearchLastTime int64

// If we are here that means we actually need to perform a search
func tuiSearch(app *tview.Application, textView *tview.TextView, searchTerm searchTermStruct) {
	queuedSearchMutex.Lock()
	queuedSearch = append(queuedSearch, searchTerm)
	queuedSearchMutex.Unlock()

	// At this point we need to stop the background process that is running then wait for the
	// result collection to finish IE the part that collects results for display
	if tuiFileWalker != nil {
		tuiFileWalker.Terminate()
	}

	// TODO still a race condition here to need to resolve as we call terminate multiple times
	// We lock here because we don't want another instance to run until
	// this one has terminated which should happen with the terminate call
	// NB at this point we have a race condition... many searches are wanting to run in here
	// but of course only one gets the lock, which might not be the most recent one...
	isRunningMutex.Lock()
	defer isRunningMutex.Unlock()

	// At this point we want to
	queuedSearchMutex.Lock()
	if queuedSearchLastTime > searchTerm.TimeStamp {
		queuedSearchMutex.Unlock()
		return
	}

	var search string
	for _, s := range queuedSearch {
		if s.TimeStamp > queuedSearchLastTime {
			search = s.SearchTerm
			queuedSearchLastTime = s.TimeStamp
		}
	}

	queuedSearch = []searchTermStruct{}
	queuedSearchMutex.Unlock()

	// if the search is not mine then return because there is something better to do
	if queuedSearchLastTime != searchTerm.TimeStamp && search != searchTerm.SearchTerm {
		return
	}

	// If the searchterm is empty then we draw out nothing and return
	if strings.TrimSpace(search) == "" {
		drawText(app, textView, "")
		return
	}

	SearchString = strings.Split(strings.TrimSpace(search), " ")

	// If the user asks we should look back till we find the .git or .hg directory and start the search
	// or in case of SVN go back till we don't find it
	startDirectory := "."
	if FindRoot {
		startDirectory = file.FindRepositoryRoot(startDirectory)
	}

	fileQueue := make(chan *file.File)                      // NB unbuffered because we want the UI to respond and this is what causes affects
	toProcessQueue := make(chan *fileJob, runtime.NumCPU()) // Files to be read into memory for processing
	summaryQueue := make(chan *fileJob, runtime.NumCPU())   // Files that match and need to be displayed

	tuiFileWalker = file.NewFileWalker(startDirectory, fileQueue)
	tuiFileWalker.IgnoreIgnoreFile = IgnoreIgnoreFile
	tuiFileWalker.IgnoreGitIgnore = IgnoreGitIgnore
	tuiFileWalker.IncludeHidden = IncludeHidden
	tuiFileWalker.PathExclude = PathDenylist
	tuiFileWalker.AllowListExtensions = AllowListExtensions
	tuiFileWalker.InstanceId = instanceCount
	tuiFileWalker.LocationExcludePattern = LocationExcludePattern
	tuiFileWalker.UniqueId = search

	tuiFileReaderWorker = NewFileReaderWorker(fileQueue, toProcessQueue)
	tuiFileReaderWorker.InstanceId = instanceCount
	tuiFileReaderWorker.SearchPDF = SearchPDF
	tuiFileReaderWorker.MaxReadSizeBytes = MaxReadSizeBytes

	tuiSearcherWorker = NewSearcherWorker(toProcessQueue, summaryQueue)
	tuiSearcherWorker.SearchString = SearchString
	tuiSearcherWorker.MatchLimit = -1 // NB this can make things slow because we keep going
	tuiSearcherWorker.InstanceId = instanceCount
	tuiSearcherWorker.IncludeBinary = IncludeBinaryFiles
	tuiSearcherWorker.CaseSensitive = CaseSensitive
	tuiSearcherWorker.IncludeMinified = IncludeMinified
	tuiSearcherWorker.MinifiedLineByteLength = MinifiedLineByteLength

	instanceCount++

	go tuiFileWalker.Start()
	go tuiFileReaderWorker.Start()
	go tuiSearcherWorker.Start()

	// Updated with results as we get them NB this is
	// painted as we go
	var results []*fileJob

	// Used to display a spinner indicating a search is happening
	var spinLocation int
	update := true

	go func() {
		for update {
			// Every 50 ms redraw the current set of results
			resultsMutex.Lock()
			drawResults(app, results, textView, search, tuiFileReaderWorker.GetFileCount(), string(spinString[spinLocation]))
			resultsMutex.Unlock()
			spinLocation++

			if spinLocation >= len(spinString) {
				spinLocation = 0
			}

			time.Sleep(50 * time.Millisecond)
		}
	}()

	for res := range summaryQueue {
		resultsMutex.Lock()
		results = append(results, res)
		resultsMutex.Unlock()
	}

	update = false
	resultsMutex.Lock()
	drawResults(app, results, textView, search, tuiFileReaderWorker.GetFileCount(), "")
	resultsMutex.Unlock()
	debugCount++
}

func drawResults(app *tview.Application, results []*fileJob, textView *tview.TextView, searchTerm string, fileCount int64, inProgress string) {
	rankResults(int(fileCount), results)

	// TODO this should not be hardcoded
	pResults := results
	if len(results) > 20 {
		pResults = results[:20]
	}

	var resultText string
	resultText += fmt.Sprintf("%d results(s) for '%s' from %d files %s\n\n", len(results), searchTerm, fileCount, inProgress)

	documentTermFrequency := calculateDocumentTermFrequency(results)
	for i, res := range pResults {
		// NB this just gets the first snippet which should in theory be the most relevant
		v3 := extractRelevantV3(res, documentTermFrequency, int(SnippetLength), "…")[0]

		resultText += fmt.Sprintf("[purple]%d. %s (%.3f)", i+1, res.Location, res.Score) + "[white]\n\n"

		// now that we have the relevant portion we need to get just the bits related to highlight it correctly
		// which this method does. It takes in the snippet, we extract and all of the locations and then returns just
		l := getLocated(res, v3)

		coloredContent := str.HighlightString(v3.Content, l, "[red]", "[white]")
		resultText += coloredContent + "\n\n"
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
	SearchMode          string = "(search box)"
	LocationExcludeMode string = "(location exclusion)"
	SnippetMode         string = "(snippet size selector)"
	TextMode            string = "(text scroll)"
)

// Param actually runs things which is only used for getting test coverage
func ProcessTui(run bool) {
	app := tview.NewApplication()

	var textView *tview.TextView
	var statusView *tview.InputField
	var searchInputField *tview.InputField
	var snippetInputField *tview.InputField
	var excludeInputField *tview.InputField
	var lastSearch string

	eventChan := make(chan string)

	// For displaying mode IE are we searching, or filtering locations or changing snippet size
	statusView = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetText(SearchMode)

	// This is where results are actually displayed
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
				app.SetFocus(snippetInputField)
				statusView.SetText(SnippetMode)
			}
		})

	// This is used to allow filtering out of paths and the like
	excludeInputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetText(strings.Join(LocationExcludePattern, ",")).
		SetFieldWidth(30).
		SetChangedFunc(func(text string) {
			text = strings.TrimSpace(text)

			var t []string
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
				app.SetFocus(snippetInputField)
				statusView.SetText(SnippetMode)
			case tcell.KeyBacktab:
				app.SetFocus(searchInputField)
				statusView.SetText(SearchMode)
			}
		})

	// Decide how large a snippet we should be displaying
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
				app.SetFocus(textView)
				statusView.SetText(TextMode)
			case tcell.KeyBacktab:
				app.SetFocus(excludeInputField)
				statusView.SetText(LocationExcludeMode)
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

	// Where the search actually happens
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

	// top flex container which holds the controls such as the text, the
	queryFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(searchInputField, 0, 8, false).
		AddItem(excludeInputField, 30, 0, false).
		AddItem(snippetInputField, 5, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 2, 0, false).
		AddItem(statusView, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(textView, 0, 3, false)

	go func() {
		for i := range eventChan {
			go tuiSearch(app, textView, searchTermStruct{
				SearchTerm: i,
				TimeStamp:  makeTimestampNano(),
			})
		}
	}()

	if run {
		if err := app.SetRoot(flex, true).SetFocus(searchInputField).Run(); err != nil {
			panic(err)
		}
	}
}
