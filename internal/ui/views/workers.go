package views

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
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
	proc         ports.ProcessorController
	width        int
	height       int
	workerCount  int
	maxWorkers   int
	lastMetrics  ports.ProcessorMetrics
	tickInterval time.Duration
	viewport     viewport.Model
}

// NewWorkersView creates a new WorkersView
func NewWorkersView(proc ports.ProcessorController) *WorkersView {
	return &WorkersView{
		proc:         proc,
		workerCount:  proc.GetWorkerCount(),
		maxWorkers:   runtime.NumCPU(),
		tickInterval: 500 * time.Millisecond,
		viewport:     viewport.New(0, 0),
	}
}

// Update handles messages for the workers view
func (v *WorkersView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle viewport scrolling first
		switch {
		case key.Matches(msg, kbind.PageUp):
			v.viewport.PageUp()
			return nil
		case key.Matches(msg, kbind.PageDown):
			v.viewport.PageDown()
			return nil
		case key.Matches(msg, kbind.GotoTop):
			v.viewport.GotoTop()
			return nil
		case key.Matches(msg, kbind.GotoBottom):
			v.viewport.GotoBottom()
			return nil
		case key.Matches(msg, kbind.Up):
			v.viewport.ScrollUp(1)
			return nil
		case key.Matches(msg, kbind.Down):
			v.viewport.ScrollDown(1)
			return nil
		}

		switch {
		case key.Matches(msg, kbind.Left):
			if v.workerCount > 1 {
				v.workerCount--
				v.proc.SetWorkers(v.workerCount)
			}
			return nil

		case key.Matches(msg, kbind.Right):
			if v.workerCount < v.maxWorkers {
				v.workerCount++
				v.proc.SetWorkers(v.workerCount)
			}
			return nil

		case key.Matches(msg, kbind.WorkerDec):
			if v.workerCount > 1 {
				v.workerCount--
				v.proc.SetWorkers(v.workerCount)
			}
			return nil

		case key.Matches(msg, kbind.WorkerInc):
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
	v.viewport.Width = width - 4
	v.viewport.Height = height - 2
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
	totalValue := metricValueStyle.Render(strconv.FormatUint(v.lastMetrics.TotalRequests, 10))
	b.WriteString(totalLabel)
	b.WriteString(" ")
	b.WriteString(totalValue)
	b.WriteString("\n")

	// Success requests
	successLabel := metricLabelStyle.Render("✓ Success:")
	successStyle := metricValueStyle.Foreground(lipgloss.Color("40"))
	successValue := successStyle.Render(strconv.FormatUint(v.lastMetrics.SuccessRequests, 10))
	b.WriteString(successLabel)
	b.WriteString(" ")
	b.WriteString(successValue)
	b.WriteString("\n")

	// Error requests
	errorLabel := metricLabelStyle.Render("✗ Errors:")
	var errorValue string
	if v.lastMetrics.ErrorRequests > 0 {
		errorStyle := metricValueStyle.Foreground(lipgloss.Color("196"))
		errorValue = errorStyle.Render(strconv.FormatUint(v.lastMetrics.ErrorRequests, 10))
	} else {
		errorValue = metricValueStyle.Render("0")
	}
	b.WriteString(errorLabel)
	b.WriteString(" ")
	b.WriteString(errorValue)
	b.WriteString("\n")

	// Lines processed
	linesLabel := metricLabelStyle.Render("Lines Processed:")
	linesValue := metricValueStyle.Render(strconv.FormatUint(v.lastMetrics.LinesProcessed, 10))
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
	activeWorkersValue := metricValueStyle.Render(strconv.Itoa(v.lastMetrics.ActiveWorkers))
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

	// Set viewport content
	v.viewport.SetContent(b.String())

	// Render viewport with scroll indicators
	viewportView := v.viewport.View()
	scrollIndicator := v.renderScrollIndicator()

	if scrollIndicator != "" {
		return workersAppStyle.Render(lipgloss.JoinVertical(lipgloss.Left, viewportView, scrollIndicator))
	}

	return workersAppStyle.Render(viewportView)
}

// renderScrollIndicator renders scroll position indicator if content is scrollable
func (v *WorkersView) renderScrollIndicator() string {
	if v.viewport.TotalLineCount() <= v.viewport.Height {
		return ""
	}

	scrollPercentage := int(v.viewport.ScrollPercent() * 100)
	indicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center)

	var indicator string
	if scrollPercentage <= 0 {
		indicator = "↓ Scroll down for more"
	} else if scrollPercentage >= 100 {
		indicator = "↑ Scroll up to see more"
	} else {
		indicator = fmt.Sprintf("↑ %d%% ↓", scrollPercentage)
	}

	return indicatorStyle.Render(indicator)
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

	for i := range sliderWidth {
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
