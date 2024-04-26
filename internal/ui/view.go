package ui

import (
	"fmt"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

func buildViewport(model string) string {
	return lipgloss.NewStyle().
		PaddingLeft(2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				viewPortTitle,
				styles.ViewPortStyle(model),
			),
		)
}

func (this Model) View() string {
	var executionLogs string
	if state.Get() == Running || state.Get() == Stale {
		executionLogs = buildViewport(this.viewport.View())
	}

	widgets := lipgloss.JoinHorizontal(
		lipgloss.Left,
		this.filesList.View(),
		executionLogs,
	)
	width := lipgloss.Width
	appTag := styles.LogoStyle(fmt.Sprintf("%s@%s", AppName, AppVersion))
	var spinner string
	if state.Get() == Running {
		spinner = this.spinner.View()
	} else {
		spinner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).Render("∙∙∙")
	}

	help := lipgloss.NewStyle().
		Width(this.width - width(appTag) - width(spinner) - 4).
		PaddingLeft(1).
		Render(truncate.StringWithTail(this.help.View(keys), uint(this.width-width(appTag)-width(spinner)), "…"))

	statusbar := lipgloss.JoinHorizontal(
		lipgloss.Left,
		appTag,
		help,
		spinner,
	)
	app := lipgloss.JoinVertical(
		lipgloss.Top,
		widgets,
		statusbar,
	)

	return styles.AppStyle(app)
}
