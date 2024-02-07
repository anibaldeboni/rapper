package cli

import (
	"fmt"
	"github.com/anibaldeboni/rapper/cli/ui"
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
	var progress string
	if c.showProgress {
		el := []string{
			ui.TitleStyle.Render("Status"),
			c.progressBar.View(),
		}
		if c.isProcessing() {
			el = append(el, fmt.Sprintf("\n%s Processing %s", ui.IconWomanDancing, ui.Green(c.file)))
		} else {
			el = append(el, fmt.Sprintf("\n%s %s done!", icon(c.errs), ui.Pink(c.file)))
			c.done()
		}
		if len(c.errs) > 0 {
			el = append(el, "\n"+strings.Join(c.errs, "\n")+"\n")
		}
		if c.alert != "" {
			el = append(el, fmt.Sprintf("\n%s  %s\n", ui.IconInformation, c.alert))
		}
		progress = lipgloss.JoinVertical(
			lipgloss.Top,
			el...,
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
