package cli

import (
	"fmt"

	"github.com/anibaldeboni/rapper/cli/ui"

	"github.com/charmbracelet/lipgloss"
)

func (c cliImpl) View() string {
	var progress string
	if state.Get() == Running || state.Get() == Stale {
		progress = lipgloss.JoinVertical(
			lipgloss.Top,
			fmt.Sprintf("%s\n%s\n\n", viewPortTitle, c.viewport.View()),
			c.progressBar.View(),
		)
	}

	widgets := lipgloss.JoinHorizontal(
		lipgloss.Top,
		c.filesList.View(),
		ui.ProgressStyle(progress),
	)

	app := lipgloss.JoinVertical(
		lipgloss.Top,
		fmt.Sprintf("[ %s @ %s ]\n", ui.Bold(AppName), ui.Pink(AppVersion)),
		widgets,
		c.help.View(keys),
	)

	return ui.AppStyle(app)
}
