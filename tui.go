// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/boyter/cs/pkg/common"
	"github.com/boyter/cs/pkg/ranker"
	"github.com/boyter/cs/pkg/snippet"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// searchResult represents a single search result
type searchResult struct {
	Filename    string
	Location    string
	Score       float64
	Snippet     string               // plain text snippet (snippet mode)
	SnippetLocs [][]int              // match positions within Snippet [start, end]
	LineRange   string               // line range info
	LineResults []snippet.LineResult // per-line results with positions (lines mode)
	Language    string
	TotalLines  int64
	Code        int64
	Comment     int64
	Blank       int64
	Complexity  int64
}

// debounceTickMsg is sent after the debounce delay to trigger a search
type debounceTickMsg struct {
	seq   int
	query string
}

// searchResultsMsg delivers incremental search results from the search goroutine
type searchResultsMsg struct {
	seq       int
	results   []searchResult
	fileJobs  []*common.FileJob
	done      bool // true = search complete
	total     int  // total files scanned so far
	textTotal int  // non-binary, successfully read files (for BM25 ranking)
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")) // fuchsia/magenta

	selectedTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("5")).
				Bold(true)

	snippetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selectedSnippetStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	matchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")). // red
			Bold(true)

	selectedMatchStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("1")).
				Bold(true)

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
	cfg           *Config
	searchInput   textinput.Model
	snippetInput  textinput.Model
	focusIndex    int // 0=search, 1=snippet
	results       []searchResult
	fileJobs      []*common.FileJob
	selectedIndex int
	scrollOffset  int
	windowHeight  int
	windowWidth   int
	chosen        string // set on Enter, printed after exit

	searchSeq     int                   // monotonic counter, incremented on every text change
	searching     bool                  // true while search is in flight
	searchCancel  context.CancelFunc    // cancels in-flight search; nil if none
	searchResults chan searchResultsMsg // channel from search goroutine
	lastQuery     string                // query that produced current results
	fileCount     int                   // total files scanned (for status line)
	textFileCount int                   // non-binary, successfully read files (for BM25 ranking)
	snippetMode   string                // "snippet" or "lines"
	searchCache   *SearchCache          // caches file locations across progressive queries
}

func initialModel(cfg *Config) model {
	si := textinput.New()
	si.Placeholder = "search query..."
	si.Prompt = "> "
	si.Focus()
	si.CharLimit = 256

	sn := textinput.New()
	sn.Placeholder = ""
	sn.Prompt = ""
	sn.SetValue(fmt.Sprintf("%d", cfg.SnippetLength))
	sn.CharLimit = 5
	sn.Width = 5

	return model{
		cfg:          cfg,
		searchInput:  si,
		snippetInput: sn,
		focusIndex:   0,
		snippetMode:  cfg.SnippetMode,
		searchCache:  NewSearchCache(),
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

	case debounceTickMsg:
		if msg.seq != m.searchSeq {
			return m, nil // stale tick, user kept typing
		}
		if m.searchCancel != nil {
			m.searchCancel() // cancel previous search
		}

		query := strings.TrimSpace(msg.query)
		if query == "" {
			m.results = nil
			m.fileJobs = nil
			m.searching = false
			m.fileCount = 0
			return m, nil
		}

		// Parse snippet length from input
		snippetLen := 300
		if v := m.snippetInput.Value(); v != "" {
			fmt.Sscanf(v, "%d", &snippetLen)
		}

		ctx, cancel := context.WithCancel(context.Background())
		m.searchCancel = cancel
		m.searching = true
		m.results = nil
		m.fileJobs = nil
		m.selectedIndex = 0
		m.scrollOffset = 0
		m.fileCount = 0
		ch := make(chan searchResultsMsg, 1)
		m.searchResults = ch
		go realSearch(ctx, m.cfg, m.searchSeq, query, snippetLen, m.snippetMode, m.searchCache, ch)
		return m, listenForResults(ch)

	case searchResultsMsg:
		if msg.seq != m.searchSeq {
			return m, nil // stale results from old search
		}

		m.results = append(m.results, msg.results...)
		m.fileJobs = append(m.fileJobs, msg.fileJobs...)
		m.fileCount = msg.total
		m.textFileCount = msg.textTotal

		if msg.done {
			m.searching = false
			m.searchCancel = nil
			m.lastQuery = m.searchInput.Value()

			// Rank all results with BM25 and re-extract snippets with global frequencies
			if len(m.fileJobs) > 0 {
				testIntent := ranker.HasTestIntent(strings.Fields(m.searchInput.Value()))
				ranked := ranker.RankResults(m.cfg.Ranker, m.textFileCount, m.fileJobs, m.cfg.StructuralRankerConfig(), m.cfg.ResolveGravityStrength(), m.cfg.ResolveNoiseSensitivity(), m.cfg.TestPenalty, testIntent)
				docFreq := ranker.CalculateDocumentTermFrequency(ranked)

				// Parse snippet length from input
				snippetLen := 300
				if v := m.snippetInput.Value(); v != "" {
					fmt.Sscanf(v, "%d", &snippetLen)
				}

				var newResults []searchResult
				for _, fj := range ranked {
					fileMode := resolveSnippetMode(m.snippetMode, fj.Filename)
					if fileMode == "lines" {
						lineResults := snippet.FindMatchingLines(fj, 2)
						lineRange := ""
						if len(lineResults) > 0 {
							lineRange = fmt.Sprintf("%d-%d",
								lineResults[0].LineNumber,
								lineResults[len(lineResults)-1].LineNumber)
						}
						fj.Content = nil
						newResults = append(newResults, searchResult{
							Filename:    fj.Location,
							Location:    fj.Location,
							Score:       fj.Score,
							LineResults: lineResults,
							LineRange:   lineRange,
							Language:    fj.Language,
							TotalLines:  fj.Lines,
							Code:        fj.Code,
							Comment:     fj.Comment,
							Blank:       fj.Blank,
							Complexity:  fj.Complexity,
						})
					} else {
						snippets := snippet.ExtractRelevant(fj, docFreq, snippetLen)
						snippetText := ""
						lineRange := ""
						var sLocs [][]int
						if len(snippets) > 0 {
							snippetText = snippets[0].Content
							lineRange = fmt.Sprintf("%d-%d", snippets[0].LineStart, snippets[0].LineEnd)
							sLocs = snippetMatchLocs(fj.MatchLocations, snippets[0].StartPos, snippets[0].EndPos)
						}
						fj.Content = nil
						newResults = append(newResults, searchResult{
							Filename:    fj.Location,
							Location:    fj.Location,
							Score:       fj.Score,
							Snippet:     snippetText,
							SnippetLocs: sLocs,
							LineRange:   lineRange,
							Language:    fj.Language,
							TotalLines:  fj.Lines,
							Code:        fj.Code,
							Comment:     fj.Comment,
							Blank:       fj.Blank,
							Complexity:  fj.Complexity,
						})
					}
				}
				m.results = newResults
				m.fileJobs = nil
				m.selectedIndex = 0
				m.scrollOffset = 0
			}

			return m, nil
		}
		// Keep listening for more results
		return m, listenForResults(m.searchResults)

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			m.scrollOffset -= 3
			m.clampScroll()
			m.syncSelectedToScroll()
			return m, nil
		case tea.MouseWheelDown:
			m.scrollOffset += 3
			m.clampScroll()
			m.syncSelectedToScroll()
			return m, nil
		case tea.MouseLeft:
			idx := m.resultIndexAtY(msg.Y)
			if idx >= 0 && idx < len(m.results) {
				m.selectedIndex = idx
			}
			return m, nil
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.searchCancel != nil {
				m.searchCancel()
			}
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.results) > 0 && m.selectedIndex < len(m.results) {
				if m.searchCancel != nil {
					m.searchCancel()
				}
				m.chosen = m.results[m.selectedIndex].Location
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
				m.adjustSnippetLength(100)
				if q := strings.TrimSpace(m.searchInput.Value()); q != "" {
					m.searchSeq++
					return m, makeDebounceCmd(m.searchSeq, m.searchInput.Value())
				}
				return m, nil
			}
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.ensureVisible()
			}
			return m, nil

		case tea.KeyDown:
			if m.focusIndex == 1 {
				m.adjustSnippetLength(-100)
				if q := strings.TrimSpace(m.searchInput.Value()); q != "" {
					m.searchSeq++
					return m, makeDebounceCmd(m.searchSeq, m.searchInput.Value())
				}
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
				if q := strings.TrimSpace(m.searchInput.Value()); q != "" {
					m.searchSeq++
					return m, makeDebounceCmd(m.searchSeq, m.searchInput.Value())
				}
				return m, nil
			}

		case tea.KeyPgDown:
			if m.focusIndex == 1 {
				m.adjustSnippetLength(-200)
				if q := strings.TrimSpace(m.searchInput.Value()); q != "" {
					m.searchSeq++
					return m, makeDebounceCmd(m.searchSeq, m.searchInput.Value())
				}
				return m, nil
			}

		case tea.KeyF1:
			m.cycleRanker()
			return m, m.retriggerSearch()
		case tea.KeyF2:
			m.cycleCodeFilter()
			return m, m.retriggerSearch()
		case tea.KeyF3:
			m.cycleGravity()
			return m, m.retriggerSearch()
		case tea.KeyF4:
			m.cycleNoise()
			return m, m.retriggerSearch()
		}
	}

	// Update the focused input
	if m.focusIndex == 0 {
		prevValue := m.searchInput.Value()
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)

		// On text change, increment seq and start debounce timer
		if m.searchInput.Value() != prevValue {
			m.searchSeq++
			cmds = append(cmds, makeDebounceCmd(m.searchSeq, m.searchInput.Value()))
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

// makeDebounceCmd returns a tea.Cmd that fires a debounceTickMsg after 200ms
func makeDebounceCmd(seq int, query string) tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return debounceTickMsg{seq: seq, query: query}
	})
}

// listenForResults returns a tea.Cmd that blocks until the next message arrives on ch
func listenForResults(ch <-chan searchResultsMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return searchResultsMsg{done: true}
		}
		return msg
	}
}

// realSearch wraps DoSearch for TUI use, streaming results in batches via the channel.
func realSearch(ctx context.Context, cfg *Config, seq int, query string, snippetLen int, snippetMode string, cache *SearchCache, ch chan<- searchResultsMsg) {
	defer close(ch)

	searchCh, stats := DoSearch(ctx, cfg, query, cache)

	var batch []searchResult
	var batchJobs []*common.FileJob

	for fj := range searchCh {
		// Build a preliminary searchResult for immediate display
		fileMode := resolveSnippetMode(snippetMode, fj.Filename)
		var sr searchResult
		if fileMode == "lines" {
			lineResults := snippet.FindMatchingLines(fj, 2)
			lineRange := ""
			if len(lineResults) > 0 {
				lineRange = fmt.Sprintf("%d-%d",
					lineResults[0].LineNumber,
					lineResults[len(lineResults)-1].LineNumber)
			}
			sr = searchResult{
				Filename:    fj.Location,
				Location:    fj.Location,
				LineResults: lineResults,
				LineRange:   lineRange,
				Language:    fj.Language,
				TotalLines:  fj.Lines,
				Code:        fj.Code,
				Comment:     fj.Comment,
				Blank:       fj.Blank,
				Complexity:  fj.Complexity,
			}
		} else {
			docFreq := make(map[string]int, len(fj.MatchLocations))
			for k, v := range fj.MatchLocations {
				docFreq[k] = len(v)
			}
			snippets := snippet.ExtractRelevant(fj, docFreq, snippetLen)
			snippetText := ""
			lineRange := ""
			var sLocs [][]int
			if len(snippets) > 0 {
				snippetText = snippets[0].Content
				lineRange = fmt.Sprintf("%d-%d", snippets[0].LineStart, snippets[0].LineEnd)
				sLocs = snippetMatchLocs(fj.MatchLocations, snippets[0].StartPos, snippets[0].EndPos)
			}
			sr = searchResult{
				Filename:    fj.Location,
				Location:    fj.Location,
				Snippet:     snippetText,
				SnippetLocs: sLocs,
				LineRange:   lineRange,
				Language:    fj.Language,
				TotalLines:  fj.Lines,
				Code:        fj.Code,
				Comment:     fj.Comment,
				Blank:       fj.Blank,
				Complexity:  fj.Complexity,
			}
		}

		batch = append(batch, sr)
		batchJobs = append(batchJobs, fj)

		if len(batch) >= 5 {
			select {
			case ch <- searchResultsMsg{
				seq: seq, results: batch, fileJobs: batchJobs,
				total: int(stats.FileCount.Load()), textTotal: int(stats.TextFileCount.Load()),
			}:
				batch = nil
				batchJobs = nil
			case <-ctx.Done():
				return
			}
		}
	}

	// Send remaining results + done signal
	select {
	case ch <- searchResultsMsg{
		seq: seq, results: batch, fileJobs: batchJobs,
		done: true, total: int(stats.FileCount.Load()), textTotal: int(stats.TextFileCount.Load()),
	}:
	case <-ctx.Done():
	}
}

// snippetMatchLocs filters match locations to those within [startPos, endPos]
// and adjusts them to be relative to startPos (matching console.go behavior).
func snippetMatchLocs(matchLocations map[string][][]int, startPos, endPos int) [][]int {
	var locs [][]int
	for _, value := range matchLocations {
		for _, s := range value {
			if s[0] >= startPos && s[1] <= endPos {
				locs = append(locs, []int{
					s[0] - startPos,
					s[1] - startPos,
				})
			}
		}
	}
	return locs
}

// resultHeight returns the number of terminal lines a result takes up
func resultHeight(r searchResult) int {
	if len(r.LineResults) > 0 {
		gaps := 0
		for i := 1; i < len(r.LineResults); i++ {
			if r.LineResults[i].LineNumber > r.LineResults[i-1].LineNumber+1 {
				gaps++
			}
		}
		return 1 + len(r.LineResults) + gaps + 1
	}
	// 1 for title + lines in snippet + 1 blank line separator
	lines := strings.Count(r.Snippet, "\n") + 1
	return 1 + lines + 1
}

func (m *model) totalContentHeight() int {
	total := 0
	for _, r := range m.results {
		total += resultHeight(r)
	}
	return total
}

func (m *model) clampScroll() {
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	availHeight := m.windowHeight - 5
	if availHeight < 1 {
		availHeight = 1
	}
	maxScroll := m.totalContentHeight() - availHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
}

func (m *model) resultIndexAtY(y int) int {
	const headerLines = 2
	if y < headerLines {
		return -1
	}
	contentLine := m.scrollOffset + (y - headerLines)
	accum := 0
	for i, r := range m.results {
		rh := resultHeight(r)
		if contentLine < accum+rh {
			return i
		}
		accum += rh
	}
	return -1
}

func (m *model) syncSelectedToScroll() {
	if len(m.results) == 0 {
		return
	}
	availHeight := m.windowHeight - 5
	if availHeight < 1 {
		availHeight = 1
	}
	accum := 0
	firstVisible, lastVisible := -1, -1
	for i, r := range m.results {
		rh := resultHeight(r)
		if accum+rh > m.scrollOffset && accum < m.scrollOffset+availHeight {
			if firstVisible == -1 {
				firstVisible = i
			}
			lastVisible = i
		}
		accum += rh
		if accum >= m.scrollOffset+availHeight && lastVisible >= 0 {
			break
		}
	}
	if firstVisible == -1 {
		return
	}
	if m.selectedIndex < firstVisible {
		m.selectedIndex = firstVisible
	}
	if m.selectedIndex > lastVisible {
		m.selectedIndex = lastVisible
	}
}

func (m *model) ensureVisible() {
	// Calculate available height for results area
	availHeight := m.windowHeight - 5 // input + status + separator + bottom bar + overhead

	// Make sure selected item is visible by adjusting scroll offset
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

// cycleRanker cycles the ranker through: simple → tfidf → bm25 → structural → simple…
func (m *model) cycleRanker() {
	order := []string{"simple", "tfidf", "bm25", "structural"}
	for i, v := range order {
		if v == m.cfg.Ranker {
			m.cfg.Ranker = order[(i+1)%len(order)]
			return
		}
	}
	m.cfg.Ranker = "simple"
}

// cycleCodeFilter cycles: default → only-code → only-comments → only-strings → default…
// Auto-switches ranker to "structural" when a filter is active.
func (m *model) cycleCodeFilter() {
	switch {
	case !m.cfg.OnlyCode && !m.cfg.OnlyComments && !m.cfg.OnlyStrings:
		m.cfg.OnlyCode = true
		m.cfg.OnlyComments = false
		m.cfg.OnlyStrings = false
	case m.cfg.OnlyCode:
		m.cfg.OnlyCode = false
		m.cfg.OnlyComments = true
		m.cfg.OnlyStrings = false
	case m.cfg.OnlyComments:
		m.cfg.OnlyCode = false
		m.cfg.OnlyComments = false
		m.cfg.OnlyStrings = true
	default:
		m.cfg.OnlyCode = false
		m.cfg.OnlyComments = false
		m.cfg.OnlyStrings = false
	}
	if m.cfg.HasContentFilter() {
		m.cfg.Ranker = "structural"
	}
}

// cycleGravity cycles: off → low → default → logic → brain → off…
func (m *model) cycleGravity() {
	order := []string{"off", "low", "default", "logic", "brain"}
	for i, v := range order {
		if v == m.cfg.GravityIntent {
			m.cfg.GravityIntent = order[(i+1)%len(order)]
			return
		}
	}
	m.cfg.GravityIntent = "off"
}

// cycleNoise cycles: silence → quiet → default → loud → raw → silence…
func (m *model) cycleNoise() {
	order := []string{"silence", "quiet", "default", "loud", "raw"}
	for i, v := range order {
		if v == m.cfg.NoiseIntent {
			m.cfg.NoiseIntent = order[(i+1)%len(order)]
			return
		}
	}
	m.cfg.NoiseIntent = "silence"
}

// retriggerSearch bumps the search sequence and returns a debounce command to re-execute the current query.
func (m *model) retriggerSearch() tea.Cmd {
	m.searchSeq++
	return makeDebounceCmd(m.searchSeq, m.searchInput.Value())
}

// codeFilterLabel returns a display label for the current code filter state.
func (m *model) codeFilterLabel() string {
	switch {
	case m.cfg.OnlyCode:
		return "only-code"
	case m.cfg.OnlyComments:
		return "only-comments"
	case m.cfg.OnlyStrings:
		return "only-strings"
	default:
		return "default"
	}
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
	} else if m.searching {
		status = fmt.Sprintf("%d results for '%s' from %d files (searching...)",
			len(m.results), query, m.fileCount)
	} else {
		plural := "s"
		if len(m.results) == 1 {
			plural = ""
		}
		status = fmt.Sprintf("%d result%s for '%s' from %d files",
			len(m.results), plural, m.lastQuery, m.fileCount)
	}
	b.WriteString(statusStyle.Render(status))
	b.WriteString("\n")

	// === Results area ===
	availHeight := m.windowHeight - 5 // input + status + separator + bottom bar + overhead
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
		resultStr := m.renderResult(r, isSelected)
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

	// === Bottom bar (nano-style keybinding hints) ===
	bottomBar := snippetLabelStyle.Render(fmt.Sprintf(
		"F1 Ranker:%s  F2 Filter:%s  F3 Gravity:%s  F4 Noise:%s",
		m.cfg.Ranker, m.codeFilterLabel(), m.cfg.GravityIntent, m.cfg.NoiseIntent))
	b.WriteString("\n")
	b.WriteString(bottomBar)

	return b.String()
}

func (m model) renderResult(r searchResult, isSelected bool) string {
	var b strings.Builder

	// Title line: indicator + filename (language) (score) [linerange] code stats
	codeStats := ""
	if r.TotalLines > 0 {
		codeStats = fmt.Sprintf(" Lines:%d (Code:%d Comment:%d Blank:%d Complexity:%d)", r.TotalLines, r.Code, r.Comment, r.Blank, r.Complexity)
	}
	var titleText string
	if r.Language != "" {
		titleText = fmt.Sprintf("%s (%s) (%0.4f) [%s]%s", r.Filename, r.Language, r.Score, r.LineRange, codeStats)
	} else {
		titleText = fmt.Sprintf("%s (%0.4f) [%s]%s", r.Filename, r.Score, r.LineRange, codeStats)
	}

	if isSelected {
		b.WriteString(selectedIndicator.String())
		b.WriteString(selectedTitleStyle.Render(titleText))
	} else {
		b.WriteString(" ")
		b.WriteString(titleStyle.Render(titleText))
	}
	b.WriteString("\n")

	// Render content: line-based or snippet-based
	if len(r.LineResults) > 0 {
		prevLine := 0
		for _, lr := range r.LineResults {
			if prevLine > 0 && lr.LineNumber > prevLine+1 {
				b.WriteString("\n")
			}
			prevLine = lr.LineNumber
			prefix := "  "
			if isSelected {
				prefix = selectedIndicator.String() + " "
			}
			lineNum := snippetLabelStyle.Render(fmt.Sprintf("%4d ", lr.LineNumber))
			highlighted := m.highlightWithLocs(lr.Content, lr.Locs, isSelected)
			b.WriteString(prefix + lineNum + highlighted)
			b.WriteString("\n")
		}
	} else {
		snippetLines := strings.Split(r.Snippet, "\n")
		offset := 0
		for _, line := range snippetLines {
			prefix := "  "
			if isSelected {
				prefix = selectedIndicator.String() + " "
			}

			// Compute per-line match locations from snippet-wide SnippetLocs
			var lineLocs [][]int
			lineEnd := offset + len(line)
			for _, loc := range r.SnippetLocs {
				if loc[1] <= offset || loc[0] >= lineEnd {
					continue // outside this line
				}
				start := loc[0] - offset
				end := loc[1] - offset
				if start < 0 {
					start = 0
				}
				if end > len(line) {
					end = len(line)
				}
				lineLocs = append(lineLocs, []int{start, end})
			}

			highlighted := m.highlightWithLocs(line, lineLocs, isSelected)
			b.WriteString(prefix + highlighted)
			b.WriteString("\n")
			offset = lineEnd + 1 // +1 for the \n
		}
	}

	return b.String()
}

func (m model) highlightWithLocs(line string, locs [][]int, isSelected bool) string {
	if !m.cfg.NoSyntax {
		return RenderLipglossLine(line, locs, isSelected)
	}
	return highlightMatchOnly(line, locs, isSelected)
}

// highlightMatchOnly is the original match-only highlighting (no syntax colors).
func highlightMatchOnly(line string, locs [][]int, isSelected bool) string {
	normal := snippetStyle
	highlight := matchStyle
	if isSelected {
		normal = selectedSnippetStyle
		highlight = selectedMatchStyle
	}

	if len(locs) == 0 {
		return normal.Render(line)
	}

	marked := make([]bool, len(line))
	for _, loc := range locs {
		for i := loc[0]; i < loc[1] && i < len(marked); i++ {
			marked[i] = true
		}
	}

	var result strings.Builder
	inMatch := false
	segStart := 0

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
