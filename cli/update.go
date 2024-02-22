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

func (c *Cli) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.keys.Quit):
			return c, tea.Quit

		case key.Matches(msg, c.keys.Select):
			item, ok := c.filesList.SelectedItem().(Option[string])
			if ok {
				c.selectItem(item)
			}
			return c, nil

		case key.Matches(msg, c.keys.Cancel):
			if c.ctx != nil {
				c.cancel()
			}
			return c, nil

		case key.Matches(msg, c.keys.LogUp):
			c.viewport.LineUp(1)
			return c, nil

		case key.Matches(msg, c.keys.LogDown):
			c.viewport.LineDown(1)
			return c, nil
		}

	case tea.WindowSizeMsg:
		c.resizeElements(msg.Width, msg.Height)
		return c, nil

	case tickMsg:
		c.viewport.SetContent(strings.Join(c.logs, "\n"))
		cmd = c.progressBar.SetPercent(c.completed)
		return c, tea.Batch(cmd, tickCmd())

	case progress.FrameMsg:
		progressModel, cmd := c.progressBar.Update(msg)
		p, ok := progressModel.(progress.Model)
		if ok {
			c.progressBar = p
		}
		return c, cmd
	}

	if c.ctx != nil {
		c.viewport.GotoBottom()
	}

	c.filesList, cmd = c.filesList.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}
