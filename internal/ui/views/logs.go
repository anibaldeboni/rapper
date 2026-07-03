package views

import (
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/ui/components"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
)

var (
	logTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).MarginBottom(1).Padding(0, 1).Bold(true)
)

const (
	metricsMinWidth     = 24
	metricsDefaultWidth = 36
	logsMarginLeft      = 2
)

// LogsView displays execution logs alongside the live metrics panel.
// It is a value-type tea.Model so AppModel can store it behind a
// uniform `map[View]viewModel` and broadcast messages to every view
// without re-dispatching on the concrete type.
//
// The log list itself is a generic DetailedList[logs.LogMessage]
// driven by a LogMessageRenderer — the view does not know how to
// render a row, only how to lay out the column that contains the
// list.
type LogsView struct {
	list     components.DetailedList[logs.LogMessage]
	logger   ports.LogProvider
	metrics  components.MetricsPanel
	title    string
	width    int
	height   int
	rightCol int
}

// Compile-time guard: LogsView must satisfy tea.Model with a value
// receiver.
var _ tea.Model = LogsView{}

// NewLogsView creates a LogsView. The proc parameter is required so the
// in-view metrics panel can refresh from the processor state.
//
// The view is a value (not a pointer) so AppModel can store it
// behind the unified viewModel type. Callers must capture the value
// returned by Update to preserve state.
func NewLogsView(logger ports.LogProvider, proc ports.ProcessorController) LogsView {
	v := LogsView{
		list:     components.NewDetailedList(components.LogMessageRenderer{}),
		logger:   logger,
		metrics:  components.NewMetricsPanel(proc),
		title:    "📝 Execution logs",
		rightCol: metricsDefaultWidth,
	}
	return v.refreshLogs()
}

// Init returns nil. The metrics tick chain is started via
// Update(MetricsVisibilityMsg{Visible: true}), not Init, so the chain
// only ticks when the Logs view is the active view.
func (v LogsView) Init() tea.Cmd { return nil }

// Update handles messages for the logs view. The view returns a
// value-receiver copy (LogsView) plus an optional command. Callers
// MUST capture the returned LogsView to preserve state — the value
// receiver means the original struct is never mutated.
//
// Recognised messages:
//   - msgs.ViewportSizeMsg: re-partition the available width between
//     the log list (left) and the metrics panel (right).
//   - msgs.MetricsVisibilityMsg: start (true) or stop (false) the
//     metrics tick chain.
//   - msgs.ProcessingStartedMsg: clear the list (each run starts
//     fresh) and re-enable autoScroll.
//   - msgs.MetricsTickMsg: refresh the list with any new log
//     messages; forward to the embedded metrics panel.
//   - tea.KeyPressMsg: forward navigation keys to the list.
//
// Horizontal navigation (Left/Right) is intentionally not handled —
// DetailedList is a vertical list.
func (v LogsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case msgs.ViewportSizeMsg:
		v = v.applyViewportSize(msg.Width, msg.Height)

	case msgs.MetricsVisibilityMsg:
		v.metrics.Visible = msg.Visible
		if msg.Visible {
			return v, metricsTickCmd()
		}
		return v, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, kbind.GotoBottom),
			key.Matches(msg, kbind.Up),
			key.Matches(msg, kbind.Down),
			key.Matches(msg, kbind.PageUp),
			key.Matches(msg, kbind.PageDown),
			key.Matches(msg, kbind.GotoTop),
			key.Matches(msg, kbind.Select):
			next, kcmd := v.list.Update(msg)
			v.list = next.(components.DetailedList[logs.LogMessage])
			return v, kcmd
		}

	case msgs.MetricsTickMsg:
		// Forward to the embedded panel so it can refresh the cached
		// metrics snapshot and reschedule its own tick cmd.
		next, mcmd := v.metrics.Update(msg)
		v.metrics = next.(components.MetricsPanel)
		cmd = mcmd
		v = v.refreshLogs()
		return v, cmd

	case msgs.ProcessingStartedMsg:
		// Clear both the in-memory buffer and the embedded list so
		// each run starts from a clean slate. Clear() must run
		// before Reset() — otherwise the next MetricsTickMsg would
		// see the old messages in the buffer and re-append them.
		v.logger.Clear()
		v.list = v.list.Reset()
		return v, nil
	}

	return v, cmd
}

// applyViewportSize re-partitions the available width between the log
// list (left) and the metrics panel (right). The partition math
// matches the historical Resize behaviour: the right column is the
// default (36) capped to ~30% of the available width with a floor of
// 24, then the left column takes whatever is left minus the
// view-local MarginLeft (2).
func (v LogsView) applyViewportSize(width, height int) LogsView {
	v.width = width
	v.height = height

	right := metricsDefaultWidth
	if width > 0 {
		maxRight := width * 30 / 100
		if maxRight < right {
			right = maxRight
		}
		if right < metricsMinWidth {
			right = metricsMinWidth
		}
	}
	left := max(width-right-logsMarginLeft, 0)

	v.list = v.list.SetSize(left, max(height-3, 0))
	v.rightCol = right
	return v
}

// View renders the logs view as a tea.View whose Content holds the
// joined (list + metrics panel) body.
func (v LogsView) View() tea.View {
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		v.list.View().Content,
		v.metrics.View().Content,
	)

	return tea.NewView(lipgloss.NewStyle().
		MarginLeft(2).
		MarginTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				logTitleStyle.Render(v.title),
				body,
			),
		))
}

// refreshLogs reads the current log buffer and appends any new
// messages to the embedded DetailedList. The List's Append honours
// autoScroll — when the user is at the tail, the cursor follows
// the new entries; when the user has scrolled away, the cursor
// stays put and the new entries queue up below.
//
// Operates on a value-receiver copy and returns the modified
// LogsView so callers preserve the new list state. The returned
// value MUST be captured by every call site — the value receiver
// means the original struct is never mutated.
func (v LogsView) refreshLogs() LogsView {
	all := v.logger.Get()
	if len(all) < v.list.Len() {
		// The buffer shrank (e.g. Clear() was called). Reset the list
		// so we don't keep stale expanded rows and the cursor
		// repositions correctly.
		v.list = v.list.Reset()
	}
	if len(all) > v.list.Len() {
		newOnes := all[v.list.Len():]
		v.list = v.list.Append(newOnes)
	}
	return v
}

// MetricsVisible returns true if the embedded metrics panel is
// currently ticking. Used by AppModel tests to assert the visibility
// state after a nav switch.
func (v LogsView) MetricsVisible() bool { return v.metrics.Visible }

// RightCol returns the assigned right-pane width. Diagnostic
// accessor for the flow tests.
func (v LogsView) RightCol() int { return v.rightCol }

// listWidthHeight returns the assigned left-pane (list) width and
// height. Diagnostic accessor for the partition-invariant tests.
func (v LogsView) listWidthHeight() (int, int) {
	return v.list.Width(), v.list.Height()
}

// metricsTickInterval is the metrics refresh cadence. Matches the
// MetricsPanel.tickInterval constant so the two stay in sync; the
// panel's own Update uses its own copy for the reschedule path.
const metricsTickInterval = 100 * time.Millisecond

// metricsTickCmd schedules the next MetricsTickMsg after the metrics
// interval. LogsView owns the scheduling; the panel just receives the
// resulting tick and reschedules by replying through Update.
func metricsTickCmd() tea.Cmd {
	return tea.Tick(metricsTickInterval, func(t time.Time) tea.Msg {
		return msgs.MetricsTickMsg(t)
	})
}
