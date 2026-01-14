package ui

import (
	"context"
	"time"

	"github.com/anibaldeboni/rapper/internal/ui/views"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global navigation keys
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.ViewFiles):
			m.nav.Set(ViewFiles)
			return m, nil

		case key.Matches(msg, keys.ViewLogs):
			m.nav.Set(ViewLogs)
			return m, nil

		case key.Matches(msg, keys.ViewSettings):
			m.nav.Set(ViewSettings)
			return m, nil

		case key.Matches(msg, keys.ViewWorkers):
			m.nav.Set(ViewWorkers)
			return m, nil

		case key.Matches(msg, keys.Cancel):
			if m.nav.Current() == ViewFiles {
				// Cancel running operation
				if m.state.Get() == Running {
					m.filesView.Cancel()
				}
			} else {
				// Go back to previous view
				m.nav.Back()
			}
			return m, nil

		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		// Delegate to current view
		return m.updateCurrentView(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeViews()
		return m, nil

	case tickMsg:
		// Update logs content
		m.logsView.UpdateContent()

		// Update spinner
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinCmd, tickCmd())
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) updateCurrentView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.nav.Current() {
	case ViewFiles:
		// Handle file selection
		if key.Matches(msg, keys.Select) {
			item := m.filesView.SelectedItem()
			if opt, ok := item.(views.Option[string]); ok {
				return m.selectFile(opt.Value), nil
			}
		}

		// Handle log scrolling in Files view (when logs are visible)
		if m.state.Get() == Running || m.state.Get() == Stale {
			if key.Matches(msg, keys.LogUp) {
				m.logsView.ScrollUp(1)
				return m, nil
			}
			if key.Matches(msg, keys.LogDown) {
				m.logsView.ScrollDown(1)
				return m, nil
			}
		}

		cmd = m.filesView.Update(msg)

	case ViewLogs:
		if key.Matches(msg, keys.LogUp) {
			m.logsView.ScrollUp(1)
		}
		if key.Matches(msg, keys.LogDown) {
			m.logsView.ScrollDown(1)
		}
		cmd = m.logsView.Update(msg)

	case ViewSettings:
		cmd = m.settingsView.Update(msg)

	case ViewWorkers:
		cmd = m.workersView.Update(msg)
	}

	return m, cmd
}

func (m AppModel) selectFile(filePath string) AppModel {
	if m.state.Get() != Running {
		m.state.Set(Stale)
		ctx, cancel := m.filesView.StartProcessing(context.Background(), filePath)
		if ctx != nil {
			m.state.Set(Running)
			m.cancel = cancel
			m.filesView.SetCancel(cancel)
			m.waitCompletion(ctx)

			// Switch to logs view when processing starts
			m.nav.Set(ViewLogs)
		}
	} else {
		m.logger.Add(operationError())
	}
	return m
}

func (m AppModel) waitCompletion(ctx context.Context) {
	go func() {
		<-ctx.Done()
		m.state.Set(Stale)
	}()
}

func (m AppModel) resizeViews() {
	m.filesView.Resize(m.width/2, m.height)
	m.logsView.Resize(m.width/2, m.height)
	m.settingsView.Resize(m.width, m.height)
	m.workersView.Resize(m.width, m.height)
}
