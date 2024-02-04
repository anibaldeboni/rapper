package cli

import (
	"math"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type progressMsg float64
type errorMsg string

func (c *Cli) toggleProgress() {
	c.showProgress = true
	c.progressBar.SetPercent(0)
	c.errs = nil
}

// nolint: funlen
func (c Cli) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case errorMsg:
		c.errs = append(c.errs, string(msg))
		return c, nil

	case progressMsg:
		c.showProgress = true
		return c, c.progressBar.SetPercent(float64(msg))

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
