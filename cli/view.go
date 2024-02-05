package cli

import (
	"fmt"
	"rapper/cli/ui"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func icon(errs []string) string {
	var icon string
	if len(errs) > 0 {
		icon = ui.IconError
	} else {
		icon = ui.IconTrophy
	}
	return icon
}

func (c *Cli) View() string {
	errs := strings.Join(c.errs, "\n")

	var status string
	if c.isProcessing() {
		status = fmt.Sprintf("\n%s Processing %s", ui.IconWomanDancing, ui.Green(c.file))
	} else {
		status = fmt.Sprintf("\n%s %s done!", icon(c.errs), ui.Pink(c.file))
		c.ctx = nil
		c.cancelFn = nil
	}

	var progress string
	if c.showProgress {
		progress = lipgloss.JoinVertical(
			lipgloss.Top,
			ui.TitleStyle.Render("Status"),
			c.progressBar.View(),
			status,
			ui.ErrStyle(errs),
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
