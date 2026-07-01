package views

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/ui/components"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
)

var (
	logTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).MarginBottom(1).Padding(0, 1).Bold(true)
)

// metricsMinWidth is the floor for the metrics column on the right side of
// the Logs view. The width is recomputed against the full available width
// in Resize, capped to ~30% of the available area so a wide terminal does
// not give the metrics an absurd amount of space. The minimum is large
// enough to hold the 20-char label column plus a separator plus a
// multi-digit value, so a narrow terminal still shows complete rows.
const metricsMinWidth = 24

// metricsDefaultWidth is the default width of the metrics column on the
// right side of the Logs view when there is plenty of room. Sized to fit
// the widest row (label + separator + value, e.g. throughput with five
// digits + " req/s") without lipgloss clipping the right edge.
const metricsDefaultWidth = 36

// logsMarginLeft is the left margin applied to the rendered logs body
// (see logs.go View() — lipgloss.NewStyle().MarginLeft(2)). It is a
// view-local concern, NOT a chrome dimension: the value is consumed
// inside the LogsView's own render path and must be subtracted from the
// available width when partitioning the viewport (left) from the metrics
// panel (right). Exported as a package constant so the regression test
// in logs_test.go can assert the partition math is exact.
const logsMarginLeft = 2

// LogsView displays execution logs alongside the live metrics panel.
// It is a value-type tea.Model so AppModel can store it behind a
// uniform `map[View]viewModel` and broadcast messages to every view
// without re-dispatching on the concrete type.
type LogsView struct {
	viewport   viewport.Model
	logger     ports.LogProvider
	metrics    components.MetricsPanel
	title      string
	width      int
	height     int
	rightCol   int
	autoScroll bool
}

// Compile-time guard: LogsView must satisfy tea.Model with a value
// receiver. Phase 2 converts the historical pointer-receiver surface
// to value receivers; this assertion fails at build time if the
// conversion regresses.
var _ tea.Model = LogsView{}

// NewLogsView creates a LogsView. The proc parameter is required so the
// in-view metrics panel can refresh from the processor state.
//
// The view is a value (not a pointer) so AppModel can store it
// behind the unified viewModel type. Callers must capture the value
// returned by Update to preserve state.
func NewLogsView(logger ports.LogProvider, proc ports.ProcessorController) LogsView {
	vp := viewport.New(viewport.WithWidth(0), viewport.WithHeight(0))

	v := LogsView{
		viewport:   vp,
		logger:     logger,
		metrics:    components.NewMetricsPanel(proc),
		title:      "📝 Execution logs",
		rightCol:   metricsDefaultWidth,
		autoScroll: true,
	}
	// Load initial logs
	v = v.updateLogs()
	return v
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
//     the log viewport (left) and the metrics panel (right).
//   - msgs.MetricsVisibilityMsg: start (true) or stop (false) the
//     metrics tick chain. The chain is self-sustaining once started:
//     MetricsPanel.Update(MetricsTickMsg) reschedules itself.
func (v LogsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case msgs.ViewportSizeMsg:
		v = v.applyViewportSize(msg.Width, msg.Height)

	case msgs.MetricsVisibilityMsg:
		// Start or stop the metrics tick chain. The panel's Visible
		// flag is the source of truth; LogsView owns the scheduling
		// (the tea.Tick cmd) per R-7 / D-5. The chain becomes
		// self-sustaining because the next MetricsTickMsg (delivered
		// by the tick cmd we return) is forwarded to the panel which
		// reschedules via LogsView at the top of Update.
		v.metrics.Visible = msg.Visible
		if msg.Visible {
			return v, metricsTickCmd()
		}
		return v, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, kbind.GotoBottom):
			v.viewport.GotoBottom()
			v.autoScroll = true
		case key.Matches(msg, kbind.Up):
			v.viewport.ScrollUp(1)
			v.autoScroll = false
		case key.Matches(msg, kbind.Down):
			v.viewport.ScrollDown(1)
			v.autoScroll = false
		case key.Matches(msg, kbind.PageUp):
			v.viewport.PageUp()
			v.autoScroll = false
		case key.Matches(msg, kbind.PageDown):
			v.viewport.PageDown()
			v.autoScroll = false
		case key.Matches(msg, kbind.GotoTop):
			v.viewport.GotoTop()
			v.autoScroll = false
		case key.Matches(msg, kbind.Right):
			v.viewport.ScrollRight(1)
			v.autoScroll = false
		case key.Matches(msg, kbind.Left):
			v.viewport.ScrollLeft(1)
			v.autoScroll = false
		}

	case msgs.MetricsTickMsg:
		// Forward to the embedded panel so it can refresh the cached
		// metrics snapshot and reschedule its own tick cmd.
		next, mcmd := v.metrics.Update(msg)
		v.metrics = next.(components.MetricsPanel)
		cmd = mcmd
		v = v.updateLogs()
		return v, cmd

	default:
		switch msg.(type) {
		case msgs.ProcessingStartedMsg:
			v.autoScroll = true
		case msgs.ProcessingStoppedMsg:
		case msgs.ProcessingProgressMsg:
		}
	}
	v = v.updateLogs()
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

// applyViewportSize re-partitions the available width between the log
// viewport (left) and the metrics panel (right). The partition math
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

	v.viewport.SetWidth(left)
	v.viewport.SetHeight(height - 3)
	v.rightCol = right
	return v
}

// View renders the logs view as a tea.View whose Content holds the
// joined (viewport + metrics panel) body.
func (v LogsView) View() tea.View {
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		v.viewport.View(),
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

// updateLogs updates the viewport content with latest logs and
// auto-scrolls to the bottom when autoScroll is true. Operates on a
// value-receiver copy and returns the modified LogsView so callers
// preserve the new viewport state. The returned value MUST be
// captured by every call site — the value receiver means the
// original struct is never mutated.
func (v LogsView) updateLogs() LogsView {
	content := strings.Join(v.logger.Get(), "\n")
	v.viewport.SetContent(content)
	if v.autoScroll {
		v.viewport.GotoBottom()
	}
	return v
}

// MetricsVisible returns true if the embedded metrics panel is
// currently ticking. Used by AppModel tests to assert the visibility
// state after a nav switch.
func (v LogsView) MetricsVisible() bool { return v.metrics.Visible }

// ViewportWidth returns the assigned viewport width. Diagnostic
// accessor for the flow tests.
func (v LogsView) ViewportWidth() int { return v.viewport.Width() }

// ViewportHeight returns the assigned viewport height. Diagnostic
// accessor for the flow tests.
func (v LogsView) ViewportHeight() int { return v.viewport.Height() }

// ViewportContent returns the current viewport content. Diagnostic
// accessor for the flow tests.
func (v LogsView) ViewportContent() string { return v.viewport.GetContent() }

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
