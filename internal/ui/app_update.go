package ui

import (
	"context"
	"time"

	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/anibaldeboni/rapper/internal/ui/views"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

// ConfigSavedMsg is sent when configuration is successfully saved
type ConfigSavedMsg struct{}

// ConfigSaveErrorMsg is sent when configuration save fails
type ConfigSaveErrorMsg struct {
	Err error
}

// ProfileSwitchedMsg is sent when profile is successfully switched
type ProfileSwitchedMsg struct {
	ProfileName string
}

// ProfileSwitchErrorMsg is sent when profile switch fails
type ProfileSwitchErrorMsg struct {
	Err error
}

// ProcessingStartedMsg is sent when file processing begins
type ProcessingStartedMsg struct {
	FilePath string
}

// ProcessingStoppedMsg is sent when processing completes or is cancelled
type ProcessingStoppedMsg struct {
	Success bool
	Err     error
}

// ProcessingProgressMsg is sent periodically during processing with metrics
type ProcessingProgressMsg struct {
	Metrics processor.Metrics
}

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

	case tickMsg:
		// Update logs content
		m.logsView.UpdateLogs()

		// Update spinner
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinCmd, tickCmd())

		// Update toast manager (remove expired toasts)
		m.toastMgr.Update()

		// Send progress message if processing
		m.cancelMu.RLock()
		hasCancel := m.cancel != nil
		m.cancelMu.RUnlock()

		if hasCancel {
			metrics := m.processor.GetMetrics()
			progressCmd := func() tea.Msg {
				return ProcessingProgressMsg{Metrics: metrics}
			}
			logsCmd := m.logsView.Update(ProcessingProgressMsg{Metrics: metrics})
			cmds = append(cmds, progressCmd, logsCmd)
		}

		// Forward tick to WorkersView for metrics updates
		if m.nav.Current() == ViewWorkers {
			// Convert tickMsg to views.TickMsg
			workersCmd := m.workersView.Update(views.TickMsg(msg))
			cmds = append(cmds, workersCmd)
		}

	case ProcessingStartedMsg:
		m.logsView.SetProcessing(true)
		return m, nil

	case ProcessingStoppedMsg:
		// Clear cancel function
		m.cancelMu.Lock()
		m.cancel = nil
		m.cancelMu.Unlock()

		m.logsView.SetProcessing(false)
		if msg.Err != nil {
			m.toastMgr.Error("Processing failed: " + msg.Err.Error())
		} else if msg.Success {
			m.toastMgr.Success("Processing completed successfully")
		}
		return m, nil

	case ProcessingProgressMsg:
		// Progress is already handled in tickMsg
		return m, nil

	case ConfigSavedMsg:
		m.toastMgr.Success("Configuration saved successfully")
		return m, nil

	case ConfigSaveErrorMsg:
		m.toastMgr.Error("Failed to save configuration: " + msg.Err.Error())
		return m, nil

	case ProfileSwitchedMsg:
		m.toastMgr.Success("Switched to profile: " + msg.ProfileName)
		return m, nil

	case ProfileSwitchErrorMsg:
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
				return ProcessingStartedMsg{FilePath: filePath}
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
		return ProcessingStoppedMsg{
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
