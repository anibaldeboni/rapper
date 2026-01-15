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

	// Join toasts (if any), content, and status bar
	if toasts != "" {
		app := lipgloss.JoinVertical(
			lipgloss.Top,
			toasts,
			content,
			statusBar,
		)
		return AppStyle(app)
	}

	// Join content and status bar
	app := lipgloss.JoinVertical(
		lipgloss.Top,
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

func (m AppModel) renderStatusBar() string {
	width := lipgloss.Width

	// App tag
	appTag := LogoStyle(fmt.Sprintf("%s@%s", AppName, AppVersion))

	// Spinner or idle indicator
	var spinner string
	if m.state.Get() == Running {
		spinner = m.spinner.View()
	} else {
		spinner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Render("∙∙∙")
	}

	// Current view indicator
	viewName := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d6acff")).
		Render(fmt.Sprintf(" [%s] ", m.nav.Current().String()))

	// Help text
	helpEmptySpace, err := safecast.ToUint(m.width - width(appTag) - width(spinner) - width(viewName) - 4)
	if err != nil {
		helpEmptySpace = 0
	}

	helpText := lipgloss.NewStyle().
		Width(m.width - width(appTag) - width(spinner) - width(viewName) - 4).
		PaddingLeft(1).
		Render(truncate.StringWithTail(m.help.View(keys), helpEmptySpace, "…"))

	// Join all status bar elements
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		appTag,
		viewName,
		helpText,
		spinner,
	)
}
