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
// overlay. The metric panel in the Logs view uses ~36 columns by default
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
// The base is assumed to be the global app layout produced by View(),
// with the header on line 0 and the status bar on the last line; the
// overlay is stacked on the right edge of the content area in between
// (lines 1..len-2). Toasts never overwrite the header or the status bar.
// Returns the base string when either input is empty, when width is
// non-positive, or when the base has no content area to overlay onto.
//
// The overlay is anchored to a fixed x-position computed from the
// terminal width (not the base line width) so that all overlay lines
// end at the same column from the right, regardless of how short the
// underlying base line is. This prevents the visual bug where a short
// middle content line caused its overlay to collapse to the LEFT edge
// while the top and bottom lines were correctly on the RIGHT.
func spliceRight(base, overlay string, width int) string {
	if base == "" || overlay == "" || width <= 0 {
		return base
	}

	baseLines := strings.Split(base, "\n")
	if len(baseLines) < 3 {
		// No content area between header and status bar: keep the
		// header and status bar untouched and drop the overlay.
		return base
	}

	// Reserve line 0 for the header and line len(baseLines)-1 for the
	// status bar. The overlay is spliced onto the lines in between so
	// every toast lands in the content area, stacked at the top.
	contentStart := 1
	contentEnd := len(baseLines) - 1
	overlayLines := strings.Split(overlay, "\n")

	for i, overlayLine := range overlayLines {
		target := contentStart + i
		if target >= contentEnd {
			// No more content lines available; drop extra overlay
			// lines instead of overwriting the status bar.
			break
		}
		baseLine := baseLines[target]
		overlayWidth := lipgloss.Width(overlayLine)

		// Compute the x-position from the RIGHT edge of the terminal.
		// A small right margin keeps the overlay from touching the
		// screen edge. This is independent of the base line width so
		// every overlay line ends at the same column.
		const rightMargin = 2
		targetX := width - overlayWidth - rightMargin
		if targetX < 0 {
			// Overlay is wider than the available space. Place it
			// at column 0 and let the terminal wrap/clip naturally.
			targetX = 0
		}

		// Keep the left `targetX` visible columns of the base line
		// (so the base appears to the left of the overlay) and pad
		// with spaces to exactly `targetX` visible cols. Padding is
		// required so that short base lines still anchor the overlay
		// to the right edge — otherwise a 17-col base would leave
		// the overlay at column 17 instead of column 78.
		truncated := truncateVisibleLeft(baseLine, targetX)
		baseVisibleWidth := lipgloss.Width(truncated)
		padding := strings.Repeat(" ", targetX-baseVisibleWidth)
		baseLines[target] = truncated + padding + overlayLine
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
