package cli

import (
	"math"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

func (c *Cli) toggleProgress() {
	c.showProgress = true
	c.completed = 0
	c.errs = nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(10*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// nolint: funlen
func (c *Cli) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.keys.Quit):
			return c, tea.Quit
		case key.Matches(msg, c.keys.Select):
			item, ok := c.filesList.SelectedItem().(Option[string])
			if ok {
				c.file = item.Title
				c.toggleProgress()
				go c.execRequests(item.Value)
			}
			return c, nil
		}

	case tea.WindowSizeMsg:
		lw := int(math.Round(float64(msg.Width) * 0.4))
		pq := msg.Width - lw + 4
		c.filesList.SetWidth(lw)
		c.progressBar.Width = pq
		return c, nil

	case tickMsg:
		cmd := c.progressBar.SetPercent(c.completed)
		return c, tea.Batch(tickCmd(), cmd)

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
