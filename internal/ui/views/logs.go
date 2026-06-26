package views

import (
	"strings"

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
// not give the metrics an absurd amount of space.
const metricsMinWidth = 18

// metricsDefaultWidth is the default width of the metrics column on the
// right side of the Logs view when there is plenty of room.
const metricsDefaultWidth = 28

// LogsView displays execution logs alongside the live metrics panel.
type LogsView struct {
	viewport   viewport.Model
	logger     ports.LogProvider
	metrics    *components.MetricsPanel
	title      string
	width      int
	height     int
	rightCol   int
	autoScroll bool
}

// NewLogsView creates a LogsView. The proc parameter is required so the
// in-view metrics panel can refresh from the processor state.
func NewLogsView(logger ports.LogProvider, proc ports.ProcessorController) *LogsView {
	vp := viewport.New(viewport.WithWidth(0), viewport.WithHeight(0))

	v := &LogsView{
		viewport:   vp,
		logger:     logger,
		metrics:    components.NewMetricsPanel(proc),
		title:      "📝 Execution logs",
		rightCol:   metricsDefaultWidth,
		autoScroll: true,
	}
	// Load initial logs
	v.updateLogs()
	return v
}

// SetMetricsVisible starts or stops the metrics tick chain. AppModel calls
// this with true when the Logs view becomes active and false when it leaves.
func (v *LogsView) SetMetricsVisible(visible bool) {
	v.metrics.SetVisible(visible)
}

// Update handles messages for the logs view. The metrics panel owns its own
// tick chain; we only forward the MetricsTickMsg to it.
func (v *LogsView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
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
		v.metrics, cmd = v.metrics.Update(msg)
		return cmd
	default:
		switch msg.(type) {
		case msgs.ProcessingStartedMsg:
			v.autoScroll = true
		case msgs.ProcessingStoppedMsg:
		case msgs.ProcessingProgressMsg:
		}
	}
	v.updateLogs()
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

// Resize updates the view dimensions and re-partitions the available width
// between the log viewport (left) and the metrics panel (right).
func (v *LogsView) Resize(width, height int) {
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
	left := max(width-right, 0)

	v.viewport.SetWidth(left)
	v.viewport.SetHeight(height - 3)
	v.rightCol = right
}

// View renders the logs view with the log viewport on the left and the
// metrics panel on the right. Both panels together cover the full content
// width assigned to the view.
func (v *LogsView) View() string {
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		v.viewport.View(),
		v.metrics.View(),
	)

	return lipgloss.NewStyle().
		MarginLeft(2).
		MarginTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				logTitleStyle.Render(v.title),
				body,
			),
		)
}

// updateLogs updates the viewport content with latest logs and auto-scrolls if processing
func (v *LogsView) updateLogs() {
	content := strings.Join(v.logger.Get(), "\n")
	v.viewport.SetContent(content)
	if v.autoScroll {
		v.viewport.GotoBottom()
	}
}
