// SPDX-License-Identifier: MIT OR Unlicense

package processor

import (
	"fmt"
	"github.com/boyter/cs/file"
	"github.com/boyter/cs/str"

	//"github.com/boyter/cs/str"
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
}

type codeResult struct {
	Title   string
	Content string
	Score   float64
}

type tuiApplicationController struct {
	Query               string
	Count               int
	Sync                sync.Mutex
	Changed             bool
	Running             bool
	Offset              int
	Results             []*FileJob
	TuiFileWalker       *file.FileWalker
	TuiFileReaderWorker *FileReaderWorker
	TuiSearcherWorker   *SearcherWorker

	// View requirements
	SpinString       string
	SpinLocation     int
	SpinRun int
}

func (cont *tuiApplicationController) SetQuery(q string) {
	cont.Sync.Lock()
	defer cont.Sync.Unlock()
	cont.Query = q
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
	defer cont.Sync.Unlock()

	if !cont.Changed || cont.TuiFileReaderWorker == nil {
		return
	}

	// NB this is just here so we can see updates in this test
	cont.Count++

	// reset the elements by clearing out every one
	//for _, t := range displayResults {
	//	t.Title.SetText("")
	//	t.Body.SetText("")
	//}


	cont.Sync.Lock()
	resultsCopy := make([]*FileJob, len(cont.Results))
	copy(resultsCopy, cont.Results)
	cont.Sync.Unlock()

	// rank all results
	// then go and get the relevant portion for display
	rankResults(int(cont.TuiFileReaderWorker.GetFileCount()), resultsCopy)
	documentTermFrequency := calculateDocumentTermFrequency(resultsCopy)

	// go and get the codeResults the user wants to see using selected as the offset to display from
	var codeResults []codeResult
	for i, res := range resultsCopy {
		if i >= cont.Offset {
			snippets := extractRelevantV3(res, documentTermFrequency, int(SnippetLength), "…")[0]

			// now that we have the relevant portion we need to get just the bits related to highlight it correctly
			// which this method does. It takes in the snippet, we extract and all of the locations and then returns just
			l := getLocated(res, snippets)
			coloredContent := str.HighlightString(snippets.Content, l, "[red]", "[white]")

			codeResults = append(codeResults, codeResult{
				Title:   res.Filename,
				Content: coloredContent,
				Score:   res.Score,
			})
		}
	}

	if len(codeResults) > 20 {
		codeResults = codeResults[:20]
	}

	// render out what the user wants to see based on the results that have been chosen
	tviewApplication.QueueUpdateDraw(func() {
		for i, t := range resultsCopy {
			displayResults[i].Title.SetText(fmt.Sprintf("%d [fuchsia]%s (%f)[-:-:-]", cont.Count, t.Location, t.Score))
			displayResults[i].Body.SetText(string(t.Content))

			//we need to update the item so that it displays everything we have put in
			resultsFlex.ResizeItem(displayResults[i].Body, len(strings.Split(string(t.Content), "\n")), 0)
		}

		statusView.SetText("something")
	})

	// we can only set that nothing
	cont.Changed = false
}

func (cont *tuiApplicationController) doSearch() {
	//cont.Sync.Lock()
	//// deal with the user clearing out the search
	//if cont.Query == "" {
	//	cont.Results = []*FileJob{}
	//	cont.Changed = true
	//	cont.Sync.Unlock()
	//	return
	//}
	//cont.Sync.Unlock()

	// keep the query we are working with
	//query := cont.Query
	//cont.Query = ""

	//if cont.TuiFileWalker != nil && cont.TuiFileWalker.Walking() {
	//	cont.TuiFileWalker.Terminate()
	//}

	fileQueue := make(chan *file.File)                                // NB unbuffered because we want the UI to respond and this is what causes affects
	toProcessQueue := make(chan *FileJob, runtime.NumCPU()) // Files to be read into memory for processing
	summaryQueue := make(chan *FileJob, runtime.NumCPU())   // Files that match and need to be displayed

	cont.TuiFileWalker = file.NewFileWalker(".", fileQueue)
	cont.TuiFileReaderWorker = NewFileReaderWorker(fileQueue, toProcessQueue)
	cont.TuiSearcherWorker = NewSearcherWorker(toProcessQueue, summaryQueue)
	cont.TuiSearcherWorker.SearchString = strings.Split(cont.Query, " ")

	go cont.TuiFileWalker.Start()
	go cont.TuiFileReaderWorker.Start()
	go cont.TuiSearcherWorker.Start()

	// Updated with results as we get them NB this is
	// painted as we go
	var results []*FileJob
	var resultsMutex sync.Mutex
	update := true

	go func() {
		for update {
			// Every 50 ms redraw the current set of results
			resultsMutex.Lock()
			cont.Sync.Lock()
			cont.Results = results
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

	update = false

	cont.Sync.Lock()
	cont.Results = results
	cont.Sync.Unlock()
	cont.SetChanged(true)
}

func (cont *tuiApplicationController) updateView() {
	// render loop running background is the only thing responsible for updating the results based on the state
	// in the applicationController
	go func() {
		for {

			// this needs to go outside the draw because if we need to do the spin update
			//status := cont.calculateStatus()

			cont.drawView()
			time.Sleep(30 * time.Millisecond)
		}
	}()
}

func (cont *tuiApplicationController) calculateStatus() string {
	status := ""
	if cont.TuiFileWalker != nil {
		status = fmt.Sprintf("%d results(s) for '%s' from %d files", len(cont.Results), cont.Query, cont.TuiFileReaderWorker.GetFileCount())
		if cont.GetRunning() {
			status = fmt.Sprintf("%d results(s) for '%s' from %d files %s", len(cont.Results), cont.Query, cont.TuiFileReaderWorker.GetFileCount(), string(cont.SpinString[cont.SpinLocation]))

			cont.SpinRun++
			if cont.SpinRun == 4 {
				cont.SpinLocation++
				if cont.SpinLocation >= len(cont.SpinString) {
					cont.SpinLocation = 0
				}
				cont.SpinRun = 0
				cont.SetChanged(true)
			}
		}
	}
	return status
}

func (cont *tuiApplicationController) processSearch() {
	// we only ever want to have one search trigger at a time which is what this does
	// searches come in... we trigger them to run
	go func() {
		for {
			cont.doSearch()
			time.Sleep(5 * time.Millisecond)
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

func NewTuiApplication() {
	tviewApplication = tview.NewApplication()
	applicationController := tuiApplicationController{
		Sync:             sync.Mutex{},
		SpinString:       `\|/-`,
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
					fmt.Println(applicationController.Results[applicationController.GetOffset()].Location)
				}
				os.Exit(0)
			case tcell.KeyTab:
				//tviewApplication.SetFocus(textView) need to change focus to the others but not the text itself
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
			applicationController.SetQuery(text)
		})

	statusView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		ScrollToBeginning()

	// setup the flex containers to have everything rendered neatly
	queryFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 8, false)

	resultsFlex = tview.NewFlex().SetDirection(tview.FlexRow)

	overallFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(statusView, 1, 0, false).
		AddItem(resultsFlex, 0, 1, false)

	// Add all of the display codeResults into the container ready to be populated
	for _, t := range displayResults {
		resultsFlex.AddItem(nil, 1, 0, false)
		resultsFlex.AddItem(t.Title, 1, 0, false)
		resultsFlex.AddItem(nil, 1, 0, false)
		resultsFlex.AddItem(t.Body, t.BodyHeight, 1, false)
	}

	// trigger the first render without user action
	applicationController.SetChanged(true)

	// trigger the jobs to start running things
	//applicationController.updateView()
	applicationController.processSearch()

	if err := tviewApplication.SetRoot(overallFlex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
