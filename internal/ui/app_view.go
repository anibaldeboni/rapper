package ui

import (
	"fmt"

	"github.com/ccoveille/go-safecast"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

func (m AppModel) View() string {
	// Render header with contextual help
	header := m.renderHeader()

	// Render based on current view
	var content string

	switch m.nav.Current() {
	case ViewFiles:
		content = m.renderFilesView()
	case ViewLogs:
		content = m.renderLogsView()
	case ViewSettings:
		content = m.settingsView.View()
	case ViewWorkers:
		content = m.workersView.View()
	}

	// Render toasts if any
	toasts := m.toastMgr.Render(m.width)

	// Render status bar
	statusBar := m.renderStatusBar()

	// Join all elements: header, toasts (if any), content, and status bar
	if toasts != "" {
		app := lipgloss.JoinVertical(
			lipgloss.Top,
			header,
			toasts,
			content,
			statusBar,
		)
		return AppStyle(app)
	}

	// Join header, content and status bar
	app := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		content,
		statusBar,
	)

	return AppStyle(app)
}

func (m AppModel) renderFilesView() string {
	// Show files list on the left
	filesWidget := m.filesView.View()

	// Show logs on the right if processing is running or complete
	var logsWidget string
	if m.state.Get() == Running || m.state.Get() == Stale {
		logsWidget = m.logsView.View()
	}

	// Join horizontally if we have logs
	if logsWidget != "" {
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			filesWidget,
			logsWidget,
		)
	}

	return filesWidget
}

func (m AppModel) renderLogsView() string {
	return m.logsView.View()
}

// renderHeader renders the contextual help bar at the top
func (m AppModel) renderHeader() string {
	// Get contextual keymap for current view
	contextualKeys := getContextualKeyMap(m.nav.Current())

	// Discrete style for help bar
	helpBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	helpText := m.help.View(contextualKeys)

	// Truncate if needed
	maxWidth := m.width - 4
	helpEmptySpace, err := safecast.ToUint(maxWidth)
	if err != nil {
		helpEmptySpace = 0
	}

	truncatedHelp := truncate.StringWithTail(helpText, helpEmptySpace, "…")

	return helpBarStyle.Width(m.width).Render(truncatedHelp)
}

// renderStatusBar renders the simplified status bar at the bottom
func (m AppModel) renderStatusBar() string {
	width := lipgloss.Width

	// App tag
	appTag := LogoStyle(fmt.Sprintf("%s@%s", AppName, AppVersion))

	// Current view indicator
	viewName := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d6acff")).
		Render(fmt.Sprintf(" [%s] ", m.nav.Current().String()))

	// Spinner or idle indicator
	var spinner string
	if m.state.Get() == Running {
		spinner = m.spinner.View()
	} else {
		spinner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Render("∙∙∙")
	}

	// Calculate spacing to push spinner to the right
	spacing := m.width - width(appTag) - width(viewName) - width(spinner) - 4
	if spacing < 0 {
		spacing = 0
	}

	spacer := lipgloss.NewStyle().Width(spacing).Render("")

	// Join all status bar elements
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		appTag,
		viewName,
		spacer,
		spinner,
	)
}
