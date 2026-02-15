// SPDX-License-Identifier: MIT OR Unlicense

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// searchResult represents a single search result with mock data
type searchResult struct {
	Filename string
	Score    float64
	Snippet  string // plain text snippet
	Lines    string // line range info
}

var mockResults = []searchResult{
	{
		Filename: "searcher.go",
		Score:    0.9847,
		Snippet:  "func SearcherWorker(input chan *FileJob, output chan *FileJob) {\n\tfor res := range input {\n\t\tif res.Binary {\n\t\t\tcontinue\n\t\t}\n\t\tfor _, term := range searchTerms {\n\t\t\tif matchContent(res.Content, term) {\n\t\t\t\toutput <- res\n\t\t\t}\n\t\t}\n\t}",
		Lines:    "45-56",
	},
	{
		Filename: "search.go",
		Score:    0.8523,
		Snippet:  "func PreParseQuery(args []string) ([]SearchTerm, bool) {\n\tvar searchTerms []SearchTerm\n\tfuzzy := false\n\tfor _, arg := range args {\n\t\tif strings.HasPrefix(arg, \"file:\") {\n\t\t\tcontinue\n\t\t}\n\t\tsearchTerms = append(searchTerms, parseSearchTerm(arg))\n\t}",
		Lines:    "12-20",
	},
	{
		Filename: "ranker.go",
		Score:    0.7234,
		Snippet:  "func rankResults(documentCount int, results []*FileJob) {\n\tswitch strings.ToLower(Ranker) {\n\tcase \"bm25\":\n\t\tfor _, r := range results {\n\t\t\tr.Score = calculateBM25(r, documentCount)\n\t\t}\n\tcase \"tfidf\":\n\t\tfor _, r := range results {\n\t\t\tr.Score = calculateTfIdf(r, documentCount)\n\t\t}",
		Lines:    "8-17",
	},
	{
		Filename: "file.go",
		Score:    0.6891,
		Snippet:  "func FileReaderWorker(input chan string, output chan *FileJob) {\n\tfor path := range input {\n\t\tdata, err := os.ReadFile(path)\n\t\tif err != nil {\n\t\t\tcontinue\n\t\t}\n\t\toutput <- &FileJob{\n\t\t\tLocation: path,\n\t\t\tContent:  data,\n\t\t}",
		Lines:    "30-40",
	},
	{
		Filename: "snippet.go",
		Score:    0.5912,
		Snippet:  "func extractRelevantV3(res *FileJob, freq map[string]int, snippetLen int) []Snippet {\n\twindowSize := snippetLen\n\tbestScore := 0\n\tbestPos := 0\n\tcontent := string(res.Content)",
		Lines:    "88-93",
	},
	{
		Filename: "main.go",
		Score:    0.4756,
		Snippet:  "func main() {\n\trootCmd := &cobra.Command{\n\t\tUse:   \"cs [query]\",\n\t\tShort: \"Code search tool\",\n\t\tRun: func(cmd *cobra.Command, args []string) {\n\t\t\tif SearchHTTP {\n\t\t\t\tStartHttpServer()\n\t\t\t} else if len(args) > 0 {\n\t\t\t\tNewConsoleSearch(args)\n\t\t\t}",
		Lines:    "15-24",
	},
	{
		Filename: "console.go",
		Score:    0.3845,
		Snippet:  "func NewConsoleSearch(args []string) {\n\tquery := strings.Join(args, \" \")\n\tfiles := FindFiles(query)\n\ttoProcessQueue := make(chan *FileJob)\n\tsummaryQueue := make(chan *FileJob)",
		Lines:    "22-27",
	},
	{
		Filename: "tui.go",
		Score:    0.3102,
		Snippet:  "func NewTuiSearch() {\n\ttviewApplication = tview.NewApplication()\n\tapplicationController := tuiApplicationController{\n\t\tMutex:      sync.Mutex{},\n\t\tSpinString: `\\|/-`,\n\t}",
		Lines:    "249-254",
	},
	{
		Filename: "globals.go",
		Score:    0.2481,
		Snippet:  "var SnippetLength int64 = 300\nvar SnippetCount int64 = 1\nvar Ranker = \"bm25\"\nvar CaseSensitive = false\nvar SearchHTTP = false",
		Lines:    "8-12",
	},
	{
		Filename: "vendor/github.com/boyter/gocodewalker/walker.go",
		Score:    0.1234,
		Snippet:  "func walkFiles(root string, output chan string) {\n\tfilepath.Walk(root, func(path string, info os.FileInfo, err error) error {\n\t\tif err != nil {\n\t\t\treturn nil\n\t\t}\n\t\tif !info.IsDir() {\n\t\t\toutput <- path\n\t\t}",
		Lines:    "45-53",
	},
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")) // fuchsia/magenta

	selectedTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("5")).
				Bold(true).
				Background(lipgloss.Color("236"))

	snippetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selectedSnippetStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236"))

	matchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")). // red
			Bold(true)

	selectedMatchStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("1")).
				Bold(true).
				Background(lipgloss.Color("236"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	snippetLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))

	selectedIndicator = lipgloss.NewStyle().
				Foreground(lipgloss.Color("5")).
				SetString("▎")
)

type model struct {
	searchInput   textinput.Model
	snippetInput  textinput.Model
	focusIndex    int // 0=search, 1=snippet
	results       []searchResult
	allResults    []searchResult // unfiltered
	selectedIndex int
	scrollOffset  int
	windowHeight  int
	windowWidth   int
	chosen        string // set on Enter, printed after exit
}

func initialModel() model {
	si := textinput.New()
	si.Placeholder = "search query..."
	si.Prompt = "> "
	si.Focus()
	si.CharLimit = 256

	sn := textinput.New()
	sn.Placeholder = ""
	sn.Prompt = ""
	sn.SetValue("300")
	sn.CharLimit = 5
	sn.Width = 5

	return model{
		searchInput:  si,
		snippetInput: sn,
		focusIndex:   0,
		results:      mockResults,
		allResults:   mockResults,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, textinput.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width
		m.clampScroll()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.results) > 0 && m.selectedIndex < len(m.results) {
				m.chosen = m.results[m.selectedIndex].Filename
				return m, tea.Quit
			}
			return m, nil

		case tea.KeyTab, tea.KeyShiftTab:
			if m.focusIndex == 0 {
				m.focusIndex = 1
				m.searchInput.Blur()
				m.snippetInput.Focus()
			} else {
				m.focusIndex = 0
				m.snippetInput.Blur()
				m.searchInput.Focus()
			}
			return m, nil

		case tea.KeyUp:
			if m.focusIndex == 1 {
				// In snippet field, increase value
				m.adjustSnippetLength(100)
				return m, nil
			}
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.ensureVisible()
			}
			return m, nil

		case tea.KeyDown:
			if m.focusIndex == 1 {
				// In snippet field, decrease value
				m.adjustSnippetLength(-100)
				return m, nil
			}
			if m.selectedIndex < len(m.results)-1 {
				m.selectedIndex++
				m.ensureVisible()
			}
			return m, nil

		case tea.KeyPgUp:
			if m.focusIndex == 1 {
				m.adjustSnippetLength(200)
				return m, nil
			}

		case tea.KeyPgDown:
			if m.focusIndex == 1 {
				m.adjustSnippetLength(-200)
				return m, nil
			}
		}
	}

	// Update the focused input
	if m.focusIndex == 0 {
		prevValue := m.searchInput.Value()
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)

		// Filter results on search change
		if m.searchInput.Value() != prevValue {
			m.filterResults()
		}
	} else {
		var cmd tea.Cmd
		m.snippetInput, cmd = m.snippetInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) adjustSnippetLength(delta int) {
	val := 300
	if v := m.snippetInput.Value(); v != "" {
		fmt.Sscanf(v, "%d", &val)
	}
	val += delta
	if val < 100 {
		val = 100
	}
	if val > 8000 {
		val = 8000
	}
	m.snippetInput.SetValue(fmt.Sprintf("%d", val))
}

func (m *model) filterResults() {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	if query == "" {
		m.results = m.allResults
	} else {
		terms := strings.Fields(query)
		var filtered []searchResult
		for _, r := range m.allResults {
			match := true
			for _, term := range terms {
				if !strings.Contains(strings.ToLower(r.Filename), term) &&
					!strings.Contains(strings.ToLower(r.Snippet), term) {
					match = false
					break
				}
			}
			if match {
				filtered = append(filtered, r)
			}
		}
		m.results = filtered
	}
	m.selectedIndex = 0
	m.scrollOffset = 0
}

// resultHeight returns the number of terminal lines a result takes up
func resultHeight(r searchResult) int {
	// 1 for title + lines in snippet + 1 blank line separator
	lines := strings.Count(r.Snippet, "\n") + 1
	return 1 + lines + 1
}

func (m *model) clampScroll() {
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *model) ensureVisible() {
	// Calculate available height for results area
	availHeight := m.windowHeight - 3 // input line + status line + separator

	// Make sure selected item is visible by adjusting scroll offset
	// Calculate the cumulative height up to the selected item
	heightBefore := 0
	for i := 0; i < m.selectedIndex; i++ {
		if i < len(m.results) {
			heightBefore += resultHeight(m.results[i])
		}
	}

	selectedH := 0
	if m.selectedIndex < len(m.results) {
		selectedH = resultHeight(m.results[m.selectedIndex])
	}

	// Scroll up if selected is above viewport
	if heightBefore < m.scrollOffset {
		m.scrollOffset = heightBefore
	}

	// Scroll down if selected is below viewport
	if heightBefore+selectedH > m.scrollOffset+availHeight {
		m.scrollOffset = heightBefore + selectedH - availHeight
	}

	m.clampScroll()
}

func (m model) View() string {
	if m.windowWidth == 0 {
		return "loading..."
	}

	var b strings.Builder

	// === Top line: search input + snippet length ===
	snippetLabel := snippetLabelStyle.Render("[" + m.snippetInput.View() + "]")
	snippetWidth := lipgloss.Width(snippetLabel)
	searchWidth := m.windowWidth - snippetWidth - 1
	if searchWidth < 10 {
		searchWidth = 10
	}
	m.searchInput.Width = searchWidth - 3 // account for prompt "> "

	topLine := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(searchWidth).Render(m.searchInput.View()),
		lipgloss.NewStyle().Width(snippetWidth).Align(lipgloss.Right).Render(snippetLabel),
	)
	b.WriteString(topLine)
	b.WriteString("\n")

	// === Status line ===
	query := m.searchInput.Value()
	var status string
	if query == "" {
		status = "type a query to search"
	} else {
		plural := "s"
		if len(m.results) == 1 {
			plural = ""
		}
		status = fmt.Sprintf("%d result%s for '%s' from 450 files", len(m.results), plural, query)
	}
	b.WriteString(statusStyle.Render(status))
	b.WriteString("\n")

	// === Results area ===
	availHeight := m.windowHeight - 3 // input + status + separator line
	if availHeight < 1 {
		availHeight = 1
	}

	linesRendered := 0
	linesSkipped := 0

	for i, r := range m.results {
		rh := resultHeight(r)

		// Skip results that are above the scroll offset
		if linesSkipped+rh <= m.scrollOffset {
			linesSkipped += rh
			continue
		}

		// Stop if we've filled the viewport
		if linesRendered >= availHeight {
			break
		}

		isSelected := i == m.selectedIndex
		resultStr := m.renderResult(r, isSelected, query)
		resultLines := strings.Split(resultStr, "\n")

		// Handle partial rendering at top (scroll offset cuts into this result)
		startLine := 0
		if linesSkipped < m.scrollOffset {
			startLine = m.scrollOffset - linesSkipped
			linesSkipped = m.scrollOffset
		}

		for j := startLine; j < len(resultLines); j++ {
			if linesRendered >= availHeight {
				break
			}
			b.WriteString(resultLines[j])
			b.WriteString("\n")
			linesRendered++
		}
	}

	// Fill remaining space so alt screen doesn't look jagged
	for linesRendered < availHeight {
		b.WriteString("\n")
		linesRendered++
	}

	return b.String()
}

func (m model) renderResult(r searchResult, isSelected bool, query string) string {
	var b strings.Builder

	// Title line: indicator + filename (score)
	titleText := fmt.Sprintf("%s (%0.4f) [%s]", r.Filename, r.Score, r.Lines)

	if isSelected {
		b.WriteString(selectedIndicator.String())
		b.WriteString(selectedTitleStyle.Render(titleText))
	} else {
		b.WriteString(" ")
		b.WriteString(titleStyle.Render(titleText))
	}
	b.WriteString("\n")

	// Snippet lines with match highlighting
	snippetLines := strings.Split(r.Snippet, "\n")
	for _, line := range snippetLines {
		prefix := "  "
		if isSelected {
			prefix = selectedIndicator.String() + " "
		}

		highlighted := m.highlightMatches(line, query, isSelected)
		b.WriteString(prefix + highlighted)
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) highlightMatches(line string, query string, isSelected bool) string {
	if query == "" {
		if isSelected {
			return selectedSnippetStyle.Render(line)
		}
		return snippetStyle.Render(line)
	}

	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		if isSelected {
			return selectedSnippetStyle.Render(line)
		}
		return snippetStyle.Render(line)
	}

	// Find all match positions
	type span struct {
		start, end int
	}
	var matches []span
	lower := strings.ToLower(line)
	for _, term := range terms {
		idx := 0
		for {
			pos := strings.Index(lower[idx:], term)
			if pos == -1 {
				break
			}
			matches = append(matches, span{idx + pos, idx + pos + len(term)})
			idx += pos + len(term)
		}
	}

	if len(matches) == 0 {
		if isSelected {
			return selectedSnippetStyle.Render(line)
		}
		return snippetStyle.Render(line)
	}

	// Sort and merge overlapping spans
	// Simple approach: mark each character as matched or not
	marked := make([]bool, len(line))
	for _, m := range matches {
		for i := m.start; i < m.end && i < len(marked); i++ {
			marked[i] = true
		}
	}

	// Build highlighted string
	var result strings.Builder
	inMatch := false
	segStart := 0

	normal := snippetStyle
	highlight := matchStyle
	if isSelected {
		normal = selectedSnippetStyle
		highlight = selectedMatchStyle
	}

	for i := 0; i <= len(line); i++ {
		currentMatch := i < len(line) && marked[i]
		if i == len(line) || currentMatch != inMatch {
			seg := line[segStart:i]
			if inMatch {
				result.WriteString(highlight.Render(seg))
			} else {
				result.WriteString(normal.Render(seg))
			}
			segStart = i
			inMatch = currentMatch
		}
	}

	return result.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if fm, ok := m.(model); ok && fm.chosen != "" {
		fmt.Println(fm.chosen)
	}
}
