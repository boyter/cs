// SPDX-License-Identifier: MIT OR Unlicense

package processor

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/boyter/cs/file"
	"github.com/boyter/cs/str"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type displayResult struct {
	Title      *tview.TextView
	Body       *tview.TextView
	BodyHeight int
	SpacerOne  *tview.TextView
	SpacerTwo  *tview.TextView
	Location   string
}

type codeResult struct {
	Title    string
	Content  string
	Score    float64
	Location string
}

type tuiApplicationController struct {
	Query               string
	Queries             []string
	Sync                sync.Mutex
	Changed             bool
	Running             bool
	Offset              int
	Results             []*FileJob
	TuiFileWalker       *file.FileWalker
	TuiFileReaderWorker *FileReaderWorker
	TuiSearcherWorker   *SearcherWorker

	// View requirements
	SpinString   string
	SpinLocation int
	SpinRun      int
}

func (cont *tuiApplicationController) SetQuery(q string) {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	cont.Queries = append(cont.Queries, q)
}

func (cont *tuiApplicationController) IncrementOffset() {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	cont.Offset++
}

func (cont *tuiApplicationController) DecrementOffset() {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	if cont.Offset != 0 {
		cont.Offset--
	}
}

func (cont *tuiApplicationController) ResetOffset() {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	cont.Offset = 0
}

func (cont *tuiApplicationController) GetOffset() int {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	return cont.Offset
}

func (cont *tuiApplicationController) SetChanged(b bool) {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	cont.Changed = b
}

func (cont *tuiApplicationController) GetChanged() bool {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	return cont.Changed
}

func (cont *tuiApplicationController) SetRunning(b bool) {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	cont.Running = b
}

func (cont *tuiApplicationController) GetRunning() bool {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	return cont.Running
}

func (cont *tuiApplicationController) Search(s string) {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	cont.Query = s
}

// After any change is made that requires something drawn on the screen this is the method that does
func (cont *tuiApplicationController) drawView() {
	cont.Sync.Lock()
	if !cont.Changed || cont.TuiFileReaderWorker == nil {
		cont.Sync.Unlock()
		return
	}
	cont.Sync.Unlock()

	// reset the elements by clearing out every one
	tviewApplication.QueueUpdateDraw(func() {
		for _, t := range displayResults {
			t.Title.SetText("")
			t.Body.SetText("")
			t.SpacerOne.SetText("")
			t.SpacerTwo.SetText("")
			resultsFlex.ResizeItem(t.Body, 0, 0)
		}
	})

	cont.Sync.Lock()
	resultsCopy := make([]*FileJob, len(cont.Results))
	copy(resultsCopy, cont.Results)
	cont.Sync.Unlock()

	// rank all results
	// then go and get the relevant portion for display
	rankResults(int(cont.TuiFileReaderWorker.GetFileCount()), resultsCopy)
	documentTermFrequency := calculateDocumentTermFrequency(resultsCopy)

	// after ranking only get the details for as many as we actually need to
	// cut down on processing
	if len(resultsCopy) > len(displayResults) {
		resultsCopy = resultsCopy[:len(displayResults)]
	}

	// We use this to swap out the highlights after we escape to ensure that we don't escape
	// out own colours
	md5Digest := md5.New()
	fmtBegin := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("begin_%d", makeTimestampNano()))))
	fmtEnd := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("end_%d", makeTimestampNano()))))

	// go and get the codeResults the user wants to see using selected as the offset to display from
	var codeResults []codeResult
	for i, res := range resultsCopy {
		if i >= cont.Offset {
			snippets := extractRelevantV3(res, documentTermFrequency, int(SnippetLength), "…")[0]

			// now that we have the relevant portion we need to get just the bits related to highlight it correctly
			// which this method does. It takes in the snippet, we extract and all of the locations and then returns just
			l := getLocated(res, snippets)
			coloredContent := str.HighlightString(snippets.Content, l, fmtBegin, fmtEnd)
			coloredContent = tview.Escape(coloredContent)

			coloredContent = strings.Replace(coloredContent, fmtBegin, "[red]", -1)
			coloredContent = strings.Replace(coloredContent, fmtEnd, "[white]", -1)

			codeResults = append(codeResults, codeResult{
				Title:    res.Location,
				Content:  coloredContent,
				Score:    res.Score,
				Location: res.Location,
			})
		}
	}

	// render out what the user wants to see based on the results that have been chosen
	tviewApplication.QueueUpdateDraw(func() {
		for i, t := range codeResults {
			displayResults[i].Title.SetText(fmt.Sprintf("[fuchsia]%s (%f)[-:-:-]", t.Title, t.Score))
			displayResults[i].Body.SetText(t.Content)
			displayResults[i].Location = t.Location

			//we need to update the item so that it displays everything we have put in
			resultsFlex.ResizeItem(displayResults[i].Body, len(strings.Split(t.Content, "\n")), 0)
		}
	})

	// we can only set that nothing
	cont.Changed = false
}

func (cont *tuiApplicationController) doSearch() {
	cont.Sync.Lock()
	//// deal with the user clearing out the search
	if len(cont.Queries) == 0 {
		cont.Sync.Unlock()
		return
	}

	// only process the last search
	cont.Query = cont.Queries[len(cont.Queries)-1]
	query := cont.Query
	cont.Queries = []string{}
	cont.Sync.Unlock()

	if query == "" {
		cont.Sync.Lock()
		cont.Results = []*FileJob{}
		cont.Changed = true
		cont.Sync.Unlock()
		return
	}

	// if we have a walker that's currently walking terminate it
	if cont.TuiFileWalker != nil && (cont.TuiFileWalker.Walking() || cont.GetRunning()) {
		cont.TuiFileWalker.Terminate()

		// wait for the current walker to stop
		for cont.TuiFileWalker.Walking() || cont.GetRunning() {
			time.Sleep(1 * time.Millisecond)
		}
	}

	fileQueue := make(chan *file.File)                      // NB unbuffered because we want to be able to cancel walking and have the UI update
	toProcessQueue := make(chan *FileJob, runtime.NumCPU()) // Files to be read into memory for processing
	summaryQueue := make(chan *FileJob, runtime.NumCPU())   // Files that match and need to be displayed

	cont.TuiFileWalker = file.NewFileWalker(".", fileQueue)
	cont.TuiFileWalker.IgnoreIgnoreFile = IgnoreIgnoreFile
	cont.TuiFileWalker.IgnoreGitIgnore = IgnoreGitIgnore
	cont.TuiFileWalker.IncludeHidden = IncludeHidden
	cont.TuiFileWalker.PathExclude = PathDenylist
	cont.TuiFileWalker.AllowListExtensions = AllowListExtensions
	cont.TuiFileWalker.InstanceId = instanceCount
	cont.TuiFileWalker.LocationExcludePattern = LocationExcludePattern

	cont.TuiFileReaderWorker = NewFileReaderWorker(fileQueue, toProcessQueue)
	cont.TuiFileReaderWorker.InstanceId = instanceCount
	cont.TuiFileReaderWorker.SearchPDF = SearchPDF
	cont.TuiFileReaderWorker.MaxReadSizeBytes = MaxReadSizeBytes

	cont.TuiSearcherWorker = NewSearcherWorker(toProcessQueue, summaryQueue)
	cont.TuiSearcherWorker.SearchString = strings.Split(query, " ")
	cont.TuiSearcherWorker.MatchLimit = -1 // NB this can make things slow because we keep going
	cont.TuiSearcherWorker.InstanceId = instanceCount
	cont.TuiSearcherWorker.IncludeBinary = IncludeBinaryFiles
	cont.TuiSearcherWorker.CaseSensitive = CaseSensitive
	cont.TuiSearcherWorker.IncludeMinified = IncludeMinified
	cont.TuiSearcherWorker.MinifiedLineByteLength = MinifiedLineByteLength

	go cont.TuiFileWalker.Start()
	go cont.TuiFileReaderWorker.Start()
	go cont.TuiSearcherWorker.Start()

	// Updated with results as we get them NB this is
	// painted as we go
	var results []*FileJob
	var resultsMutex sync.Mutex
	cont.SetRunning(true)

	go func() {
		for cont.GetRunning() {
			// Every 50 ms redraw the current set of results
			resultsMutex.Lock()
			cont.Sync.Lock()
			cont.Results = results
			cont.Changed = true
			cont.Sync.Unlock()
			resultsMutex.Unlock()

			time.Sleep(50 * time.Millisecond)
		}
	}()

	for res := range summaryQueue {
		resultsMutex.Lock()
		results = append(results, res)
		resultsMutex.Unlock()
	}
	// once we get out of the collection we can indicate we are not running anymore
	cont.SetRunning(false)

	cont.Sync.Lock()
	cont.Changed = true
	cont.Results = results
	cont.Sync.Unlock()
}

func (cont *tuiApplicationController) updateView() {
	// render loop running background is the only thing responsible for updating the results based on the state
	// in the applicationController
	go func() {
		for {
			cont.drawView()
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

func (cont *tuiApplicationController) updateStatus() {
	// render loop running background is the only thing responsible for updating the results based on the state
	// in the applicationController
	go func() {
		for {
			cont.Sync.Lock()
			if cont.TuiFileWalker != nil {

				plural := "s"
				if len(cont.Results) == 1 {
					plural = ""
				}

				status := fmt.Sprintf("%d result%s for '%s' from %d files", len(cont.Results), plural, cont.Query, cont.TuiFileReaderWorker.GetFileCount())
				if cont.Running {
					status = fmt.Sprintf("%d result%s for '%s' from %d files %s", len(cont.Results), plural, cont.Query, cont.TuiFileReaderWorker.GetFileCount(), string(cont.SpinString[cont.SpinLocation]))

					cont.SpinRun++
					if cont.SpinRun == 4 {
						cont.SpinLocation++
						if cont.SpinLocation >= len(cont.SpinString) {
							cont.SpinLocation = 0
						}
						cont.SpinRun = 0
						cont.Changed = true
					}
				}

				tviewApplication.QueueUpdateDraw(func() {
					statusView.SetText(status)
				})
			}
			cont.Sync.Unlock()

			time.Sleep(30 * time.Millisecond)
		}
	}()
}

func (cont *tuiApplicationController) processSearch() {
	// we only ever want to have one search trigger at a time which is what this does
	// searches come in... we trigger them to run
	go func() {
		for {
			cont.doSearch()
			time.Sleep(15 * time.Millisecond)
		}
	}()
}

// Sets up all of the UI components we need to actually display
var overallFlex *tview.Flex
var inputField *tview.InputField
var queryFlex *tview.Flex
var resultsFlex *tview.Flex
var statusView *tview.TextView
var displayResults []displayResult
var tviewApplication *tview.Application
var snippetInputField *tview.InputField
var excludeInputField *tview.InputField

func NewTuiApplication() {
	tviewApplication = tview.NewApplication()
	applicationController := tuiApplicationController{
		Sync:       sync.Mutex{},
		SpinString: `\|/-`,
	}

	// Create the elements we use to display the code results here
	for i := 1; i < 50; i++ {
		var textViewTitle *tview.TextView
		var textViewBody *tview.TextView

		textViewTitle = tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			ScrollToBeginning()

		textViewBody = tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			ScrollToBeginning()

		displayResults = append(displayResults, displayResult{
			Title:      textViewTitle,
			Body:       textViewBody,
			BodyHeight: -1,
			SpacerOne:  tview.NewTextView(),
			SpacerTwo:  tview.NewTextView(),
		})
	}

	// input field which deals with the user input for the main search which ultimately triggers a search
	inputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(0).
		SetDoneFunc(func(key tcell.Key) {
			// this deals with the keys that trigger "done" functions such as up/down/enter
			switch key {
			case tcell.KeyEnter:
				tviewApplication.Stop()
				// we want to work like fzf for piping into other things hence print out the selected version
				if len(applicationController.Results) != 0 {
					fmt.Println(displayResults[applicationController.GetOffset()].Location)
				}
				os.Exit(0)
			case tcell.KeyTab:
				tviewApplication.SetFocus(excludeInputField)
			case tcell.KeyBacktab:
				tviewApplication.SetFocus(snippetInputField)
			case tcell.KeyUp:
				applicationController.DecrementOffset()
				applicationController.SetChanged(true)
			case tcell.KeyDown:
				applicationController.IncrementOffset()
				applicationController.SetChanged(true)
			case tcell.KeyESC:
				tviewApplication.Stop()
				os.Exit(0)
			}
		}).
		SetChangedFunc(func(text string) {
			// after the text has changed set the query so we can trigger a search
			text = strings.TrimSpace(text)
			applicationController.ResetOffset()
			applicationController.SetQuery(text)
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
			triggerSearch(&applicationController)
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				tviewApplication.SetFocus(snippetInputField)
			case tcell.KeyBacktab:
				tviewApplication.SetFocus(inputField)
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
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				tviewApplication.SetFocus(inputField)
			case tcell.KeyBacktab:
				tviewApplication.SetFocus(excludeInputField)
			case tcell.KeyEnter:
			case tcell.KeyUp:
				SnippetLength = min(SnippetLength+100, 8000)
				triggerSearch(&applicationController)
			case tcell.KeyPgUp:
				SnippetLength = min(SnippetLength+200, 8000)
				triggerSearch(&applicationController)
			case tcell.KeyDown:
				SnippetLength = max(100, SnippetLength-100)
				triggerSearch(&applicationController)
			case tcell.KeyPgDn:
				SnippetLength = max(100, SnippetLength-200)
				triggerSearch(&applicationController)
			}
		})

	statusView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		ScrollToBeginning()

	// setup the flex containers to have everything rendered neatly
	queryFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 8, false).
		AddItem(excludeInputField, 30, 0, false).
		AddItem(snippetInputField, 5, 1, false)

	resultsFlex = tview.NewFlex().SetDirection(tview.FlexRow)

	overallFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(statusView, 1, 0, false).
		AddItem(resultsFlex, 0, 1, false)

	// Add all of the display codeResults into the container ready to be populated
	for _, t := range displayResults {
		resultsFlex.AddItem(t.SpacerOne, 1, 0, false)
		resultsFlex.AddItem(t.Title, 1, 0, false)
		resultsFlex.AddItem(t.SpacerTwo, 1, 0, false)
		resultsFlex.AddItem(t.Body, t.BodyHeight, 1, false)
	}

	// trigger the first render without user action
	applicationController.SetChanged(true)

	// trigger the jobs to start running things
	applicationController.updateView()
	applicationController.processSearch()
	applicationController.updateStatus()

	if err := tviewApplication.SetRoot(overallFlex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}

func triggerSearch(applicationController *tuiApplicationController) {
	snippetInputField.SetText(strconv.Itoa(int(SnippetLength)))
	applicationController.ResetOffset()
	applicationController.SetQuery(strings.TrimSpace(inputField.GetText()))
}
