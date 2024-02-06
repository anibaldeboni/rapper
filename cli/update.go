package cli

import (
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
		}

	case tea.WindowSizeMsg:
		c.resizeElements(msg.Width)
		return c, nil

	case tickMsg:
		cmd := c.progressBar.SetPercent(c.completed)
		return c, tea.Batch(cmd, tickCmd())

	case progress.FrameMsg:
		progressModel, cmd := c.progressBar.Update(msg)
		p, ok := progressModel.(progress.Model)
		if ok {
			c.progressBar = p
		}
		return c, cmd

	default:
		return c, nil
	}
	var cmd tea.Cmd
	c.filesList, cmd = c.filesList.Update(msg)
	return c, cmd
}
