package cli

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

func (this cliModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return this, tea.Quit

		case key.Matches(msg, keys.Select):
			item, ok := this.filesList.SelectedItem().(Option[string])
			if ok {
				return this.selectItem(item), nil
			}

		case key.Matches(msg, keys.Cancel):
			if state.Get() == Running {
				stop()
			}

		case key.Matches(msg, keys.LogUp):
			this.viewport.LineUp(1)

		case key.Matches(msg, keys.LogDown):
			this.viewport.LineDown(1)

		case key.Matches(msg, keys.Help):
			this.help.ShowAll = !this.help.ShowAll
		}

	case tea.WindowSizeMsg:
		return this.resizeElements(msg.Width, msg.Height), nil

	case tickMsg:
		if logs.HasNewLogs() {
			this.viewport.SetContent(strings.Join(logs.Get(), "\n"))
			this.viewport.GotoBottom()
		}
		return this, tea.Batch(cmd, tickCmd())

	}

	this.filesList, cmd = this.filesList.Update(msg)
	cmds = append(cmds, cmd)
	this.spinner, cmd = this.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return this, tea.Batch(cmds...)
}
