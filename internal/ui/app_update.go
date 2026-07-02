package ui

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
)

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return msgs.TickMsg(t)
	})
}

// emit returns a tea.Cmd that yields m when run. Use in tea.Batch
// compositions to inject messages into the message stream without
// allocating a dedicated helper per message type. Mirrors the
// emitItemSelected pattern in views/files.go so the two read
// consistently.
func emit(m tea.Msg) tea.Cmd {
	return func() tea.Msg { return m }
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
			m.currentView = ViewFiles
			return m, m.routeToAllViews(msgs.MetricsVisibilityMsg{Visible: false})

		case key.Matches(msg, kbind.ViewLogs):
			m.currentView = ViewLogs
			return m, m.routeToAllViews(msgs.MetricsVisibilityMsg{Visible: true})

		case key.Matches(msg, kbind.ViewSettings):
			m.currentView = ViewSettings
			return m, m.routeToAllViews(msgs.MetricsVisibilityMsg{Visible: false})

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

		// Delegate to current view only — not the whole map.
		return m.updateCurrentView(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Synthesize a chrome-adjusted per-view ViewportSizeMsg and
		// route it to every view. The historical imperative dispatch
		// is gone; this is the message-based equivalent.
		aw := m.chrome.AvailableWidth(m.width)
		ah := max(m.chrome.AvailableHeight(m.height), 10)
		return m, m.routeToAllViews(msgs.ViewportSizeMsg{Width: aw, Height: ah})

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
			next, logsCmd := m.views[ViewLogs].Update(msgs.ProcessingProgressMsg{
				TotalRequests:   metrics.TotalRequests,
				SuccessRequests: metrics.SuccessRequests,
				ErrorRequests:   metrics.ErrorRequests,
				LinesProcessed:  metrics.LinesProcessed,
				ActiveWorkers:   metrics.ActiveWorkers,
				RequestsPerSec:  metrics.RequestsPerSec,
				StartTime:       metrics.StartTime,
				IsProcessing:    metrics.IsProcessing,
			})
			m.views[ViewLogs] = next
			cmds = append(cmds, logsCmd)
		}

		// MetricsTickMsg forwarding was removed: the dedicated
		// `case msgs.MetricsTickMsg` below now broadcasts the tick
		// to every view, so duplicating it here would tick the
		// LogsView's panel twice per interval (2x refresh rate).

	case msgs.ProcessingStartedMsg:
		next, cmd := m.views[ViewLogs].Update(msg)
		m.views[ViewLogs] = next
		return m, cmd

	case msgs.MetricsTickMsg:
		// Broadcast MetricsTickMsg to every view. Only LogsView is
		// required to act on it; the others no-op and return their
		// model unchanged with a nil command. The chain self-sustains
		// because LogsView.Update(MetricsTickMsg) returns the next
		// metricsTickCmd via the embedded panel.
		return m, m.routeToAllViews(msg)

	case msgs.MetricsVisibilityMsg:
		// Broadcast MetricsVisibilityMsg to every view. Only LogsView
		// flips the embedded MetricsPanel.Visible flag and schedules
		// (or stops) the tick chain; the others no-op. This case is
		// what the selectFile batch relies on: the batched message
		// is delivered by the framework on the next Update tick and
		// must reach the LogsView for the chain to self-sustain.
		return m, m.routeToAllViews(msg)

	case msgs.ProcessingStoppedMsg:
		// Clear cancel function
		m.cancelMu.Lock()
		m.cancel = nil
		m.cancelMu.Unlock()

		// Forward message to LogsView
		next, logsCmd := m.views[ViewLogs].Update(msg)
		m.views[ViewLogs] = next

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

	case msgs.ItemSelectedMsg:
		// FilesView (Phase 4) emits ItemSelectedMsg on Select. The
		// AppModel routes it to selectFile which starts processing
		// and switches to the LogsView. This is the unidirectional
		// equivalent of the historical AppModel interception +
		// SelectedItem() query.
		return m.selectFile(msg.FilePath)
	}

	return m, tea.Batch(cmds...)
}

// routeToAllViews broadcasts a message to every view in the views map,
// captures the returned model for each, and batches the returned cmds.
// This is the only mechanism for cross-cutting messages (size, theme,
// visibility) — AppModel never calls a view-specific method.
func (m AppModel) routeToAllViews(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for k, v := range m.views {
		next, cmd := v.Update(msg)
		m.views[k] = next
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

// updateCurrentView routes a keypress to the active view only. The
// returned next is captured into the views map; the returned cmd is
// routed back through the top-level Update so ItemSelectedMsg and
// other messages produced by the view flow through the normal dispatch.
func (m AppModel) updateCurrentView(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	active := m.currentView
	next, cmd := m.views[active].Update(msg)
	m.views[active] = next
	return m, cmd
}

func (m AppModel) selectFile(filePath string) (tea.Model, tea.Cmd) {
	m.cancelMu.RLock()
	hasCancel := m.cancel != nil
	m.cancelMu.RUnlock()

	if hasCancel {
		// Already processing - show error
		m.logger.Add(operationError())
		m.currentView = ViewLogs
		return m, nil
	}

	// Start processing
	ctx, cancel := m.processor.Do(context.Background(), filePath)
	if ctx != nil {
		m.cancelMu.Lock()
		m.cancel = cancel
		m.cancelMu.Unlock()

		// Switch to logs view when processing starts
		m.currentView = ViewLogs

		// Return batch of commands: emit ProcessingStartedMsg,
		// start the metrics tick chain via MetricsVisibilityMsg,
		// and wait for completion. The visibility message is
		// routed to every view; only LogsView acts on it (it flips
		// the embedded MetricsPanel.Visible flag and schedules the
		// first metrics tick cmd).
		return m, tea.Batch(
			func() tea.Msg {
				return msgs.ProcessingStartedMsg{FilePath: filePath}
			},
			emit(msgs.MetricsVisibilityMsg{Visible: true}),
			m.waitCompletion(ctx),
		)
	}

	m.currentView = ViewLogs
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
