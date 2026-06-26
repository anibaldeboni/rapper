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
// execution log viewport. It owns its own tick scheduling so callers only
// have to flip SetVisible when the host view becomes active or inactive.
type MetricsPanel struct {
	proc     ports.ProcessorController
	Visible  bool
	last     ports.ProcessorMetrics
	tickEach time.Duration
}

// NewMetricsPanel wires a metrics panel to the supplied processor controller.
// The panel starts hidden; call SetVisible(true) to start ticking.
func NewMetricsPanel(proc ports.ProcessorController) *MetricsPanel {
	return &MetricsPanel{
		proc:     proc,
		Visible:  false,
		tickEach: tickInterval,
	}
}

// SetVisible starts or stops the internal tick chain. The function is
// idempotent; calling it with the current visibility is a no-op.
func (p *MetricsPanel) SetVisible(visible bool) {
	p.Visible = visible
}

// Update handles a single tea message. MetricsTickMsg refreshes the cached
// snapshot and reschedules the next tick when the panel is visible. All other
// messages pass through with no effect and no command.
func (p *MetricsPanel) Update(msg tea.Msg) (*MetricsPanel, tea.Cmd) {
	tick, ok := msg.(msgs.MetricsTickMsg)
	if !ok {
		return p, nil
	}
	if !p.Visible {
		return p, nil
	}
	p.last = p.proc.GetMetrics()
	_ = tick
	return p, p.tickCmd()
}

// View renders the eight metric rows in a fixed order. The output is the
// same regardless of visibility so the parent view can pre-render without
// flickering on activation.
func (p *MetricsPanel) View() string {
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
func (p *MetricsPanel) tickCmd() tea.Cmd {
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
