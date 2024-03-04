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

func (c cliImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.keyMap.Quit):
			return c, tea.Quit

		case key.Matches(msg, c.keyMap.Select):
			item, ok := c.filesList.SelectedItem().(Option[string])
			if ok {
				return c.selectItem(item), nil
			}

		case key.Matches(msg, c.keyMap.Cancel):
			if ctx != nil {
				cancel()
			}

		case key.Matches(msg, c.keyMap.LogUp):
			c.viewport.LineUp(1)

		case key.Matches(msg, c.keyMap.LogDown):
			c.viewport.LineDown(1)

		case key.Matches(msg, c.keyMap.Help):
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
