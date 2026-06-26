package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ccoveille/go-safecast"
	"github.com/muesli/reflow/truncate"
)

// toastOverlayWidth is the width reserved for the top-right toast corner
// overlay. The metric panel in the Logs view uses ~28 columns by default
// so 40 leaves a small visual gap between the two top-right elements.
const toastOverlayWidth = 40

func (m AppModel) View() tea.View {
	// Render the active view into the full content area; each view owns its
	// own internal layout (see views-own-layout decision, Engram #46).
	var content string

	switch m.nav.Current() {
	case ViewFiles:
		content = m.filesView.View()
	case ViewLogs:
		content = m.logsView.View()
	case ViewSettings:
		content = m.settingsView.View()
	}

	// Join all elements: header, content, and status bar. Toasts are layered
	// on top of the content via line-splice rather than as a separate
	// column so the active view can use the full content width.
	app := lipgloss.NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height).
		AlignVertical(lipgloss.Center).
		Margin(0, 2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				m.renderHeader(),
				content,
				m.renderStatusBar(),
			),
		)

	// Layer toast overlay on top of the rendered app string when there are
	// active toasts. spliceRight replaces the rightmost characters of the
	// first N content lines with the overlay lines.
	if overlay := m.toastMgr.RenderOverlay(toastOverlayWidth); overlay != "" {
		app = spliceRight(app, overlay, m.width)
	}

	v := tea.NewView(lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		app,
	))
	v.AltScreen = true
	v.ReportFocus = true
	v.WindowTitle = fmt.Sprintf("%s@%s", AppName, AppVersion)
	v.KeyboardEnhancements.ReportEventTypes = true
	return v
}

// renderHeader renders the global navigation help bar at the top
func (m AppModel) renderHeader() string {
	// Get global keymap (F1-F4, q)
	globalKeys := getGlobalKeyMap()

	// Discrete style for help bar
	helpBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("235"))

		// App tag
	viewName := LogoStyle(m.nav.Current().String())
	helpText := viewName + " " + m.help.View(globalKeys)

	// Truncate if needed
	maxWidth := m.width - 4
	helpEmptySpace, err := safecast.ToUint(maxWidth)
	if err != nil {
		helpEmptySpace = 0
	}

	truncatedHelp := truncate.StringWithTail(helpText, helpEmptySpace, "…")

	return helpBarStyle.Width(m.width).Render(truncatedHelp)
}

// renderStatusBar renders the status bar with view-specific commands at the bottom
func (m AppModel) renderStatusBar() string {
	width := lipgloss.Width

	// Current view indicator
	appName := LogoStyle(fmt.Sprintf("%s@%s", AppName, AppVersion))

	// Get view-specific commands
	viewSpecificKeys := getViewSpecificKeyMap(m.nav.Current())
	helpText := m.help.View(viewSpecificKeys)

	// Truncate help text if needed
	availableWidth := max(m.width-lipgloss.Width(appName)-10, 0)
	helpEmptySpace, err := safecast.ToUint(availableWidth)
	if err != nil {
		helpEmptySpace = 0
	}
	truncatedHelp := truncate.StringWithTail(helpText, helpEmptySpace, "…")

	contextHelp := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		PaddingLeft(1).
		Render(truncatedHelp)

	// Spinner or idle indicator
	var spinner string
	if m.processor.GetMetrics().IsProcessing {
		spinner = m.spinner.View()
	} else {
		spinner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Render("∙∙∙")
	}

	// Calculate spacing to push spinner to the right
	spacing := max(m.width-width(appName)-width(contextHelp)-width(spinner)-4, 0)

	spacer := lipgloss.NewStyle().Width(spacing).Render("")

	// Join all status bar elements
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		appName,
		contextHelp,
		spacer,
		spinner,
	)
}

// spliceRight overlays the given text on the right side of the base string.
// Both inputs are multi-line. The overlay's first line replaces the rightmost
// N characters of the base's first line, the second line replaces the right
// of the base's second line, and so on. Lines beyond the overlay length are
// untouched. Returns the base string when either input is empty.
func spliceRight(base, overlay string, _ int) string {
	if base == "" || overlay == "" {
		return base
	}

	overlayLines := strings.Split(overlay, "\n")
	baseLines := strings.Split(base, "\n")
	if len(baseLines) == 0 {
		return base
	}

	for i, overlayLine := range overlayLines {
		if i >= len(baseLines) {
			break
		}
		baseLine := baseLines[i]
		overlayWidth := lipgloss.Width(overlayLine)
		baseWidth := lipgloss.Width(baseLine)
		// Trim the base line so that the overlay fits on the right.
		if overlayWidth >= baseWidth {
			baseLines[i] = overlayLine
			continue
		}
		// Keep the left portion of the base line and append the overlay.
		// lipgloss Width returns visible columns ignoring ANSI escapes so we
		// need to slice on visible width; using a simple approach: keep
		// baseWidth-overlayWidth characters of the visible base.
		keep := baseWidth - overlayWidth
		baseLines[i] = truncateVisibleLeft(baseLine, keep) + overlayLine
	}
	return strings.Join(baseLines, "\n")
}

// truncateVisibleLeft keeps the first n visible characters of the input and
// returns the result as a string with ANSI escapes preserved. It uses a
// simple rune walk which is correct for ASCII labels and emoji-free text
// (which is the only thing we splice through here — view content has its
// own ANSI handling).
func truncateVisibleLeft(s string, n int) string {
	if n <= 0 {
		return ""
	}
	count := 0
	runes := []rune(s)
	for i := range runes {
		if count >= n {
			return string(runes[:i])
		}
		count++
	}
	return s
}
