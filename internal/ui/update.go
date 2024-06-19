package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmdfls tea.Cmd
		cmdspp tea.Cmd
		cmds   []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Select):
			item, ok := m.filesList.SelectedItem().(Option[string])
			if ok {
				return m.selectItem(item), nil
			}

		case key.Matches(msg, keys.Cancel):
			if state.Get() == Running {
				cancel()
			}

		case key.Matches(msg, keys.LogUp):
			m.viewport.LineUp(1)

		case key.Matches(msg, keys.LogDown):
			m.viewport.LineDown(1)

		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}

	case tea.WindowSizeMsg:
		return m.resizeElements(msg.Width, msg.Height), nil

	case tickMsg:
		m.viewport.SetContent(strings.Join(logger.Get(), "\n"))
		m.viewport.GotoBottom()
		cmds = append(cmds, tickCmd())
	}

	m.filesList, cmdfls = m.filesList.Update(msg)
	m.spinner, cmdspp = m.spinner.Update(msg)
	cmds = append(cmds, cmdfls, cmdspp)

	return m, tea.Batch(cmds...)
}
