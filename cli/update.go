package cli

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (c cliModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return c, tea.Quit

		case key.Matches(msg, keys.Select):
			item, ok := c.filesList.SelectedItem().(Option[string])
			if ok {
				return c.selectItem(item), nil
			}

		case key.Matches(msg, keys.Cancel):
			if state.Get() == Running {
				stop()
			}

		case key.Matches(msg, keys.LogUp):
			c.viewport.LineUp(1)

		case key.Matches(msg, keys.LogDown):
			c.viewport.LineDown(1)

		case key.Matches(msg, keys.Help):
			c.help.ShowAll = !c.help.ShowAll
		}

	case tea.WindowSizeMsg:
		return c.resizeElements(msg.Width, msg.Height), nil

	case tickMsg:
		cmd = c.progressBar.SetPercent(completed)
		if logs.HasNewLogs() {
			c.viewport.SetContent(strings.Join(logs.Get(), "\n"))
			c.viewport.GotoBottom()
		}
		return c, tea.Batch(cmd, tickCmd())

	case progress.FrameMsg:
		progressModel, cmd := c.progressBar.Update(msg)
		p, ok := progressModel.(progress.Model)
		if ok {
			c.progressBar = p
		}
		return c, cmd
	}

	c.filesList, cmd = c.filesList.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}
