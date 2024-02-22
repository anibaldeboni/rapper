package cli

import (
	"fmt"

	"github.com/anibaldeboni/rapper/cli/ui"

	"github.com/charmbracelet/lipgloss"
)

func (c *Cli) View() string {
	var progress string
	if c.showProgress {
		progress = lipgloss.JoinVertical(
			lipgloss.Top,
			c.logView(),
			c.progressBar.View(),
		)
	}

	widgets := lipgloss.JoinHorizontal(
		lipgloss.Top,
		ui.ListStyle(c.filesList.View()),
		ui.ProgressStyle(progress),
	)

	app := lipgloss.JoinVertical(
		lipgloss.Top,
		fmt.Sprintf("[ %s @ %s ]\n", ui.Bold(name), ui.Pink(version)),
		widgets,
		c.help.View(c.keys),
	)

	return ui.AppStyle(app)
}
