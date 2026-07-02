package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ccoveille/go-safecast"
	"github.com/muesli/reflow/truncate"
)

func (m AppModel) View() tea.View {
	// Render the active view into the full content area; each view owns its
	// own internal layout (see views-own-layout decision, Engram #46).
	var content string
	if v, ok := m.views[m.currentView]; ok && v != nil {
		content = v.View().Content
	}

	app := lipgloss.NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height).
		AlignVertical(lipgloss.Top).
		Margin(0, 2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				m.renderHeader(),
				content,
				m.renderStatusBar(),
			),
		)

	bgLayer := lipgloss.NewLayer(app).Z(0)
	toastLayers := m.toastMgr.Layers(m.width)

	var framed string
	if len(toastLayers) == 0 {
		framed = app
	} else {
		compositor := lipgloss.NewCompositor(bgLayer)
		compositor.AddLayers(toastLayers...)
		framed = compositor.Render()
	}

	v := tea.NewView(lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Top,
		lipgloss.Top,
		framed,
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
	viewName := LogoStyle(m.currentView.String())
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
	viewSpecificKeys := getViewSpecificKeyMap(m.currentView)
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
