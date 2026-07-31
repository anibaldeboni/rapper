package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ccoveille/go-safecast"
	"github.com/muesli/reflow/truncate"
)

var uiLogoStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#F25D94")).
	Foreground(lipgloss.Color("#ffffff")).
	Bold(true).
	Padding(0, 1).
	Render

func (m AppModel) View() tea.View {
	var content string
	if v, ok := m.views[m.currentView]; ok && v != nil {
		content = v.View().Content
	}

	app := lipgloss.NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height).
		AlignVertical(lipgloss.Top).
		Margin(1, 2).
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

func (m AppModel) renderHeader() string {
	globalKeys := getGlobalKeyMap()
	helpBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("235"))

	viewName := uiLogoStyle(m.currentView.String())
	helpText := viewName + " " + m.help.View(globalKeys)

	maxWidth := m.width - 4
	helpEmptySpace, err := safecast.ToUint(maxWidth)
	if err != nil {
		helpEmptySpace = 0
	}

	truncatedHelp := truncate.StringWithTail(helpText, helpEmptySpace, "…")

	return helpBarStyle.Width(m.width).Render(truncatedHelp)
}

func (m AppModel) renderStatusBar() string {
	width := lipgloss.Width

	appName := uiLogoStyle(fmt.Sprintf("%s@%s", AppName, AppVersion))

	helpText := m.help.View(m.currentStatusKeymap)

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

	var spinner string
	if m.processor.GetMetrics().IsProcessing {
		spinner = m.spinner.View()
	} else {
		spinner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Render("∙∙∙")
	}

	spacing := max(m.width-width(appName)-width(contextHelp)-width(spinner)-4, 0)

	spacer := lipgloss.NewStyle().Width(spacing).Render("")

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		appName,
		contextHelp,
		spacer,
		spinner,
	)
}
