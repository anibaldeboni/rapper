package cli

import (
	"fmt"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

func (c cliModel) View() string {
	var executionLogs string
	if state.Get() == Running || state.Get() == Stale {
		executionLogs = lipgloss.NewStyle().
			PaddingLeft(2).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Top,
					viewPortTitle,
					styles.ViewPortStyle(c.viewport.View()),
				),
			)
	}

	widgets := lipgloss.JoinHorizontal(
		lipgloss.Left,
		c.filesList.View(),
		executionLogs,
	)
	width := lipgloss.Width
	logo := styles.LogoStyle(fmt.Sprintf("%s@%s", AppName, AppVersion))
	var spinner string
	if state.Get() == Running {
		spinner = c.spinner.View()
	} else {
		spinner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).Render("∙∙∙")
	}

	help := lipgloss.NewStyle().
		Height(1).
		Width(c.width - width(logo) - width(spinner) - 3).
		PaddingLeft(1).
		Render(truncate.StringWithTail(c.help.View(keys), uint(c.width-width(logo)-width(spinner)), "…"))

	statusbar := lipgloss.JoinHorizontal(
		lipgloss.Left,
		logo,
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
