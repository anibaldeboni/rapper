package ui

import (
	"context"
	"time"

	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/views"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return msgs.TickMsg(t)
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
				m.cancelMu.RLock()
				hasCancel := m.cancel != nil
				m.cancelMu.RUnlock()

				if hasCancel {
					m.cancelMu.Lock()
					if m.cancel != nil {
						m.cancel()
					}
					m.cancelMu.Unlock()
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

	case msgs.TickMsg:
		// Update spinner
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinCmd, tickCmd())

		// Update toast manager (remove expired toasts)
		m.toastMgr.Update()

		// Send progress message to LogsView if processing
		m.cancelMu.RLock()
		hasCancel := m.cancel != nil
		m.cancelMu.RUnlock()

		if hasCancel {
			// Forward progress message to LogsView to update content
			logsCmd := m.logsView.Update(msgs.ProcessingProgressMsg{Metrics: m.processor.GetMetrics()})
			cmds = append(cmds, logsCmd)
		}

		// Forward tick to WorkersView for metrics updates
		if m.nav.Current() == ViewWorkers {
			// Convert tickMsg to views.TickMsg
			workersCmd := m.workersView.Update(views.TickMsg(msg))
			cmds = append(cmds, workersCmd)
		}

	case msgs.ProcessingStartedMsg:
		cmd := m.logsView.Update(msg)
		return m, cmd

	case msgs.ProcessingStoppedMsg:
		// Clear cancel function
		m.cancelMu.Lock()
		m.cancel = nil
		m.cancelMu.Unlock()

		// Forward message to LogsView
		logsCmd := m.logsView.Update(msg)

		if msg.Err != nil {
			m.toastMgr.Error("Processing failed: " + msg.Err.Error())
		} else if msg.Success {
			m.toastMgr.Success("Processing completed successfully")
		}
		return m, logsCmd

	case msgs.ConfigSavedMsg:
		m.toastMgr.Success("Configuration saved successfully")
		return m, nil

	case msgs.ConfigSaveErrorMsg:
		m.toastMgr.Error("Failed to save configuration: " + msg.Err.Error())
		return m, nil

	case msgs.ProfileSwitchedMsg:
		m.toastMgr.Success("Switched to profile: " + msg.ProfileName)
		return m, nil

	case msgs.ProfileSwitchErrorMsg:
		m.toastMgr.Error("Failed to switch profile: " + msg.Err.Error())
		return m, nil
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
				newModel, cmd := m.selectFile(opt.Value)
				return newModel, cmd
			}
		}

		cmd = m.filesView.Update(msg)

	case ViewLogs:
		cmd = m.logsView.Update(msg)

	case ViewSettings:
		cmd = m.settingsView.Update(msg)

	case ViewWorkers:
		cmd = m.workersView.Update(msg)
	}

	return m, cmd
}

func (m AppModel) selectFile(filePath string) (AppModel, tea.Cmd) {
	m.cancelMu.RLock()
	hasCancel := m.cancel != nil
	m.cancelMu.RUnlock()

	if hasCancel {
		// Already processing - show error
		m.logger.Add(operationError())
		m.nav.Set(ViewLogs)
		return m, nil
	}

	// Start processing
	ctx, cancel := m.processor.Do(context.Background(), filePath)
	if ctx != nil {
		m.cancelMu.Lock()
		m.cancel = cancel
		m.cancelMu.Unlock()

		// Switch to logs view when processing starts
		m.nav.Set(ViewLogs)

		// Return batch of commands: emit ProcessingStartedMsg and wait for completion
		return m, tea.Batch(
			func() tea.Msg {
				return msgs.ProcessingStartedMsg{FilePath: filePath}
			},
			m.waitCompletion(ctx),
		)
	}

	m.nav.Set(ViewLogs)
	return m, nil
}

func (m *AppModel) waitCompletion(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		<-ctx.Done()
		// Check if it was cancelled or completed successfully
		err := ctx.Err()
		success := err == nil || err == context.Canceled
		return msgs.ProcessingStoppedMsg{
			Success: success,
			Err:     nil,
		}
	}
}

func (m AppModel) resizeViews() {
	// Calculate available height for content
	// Subtract:
	// - 1 line for header (help bar)
	// - 1 line for status bar
	// - 2 lines for AppStyle margins (top + bottom)
	// Ensure minimum height
	availableHeight := max(m.height-4, 10)

	m.filesView.Resize(m.width/2, availableHeight)
	m.logsView.Resize(m.width/2, availableHeight)
	m.settingsView.Resize(m.width, availableHeight)
	m.workersView.Resize(m.width, availableHeight)
}
