package cli

import (
	"fmt"

	"github.com/anibaldeboni/rapper/cli/ui"

	"github.com/charmbracelet/lipgloss"
)

func (c cliModel) View() string {
	var progress string
	if state.Get() == Running || state.Get() == Stale {
		progress = lipgloss.JoinVertical(
			lipgloss.Top,
			ui.TitleStyle.Render(viewPortTitle),
			ui.ViewPortStyle(c.viewport.View()),
			c.progressBar.View(),
		)
	}

	widgets := lipgloss.JoinHorizontal(
		lipgloss.Left,
		c.filesList.View(),
		ui.ProgressStyle(progress),
	)
	help := lipgloss.JoinHorizontal(
		lipgloss.Left,
		fmt.Sprintf("%s@%s: ", ui.Bold(AppName), ui.Pink(AppVersion)),
		c.help.View(keys),
	)
	app := lipgloss.JoinVertical(
		lipgloss.Top,
		widgets,
		help,
	)

	return ui.AppStyle(app)
}
