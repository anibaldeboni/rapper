package cli

import (
	"fmt"

	"github.com/anibaldeboni/rapper/cli/ui"

	"github.com/charmbracelet/lipgloss"
)

func (c cliModel) View() string {
	var progress string
	if state.Get() == Running || state.Get() == Stale {
		progress = lipgloss.NewStyle().
			PaddingLeft(2).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Top,
					viewPortTitle,
					ui.ViewPortStyle(c.viewport.View()),
					c.progressBar.View(),
				),
			)
	}

	widgets := lipgloss.JoinHorizontal(
		lipgloss.Left,
		c.filesList.View(),
		progress,
	)
	help := lipgloss.JoinHorizontal(
		lipgloss.Left,
		ui.LogoStyle(fmt.Sprintf("%s@%s", AppName, AppVersion)),
		ui.HelpStyle(c.help.View(keys)),
	)
	app := lipgloss.JoinVertical(
		lipgloss.Top,
		widgets,
		help,
	)

	return ui.AppStyle(app)
}
