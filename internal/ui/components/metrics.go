package components

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
)

// tickInterval is the metrics refresh cadence when the panel is visible. At
// 100ms the throughput counter and progress indicators feel live without
// producing noticeable flicker on the host terminal.
const tickInterval = 100 * time.Millisecond

var (
	metricsTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	metricsLabelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Width(20)
	metricsValueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	metricsValueDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	metricsValueOK      = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
	metricsValueError   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	metricsElapsedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)

// MetricsPanel renders the live processor metrics used to live next to the
// execution log viewport. It is a value-type tea.Model so the parent
// LogsView can store it (as a pointer) and forward messages to it
// uniformly with every other component.
type MetricsPanel struct {
	proc     ports.ProcessorController
	Visible  bool
	last     ports.ProcessorMetrics
	tickEach time.Duration
}

// Compile-time guard: MetricsPanel must satisfy tea.Model with a value
// receiver. Phase 3.5 converts the historical pointer-receiver Update
// surface to value receivers; this assertion fails at build time if
// the conversion regresses.
var _ tea.Model = MetricsPanel{}

// NewMetricsPanel wires a metrics panel to the supplied processor controller.
// The panel starts hidden; callers flip visibility through
// Update(msgs.MetricsVisibilityMsg{Visible: ...}).
func NewMetricsPanel(proc ports.ProcessorController) MetricsPanel {
	return MetricsPanel{
		proc:     proc,
		Visible:  false,
		tickEach: tickInterval,
	}
}

// SetVisible is a convenience for the constructor and the panel's own
// internal Update path. External callers should prefer the message path
// (Update(msgs.MetricsVisibilityMsg{...})) so the change flows through
// the same code path as everything else.
func (p MetricsPanel) SetVisible(visible bool) MetricsPanel {
	p.Visible = visible
	return p
}

// Init returns nil per R-6.
func (p MetricsPanel) Init() tea.Cmd { return nil }

// Update handles a single tea message.
//
//   - msgs.MetricsVisibilityMsg flips the Visible flag and returns a
//     tick cmd when Visible is true (or nil when false).
//   - msgs.MetricsTickMsg refreshes the cached metrics snapshot and
//     reschedules the next tick when Visible is true.
//
// All other messages pass through with no effect and no command.
func (p MetricsPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case msgs.MetricsVisibilityMsg:
		p.Visible = msg.Visible
		if msg.Visible {
			return p, p.tickCmd()
		}
		return p, nil

	case msgs.MetricsTickMsg:
		if !p.Visible {
			return p, nil
		}
		p.last = p.proc.GetMetrics()
		_ = msg
		return p, p.tickCmd()
	}
	return p, nil
}

// View renders the eight metric rows in a fixed order. The output is the
// same regardless of visibility so the parent view can pre-render without
// flickering on activation.
func (p MetricsPanel) View() tea.View {
	return tea.NewView(p.view())
}

// view returns the raw string content. Extracted so the same logic
// can be used by tests that pre-Phase 3.5 expected a string and by
// the View() method that wraps it in a tea.View.
func (p MetricsPanel) view() string {
	m := p.last

	var status string
	if m.IsProcessing {
		status = metricsValueOK.Render("🟢 Processing")
	} else {
		status = metricsValueDim.Render("⚪ Idle")
	}

	errVal := metricsValueStyle.Render("0")
	if m.ErrorRequests > 0 {
		errVal = metricsValueError.Render(strconv.FormatUint(m.ErrorRequests, 10))
	}

	rows := []string{
		metricsTitleStyle.Render("📊 Real-Time Metrics"),
		"",
		metricsLabelStyle.Render("Status:") + " " + status,
		metricsLabelStyle.Render("Total Requests:") + " " + metricsValueStyle.Render(strconv.FormatUint(m.TotalRequests, 10)),
		metricsLabelStyle.Render("✓ Success:") + " " + metricsValueOK.Render(strconv.FormatUint(m.SuccessRequests, 10)),
		metricsLabelStyle.Render("✗ Errors:") + " " + errVal,
		metricsLabelStyle.Render("Lines Processed:") + " " + metricsValueStyle.Render(strconv.FormatUint(m.LinesProcessed, 10)),
		metricsLabelStyle.Render("Throughput:") + " " + metricsValueStyle.Render(fmt.Sprintf("%.2f req/s", m.RequestsPerSec)),
		metricsLabelStyle.Render("Active Workers:") + " " + metricsValueStyle.Render(strconv.Itoa(m.ActiveWorkers)),
	}

	if m.IsProcessing && !m.StartTime.IsZero() {
		rows = append(rows, metricsLabelStyle.Render("Elapsed Time:")+" "+metricsElapsedStyle.Render(formatDuration(time.Since(m.StartTime))))
	}

	var b strings.Builder
	for i, r := range rows {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(r)
	}
	return b.String()
}

// tickCmd schedules the next MetricsTickMsg after the configured interval.
func (p MetricsPanel) tickCmd() tea.Cmd {
	return tea.Tick(p.tickEach, func(t time.Time) tea.Msg {
		return msgs.MetricsTickMsg(t)
	})
}

// formatDuration formats an elapsed duration in a compact, human-readable way
// matching the original Workers view rendering (ms / s / m s).
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}
