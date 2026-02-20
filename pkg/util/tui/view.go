package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Styles for the TUI
var (
	// Status bar styles
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("63")).
			Padding(0, 1)

	statusBarErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("196")).
				Padding(0, 1)

	// Search/query input bar style
	inputBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("25")).
			Padding(0, 1)

	// Search results navigation bar style
	searchResultsBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("136")).
				Padding(0, 1)

	// Help styles
	helpStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	helpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63")).
			MarginBottom(1)

	// Spinner style
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
)

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if !m.ready {
		return fmt.Sprintf("\n  %s Loading...\n\n", m.spinner.View())
	}

	var b strings.Builder

	// Render the main content viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Render help overlay if visible
	if m.showHelp {
		b.WriteString("\n")
		b.WriteString(m.renderHelpOverlay())
		b.WriteString("\n")
	}

	// Render bottom bar based on current mode
	switch m.mode {
	case modeSearch:
		b.WriteString(m.renderSearchBar())
	case modeSearchResults:
		b.WriteString(m.renderSearchResultsBar())
	case modeQuery:
		b.WriteString(m.renderQueryBar())
	default:
		b.WriteString(m.renderStatusBar())
	}

	return b.String()
}

// renderSearchBar renders the search input bar at the bottom.
func (m Model) renderSearchBar() string {
	matchInfo := ""
	if m.searchInput.Value() != "" {
		total := len(m.searchMatches)
		if total > 0 {
			matchInfo = fmt.Sprintf(" [%d/%d]", m.searchIndex+1, total)
		} else {
			matchInfo = " [no match]"
		}
	}

	text := "/" + m.searchInput.View() + matchInfo
	return inputBarStyle.Width(m.width).Render(text)
}

// renderSearchResultsBar renders the bar shown while navigating search results.
func (m Model) renderSearchResultsBar() string {
	total := len(m.searchMatches)
	pos := ""
	if total > 0 {
		pos = fmt.Sprintf("[%d/%d]", m.searchIndex+1, total)
	}
	text := fmt.Sprintf("/%s  %s  (n/N: next/prev  Esc: exit search  /: new search)", m.searchTerm, pos)
	return searchResultsBarStyle.Width(m.width).Render(text)
}

// renderQueryBar renders the query input bar at the bottom.
func (m Model) renderQueryBar() string {
	text := ":" + m.queryInput.View()
	return inputBarStyle.Width(m.width).Render(text)
}

// renderStatusBar renders the status bar at the bottom
func (m Model) renderStatusBar() string {
	var parts []string

	// Loading indicator
	if m.loading {
		parts = append(parts, m.spinner.View()+" Refreshing...")
	}

	// Last update time
	elapsed := time.Since(m.lastUpdate)
	if elapsed < time.Minute {
		parts = append(parts, fmt.Sprintf("Updated %ds ago", int(elapsed.Seconds())))
	} else {
		parts = append(parts, fmt.Sprintf("Updated %s ago", elapsed.Round(time.Second)))
	}

	// Refresh interval
	parts = append(parts, fmt.Sprintf("Refresh: %ds", int(m.refreshInterval.Seconds())))

	// Search match count (persists in normal mode after confirmed search)
	if m.searchTerm != "" && len(m.searchMatches) > 0 {
		parts = append(parts, fmt.Sprintf("Match: %d/%d", m.searchIndex+1, len(m.searchMatches)))
	}

	// Active query indicator
	if m.currentQuery != "" && m.queryUpdater != nil {
		parts = append(parts, fmt.Sprintf("Query: %s", m.currentQuery))
	}

	// Scroll position hint
	scrollPercent := m.viewport.ScrollPercent()
	if scrollPercent > 0 || scrollPercent < 1 {
		parts = append(parts, fmt.Sprintf("Scroll: %d%%", int(scrollPercent*100)))
	}

	// Quick help hint
	parts = append(parts, "Press ? for help")

	statusText := strings.Join(parts, " • ")

	// Use error style if there's an error, otherwise normal style
	style := statusBarStyle
	if m.lastError != nil {
		style = statusBarErrorStyle
		errorMsg := fmt.Sprintf("Error: %v", m.lastError)
		if len(errorMsg) > m.width-10 {
			errorMsg = errorMsg[:m.width-13] + "..."
		}
		statusText = errorMsg + " • " + statusText
	}

	// Ensure the status bar fits the width
	if len(statusText) > m.width-4 {
		statusText = statusText[:m.width-7] + "..."
	}

	return style.Width(m.width).Render(statusText)
}

// renderHelpOverlay renders the help panel overlay
func (m Model) renderHelpOverlay() string {
	helpContent := helpTitleStyle.Render("Keyboard Shortcuts") + "\n\n"
	helpContent += m.help.View(m.keys)

	helpContent += "\n\n"

	tips := "TIP: Use +/- to adjust refresh interval\n"
	if m.queryUpdater != nil {
		tips += "Use : to enter a TSL query filter\n"
	}
	tips += "Press ? again to close this help"

	helpContent += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(tips)

	// Center the help panel
	helpBox := helpStyle.Render(helpContent)

	helpWidth := lipgloss.Width(helpBox)
	helpHeight := lipgloss.Height(helpBox)

	horizontalMargin := (m.width - helpWidth) / 2
	verticalMargin := (m.height - helpHeight - 3) / 2

	if horizontalMargin < 0 {
		horizontalMargin = 0
	}
	if verticalMargin < 0 {
		verticalMargin = 0
	}

	centeredHelp := lipgloss.NewStyle().
		MarginLeft(horizontalMargin).
		MarginTop(verticalMargin).
		Render(helpBox)

	return centeredHelp
}
