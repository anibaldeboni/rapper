package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	workersTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1).Bold(true)
	workersAppStyle   = lipgloss.NewStyle().Margin(1, 2)
	metricLabelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Width(20)
	metricValueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	sliderTrackStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sliderFillStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	sliderHandleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("205")).Bold(true)
)

// TickMsg is sent periodically to update metrics
type TickMsg time.Time

// WorkersView displays and controls worker pool
type WorkersView struct {
	proc         processor.Processor
	width        int
	height       int
	workerCount  int
	maxWorkers   int
	lastMetrics  processor.Metrics
	tickInterval time.Duration
}

// NewWorkersView creates a new WorkersView
func NewWorkersView(proc processor.Processor) *WorkersView {
	return &WorkersView{
		proc:         proc,
		workerCount:  proc.GetWorkerCount(),
		maxWorkers:   processor.MaxWorkers,
		tickInterval: 500 * time.Millisecond,
	}
}

// Update handles messages for the workers view
func (v *WorkersView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if v.workerCount > 1 {
				v.workerCount--
				v.proc.SetWorkers(v.workerCount)
			}
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if v.workerCount < v.maxWorkers {
				v.workerCount++
				v.proc.SetWorkers(v.workerCount)
			}
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("-"))):
			if v.workerCount > 1 {
				v.workerCount--
				v.proc.SetWorkers(v.workerCount)
			}
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("+", "="))):
			if v.workerCount < v.maxWorkers {
				v.workerCount++
				v.proc.SetWorkers(v.workerCount)
			}
			return nil
		}

	case TickMsg:
		// Update metrics
		v.lastMetrics = v.proc.GetMetrics()
		// Schedule next tick
		return v.tick()
	}

	return nil
}

// tick returns a command that sends a TickMsg after the configured interval
func (v *WorkersView) tick() tea.Cmd {
	return tea.Tick(v.tickInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Init starts the periodic tick for metrics updates
func (v *WorkersView) Init() tea.Cmd {
	return v.tick()
}

// Resize updates the view dimensions
func (v *WorkersView) Resize(width, height int) {
	v.width = width
	v.height = height
}

// View renders the workers view
func (v *WorkersView) View() string {
	var b strings.Builder

	// Title
	title := workersTitleStyle.Render("👷 Workers Control")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Worker count slider
	b.WriteString(metricLabelStyle.Render("Worker Count:"))
	b.WriteString(" ")
	b.WriteString(v.renderSlider())
	b.WriteString(fmt.Sprintf(" %d / %d", v.workerCount, v.maxWorkers))
	b.WriteString("\n\n")

	// Metrics section
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render("📊 Real-Time Metrics"))
	b.WriteString("\n\n")

	// Processing status
	statusLabel := metricLabelStyle.Render("Status:")
	var statusValue string
	if v.lastMetrics.IsProcessing {
		statusValue = metricValueStyle.Render("🟢 Processing")
	} else {
		statusValue = metricValueStyle.Render("⚪ Idle")
	}
	b.WriteString(statusLabel)
	b.WriteString(" ")
	b.WriteString(statusValue)
	b.WriteString("\n")

	// Total requests
	totalLabel := metricLabelStyle.Render("Total Requests:")
	totalValue := metricValueStyle.Render(fmt.Sprintf("%d", v.lastMetrics.TotalRequests))
	b.WriteString(totalLabel)
	b.WriteString(" ")
	b.WriteString(totalValue)
	b.WriteString("\n")

	// Success requests
	successLabel := metricLabelStyle.Render("✓ Success:")
	successValue := metricValueStyle.Copy().Foreground(lipgloss.Color("40")).Render(fmt.Sprintf("%d", v.lastMetrics.SuccessRequests))
	b.WriteString(successLabel)
	b.WriteString(" ")
	b.WriteString(successValue)
	b.WriteString("\n")

	// Error requests
	errorLabel := metricLabelStyle.Render("✗ Errors:")
	var errorValue string
	if v.lastMetrics.ErrorRequests > 0 {
		errorValue = metricValueStyle.Copy().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("%d", v.lastMetrics.ErrorRequests))
	} else {
		errorValue = metricValueStyle.Render("0")
	}
	b.WriteString(errorLabel)
	b.WriteString(" ")
	b.WriteString(errorValue)
	b.WriteString("\n")

	// Lines processed
	linesLabel := metricLabelStyle.Render("Lines Processed:")
	linesValue := metricValueStyle.Render(fmt.Sprintf("%d", v.lastMetrics.LinesProcessed))
	b.WriteString(linesLabel)
	b.WriteString(" ")
	b.WriteString(linesValue)
	b.WriteString("\n")

	// Requests per second
	reqPerSecLabel := metricLabelStyle.Render("Throughput:")
	reqPerSecValue := metricValueStyle.Render(fmt.Sprintf("%.2f req/s", v.lastMetrics.RequestsPerSec))
	b.WriteString(reqPerSecLabel)
	b.WriteString(" ")
	b.WriteString(reqPerSecValue)
	b.WriteString("\n")

	// Active workers
	activeWorkersLabel := metricLabelStyle.Render("Active Workers:")
	activeWorkersValue := metricValueStyle.Render(fmt.Sprintf("%d", v.lastMetrics.ActiveWorkers))
	b.WriteString(activeWorkersLabel)
	b.WriteString(" ")
	b.WriteString(activeWorkersValue)
	b.WriteString("\n\n")

	// Elapsed time (if processing)
	if v.lastMetrics.IsProcessing && !v.lastMetrics.StartTime.IsZero() {
		elapsed := time.Since(v.lastMetrics.StartTime)
		elapsedLabel := metricLabelStyle.Render("Elapsed Time:")
		elapsedValue := metricValueStyle.Render(formatDuration(elapsed))
		b.WriteString(elapsedLabel)
		b.WriteString(" ")
		b.WriteString(elapsedValue)
		b.WriteString("\n\n")
	}

	return workersAppStyle.Render(b.String())
}

// renderSlider renders a visual slider for worker count
func (v *WorkersView) renderSlider() string {
	sliderWidth := 30
	if v.maxWorkers <= 0 {
		return sliderTrackStyle.Render(strings.Repeat("─", sliderWidth))
	}

	// Calculate position (0 to sliderWidth-1)
	pos := int(float64(v.workerCount-1) / float64(v.maxWorkers-1) * float64(sliderWidth-1))

	var slider strings.Builder

	for i := 0; i < sliderWidth; i++ {
		if i == pos {
			slider.WriteString(sliderHandleStyle.Render("●"))
		} else if i < pos {
			slider.WriteString(sliderFillStyle.Render("━"))
		} else {
			slider.WriteString(sliderTrackStyle.Render("─"))
		}
	}

	return slider.String()
}

// formatDuration formats a duration in a human-readable way
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
