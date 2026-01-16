package ui

import (
	"fmt"

	"github.com/ccoveille/go-safecast"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

func (m AppModel) View() string {
	// Render based on current view
	var content string

	switch m.nav.Current() {
	case ViewFiles:
		content = m.filesView.View()
	case ViewLogs:
		content = m.logsView.View()
	case ViewSettings:
		content = m.settingsView.View()
	case ViewWorkers:
		content = m.workersView.View()
	}

	// Render toasts if any
	toasts := m.toastMgr.Render()

	// Join all elements: header, toasts (if any), content, and status bar
	app := lipgloss.NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height).
		AlignVertical(lipgloss.Center).
		Margin(0, 2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				m.renderHeader(),
				lipgloss.JoinHorizontal(
					lipgloss.Top,
					lipgloss.Place(
						(m.width/2)-2,
						m.height-4,
						lipgloss.Left,
						lipgloss.Top,
						content,
					),
					lipgloss.Place(
						(m.width/2)-2,
						m.height-4,
						lipgloss.Right,
						lipgloss.Top,
						toasts,
					),
				),
				m.renderStatusBar(),
			),
		)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		app,
	)
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
