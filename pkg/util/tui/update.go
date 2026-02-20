package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case modeSearch:
			return m.handleSearchKey(msg)
		case modeSearchResults:
			return m.handleSearchResultsKey(msg)
		case modeQuery:
			return m.handleQueryKey(msg)
		default:
			return m.handleKeyPress(msg)
		}

	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)

	case tickMsg:
		// Pause auto-refresh while user is in search or query mode
		if m.mode != modeNormal {
			return m, tickCmd(m.refreshInterval)
		}
		m.loading = true
		return m, tea.Batch(
			fetchData(m.dataFetcher),
			tickCmd(m.refreshInterval),
		)

	case fetchDataMsg:
		m.loading = false
		m.lastUpdate = time.Now()

		if msg.err != nil {
			m.lastError = msg.err
		} else {
			m.lastError = nil
			m.content = msg.content
			if m.searchTerm != "" {
				m.reapplySearch()
			} else {
				m.viewport.SetContent(m.content)
			}
		}

		if !m.ready {
			m.ready = true
		}

		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Handle viewport updates
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleKeyPress handles keyboard input in normal mode
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle help toggle
	if key.Matches(msg, m.keys.Help) {
		m.showHelp = !m.showHelp
		if m.showHelp {
			m.help.ShowAll = true
		}
		return m, nil
	}

	// If help is showing, hide it on any other key
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	// Handle quit
	if key.Matches(msg, m.keys.Quit) {
		m.quitting = true
		return m, tea.Quit
	}

	// Enter search mode
	if key.Matches(msg, m.keys.Search) {
		m.mode = modeSearch
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		return m, m.searchInput.Cursor.BlinkCmd()
	}

	// Enter query mode (only when QueryUpdater is set)
	if key.Matches(msg, m.keys.Query) && m.queryUpdater != nil {
		m.mode = modeQuery
		m.queryInput.SetValue(m.currentQuery)
		m.queryInput.Focus()
		m.queryInput.CursorEnd()
		return m, m.queryInput.Cursor.BlinkCmd()
	}

	// Handle refresh
	if key.Matches(msg, m.keys.Refresh) {
		m.loading = true
		return m, fetchData(m.dataFetcher)
	}

	// Handle interval adjustments
	if key.Matches(msg, m.keys.IncreaseInt) {
		m.refreshInterval += 5 * time.Second
		if m.refreshInterval > 300*time.Second {
			m.refreshInterval = 300 * time.Second
		}
		return m, nil
	}

	if key.Matches(msg, m.keys.DecreaseInt) {
		m.refreshInterval -= 5 * time.Second
		if m.refreshInterval < 5*time.Second {
			m.refreshInterval = 5 * time.Second
		}
		return m, nil
	}

	// Handle viewport navigation
	var cmd tea.Cmd
	if key.Matches(msg, m.keys.Up) {
		m.viewport.ScrollUp(1)
	} else if key.Matches(msg, m.keys.Down) {
		m.viewport.ScrollDown(1)
	} else if key.Matches(msg, m.keys.PageUp) {
		m.viewport.HalfPageUp()
	} else if key.Matches(msg, m.keys.PageDown) {
		m.viewport.HalfPageDown()
	} else if key.Matches(msg, m.keys.Home) {
		m.viewport.GotoTop()
	} else if key.Matches(msg, m.keys.End) {
		m.viewport.GotoBottom()
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

// handleSearchKey handles keyboard input in search mode (typing the search term).
func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Confirm search — enter search-results navigation mode
		m.searchInput.Blur()
		if len(m.searchMatches) > 0 {
			m.mode = modeSearchResults
			m.updateFocusHighlight()
		} else {
			m.mode = modeNormal
		}
		return m, nil

	case tea.KeyEscape:
		// Cancel search — clear highlights, return to normal
		m.mode = modeNormal
		m.searchInput.Blur()
		m.clearSearch()
		return m, nil
	}

	// Forward input to textinput
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)

	// Re-search on each keystroke
	m.performSearch()

	return m, cmd
}

// handleSearchResultsKey handles keyboard input while navigating search results.
// Refresh is paused. n/N jump between matches with a focused highlight.
func (m Model) handleSearchResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.NextMatch):
		if len(m.searchMatches) > 0 {
			m.searchIndex = (m.searchIndex + 1) % len(m.searchMatches)
			m.updateFocusHighlight()
			m.scrollToMatch(m.searchIndex)
		}
		return m, nil

	case key.Matches(msg, m.keys.PrevMatch):
		if len(m.searchMatches) > 0 {
			m.searchIndex--
			if m.searchIndex < 0 {
				m.searchIndex = len(m.searchMatches) - 1
			}
			m.updateFocusHighlight()
			m.scrollToMatch(m.searchIndex)
		}
		return m, nil

	case key.Matches(msg, m.keys.Search):
		// Start a new search from search-results mode
		m.mode = modeSearch
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		return m, m.searchInput.Cursor.BlinkCmd()

	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit

	case msg.Type == tea.KeyEscape:
		// Exit search results — clear highlights, resume refresh
		m.mode = modeNormal
		m.clearSearch()
		return m, nil
	}

	// Allow viewport scrolling in search-results mode
	var cmd tea.Cmd
	if key.Matches(msg, m.keys.Up) {
		m.viewport.ScrollUp(1)
	} else if key.Matches(msg, m.keys.Down) {
		m.viewport.ScrollDown(1)
	} else if key.Matches(msg, m.keys.PageUp) {
		m.viewport.HalfPageUp()
	} else if key.Matches(msg, m.keys.PageDown) {
		m.viewport.HalfPageDown()
	} else if key.Matches(msg, m.keys.Home) {
		m.viewport.GotoTop()
	} else if key.Matches(msg, m.keys.End) {
		m.viewport.GotoBottom()
	}

	return m, cmd
}

// handleQueryKey handles keyboard input in query mode.
func (m Model) handleQueryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Apply the new query, trigger a re-fetch
		newQuery := m.queryInput.Value()
		m.currentQuery = newQuery
		if m.queryUpdater != nil {
			m.queryUpdater(newQuery)
		}
		m.mode = modeNormal
		m.queryInput.Blur()
		m.loading = true
		return m, fetchData(m.dataFetcher)

	case tea.KeyEscape:
		// Cancel — discard changes, return to normal
		m.mode = modeNormal
		m.queryInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.queryInput, cmd = m.queryInput.Update(msg)
	return m, cmd
}

// performSearch updates search state based on the current search input.
// During typing (modeSearch) all matches use the same highlight — no focus yet.
func (m *Model) performSearch() {
	term := m.searchInput.Value()
	m.searchTerm = term

	if term == "" {
		m.searchMatches = nil
		m.searchIndex = -1
		m.highlightedContent = ""
		m.viewport.SetContent(m.content)
		return
	}

	m.searchMatches = findMatches(m.content, term)
	m.highlightedContent = highlightContent(m.content, term)
	m.viewport.SetContent(m.highlightedContent)

	if len(m.searchMatches) > 0 {
		m.searchIndex = 0
		m.scrollToMatch(0)
	} else {
		m.searchIndex = -1
	}
}

// reapplySearch re-highlights search matches after content refresh.
func (m *Model) reapplySearch() {
	if m.searchTerm == "" {
		m.viewport.SetContent(m.content)
		return
	}

	m.searchMatches = findMatches(m.content, m.searchTerm)

	// Clamp searchIndex to new match count
	if m.searchIndex >= len(m.searchMatches) {
		if len(m.searchMatches) > 0 {
			m.searchIndex = len(m.searchMatches) - 1
		} else {
			m.searchIndex = -1
		}
	}

	if m.mode == modeSearchResults {
		m.updateFocusHighlight()
	} else {
		m.highlightedContent = highlightContent(m.content, m.searchTerm)
		m.viewport.SetContent(m.highlightedContent)
	}
}

// updateFocusHighlight re-renders the content with the current match visually focused.
func (m *Model) updateFocusHighlight() {
	m.highlightedContent = highlightContentWithFocus(m.content, m.searchTerm, m.searchMatches, m.searchIndex)
	m.viewport.SetContent(m.highlightedContent)
}

// clearSearch resets all search state and restores the original content.
func (m *Model) clearSearch() {
	m.searchTerm = ""
	m.searchMatches = nil
	m.searchIndex = -1
	m.highlightedContent = ""
	m.viewport.SetContent(m.content)
}

// scrollToMatch scrolls the viewport to the line containing match at the given index.
func (m *Model) scrollToMatch(index int) {
	line := lineForMatch(m.searchMatches, index)
	m.viewport.SetYOffset(line)
}

// handleWindowResize handles terminal window resize events
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	headerHeight := 0
	footerHeight := 2 // Status bar + spacing
	verticalMarginHeight := headerHeight + footerHeight

	if !m.ready {
		m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
		m.viewport.YPosition = headerHeight
		m.viewport.SetContent(m.content)
	} else {
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight
	}

	m.help.Width = msg.Width

	return m, nil
}

// RunWithOptions starts the TUI program with optional configuration.
func RunWithOptions(dataFetcher DataFetcher, refreshInterval time.Duration, opts ...Option) error {
	model := NewModel(dataFetcher, refreshInterval, opts...)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

// Run starts the TUI program (backward-compatible wrapper).
func Run(dataFetcher DataFetcher, refreshInterval time.Duration) error {
	return RunWithOptions(dataFetcher, refreshInterval)
}
