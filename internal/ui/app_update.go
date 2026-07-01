package ui

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/views"
)

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return msgs.TickMsg(t)
	})
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Global navigation keys
		switch {
		case key.Matches(msg, kbind.Quit):
			return m, tea.Quit

		case key.Matches(msg, kbind.ViewFiles):
			m.nav.Set(ViewFiles)
			m.logsView.SetMetricsVisible(false)
			return m, nil

		case key.Matches(msg, kbind.ViewLogs):
			m.nav.Set(ViewLogs)
			m.logsView.SetMetricsVisible(true)
			return m, nil

		case key.Matches(msg, kbind.ViewSettings):
			m.nav.Set(ViewSettings)
			m.logsView.SetMetricsVisible(false)
			return m, nil

		case key.Matches(msg, kbind.CancelOperation):
			if m.cancel != nil {
				m.cancelMu.Lock()
				m.cancel()
				m.cancelMu.Unlock()
			} else {
				m.toastMgr.Warning("Batch processing isn't running")
			}

			return m, nil
		}

		// Delegate to current view
		return m.updateCurrentView(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.broadcastResize()
		return m, nil

	case tea.BackgroundColorMsg:
		if msg.IsDark() != m.isDark {
			m.applyTheme(msg.IsDark())
		}
		return m, nil

	case tea.KeyboardEnhancementsMsg:
		m.hasKeyEventTypes = msg.SupportsEventTypes()
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
			metrics := m.processor.GetMetrics()
			logsCmd := m.logsView.Update(msgs.ProcessingProgressMsg{
				TotalRequests:   metrics.TotalRequests,
				SuccessRequests: metrics.SuccessRequests,
				ErrorRequests:   metrics.ErrorRequests,
				LinesProcessed:  metrics.LinesProcessed,
				ActiveWorkers:   metrics.ActiveWorkers,
				RequestsPerSec:  metrics.RequestsPerSec,
				StartTime:       metrics.StartTime,
				IsProcessing:    metrics.IsProcessing,
			})
			cmds = append(cmds, logsCmd)
		}

		// Forward metrics tick to LogsView only when active. The metrics
		// panel owns its own tick chain and stops it via SetVisible(false)
		// when the user navigates away.
		if m.nav.Current() == ViewLogs {
			metricsCmd := m.logsView.Update(msgs.MetricsTickMsg(msg))
			cmds = append(cmds, metricsCmd)
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
			m.toastMgr.Success("Processing completed")
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

func (m AppModel) updateCurrentView(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.nav.Current() {
	case ViewFiles:
		// Handle file selection
		if key.Matches(msg, kbind.Select) {
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

func (m AppModel) broadcastResize() {
	availableHeight := max(m.chrome.AvailableHeight(m.height), 10)
	availableWidth := m.chrome.AvailableWidth(m.width)

	m.filesView.Resize(availableWidth, availableHeight)
	m.logsView.Resize(availableWidth, availableHeight)
	m.settingsView.Resize(availableWidth, availableHeight)
}
