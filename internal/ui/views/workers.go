package views

import (
	"github.com/anibaldeboni/rapper/internal/processor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	workersTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1).Bold(true)
	workersAppStyle   = lipgloss.NewStyle().Margin(1, 1, 1, 2).Render
)

// WorkersView displays and controls worker pool
type WorkersView struct {
	processor processor.Processor
	width     int
	height    int
}

// NewWorkersView creates a new WorkersView
func NewWorkersView(processor processor.Processor) *WorkersView {
	return &WorkersView{
		processor: processor,
	}
}

// Update handles messages for the workers view
func (v *WorkersView) Update(msg tea.Msg) tea.Cmd {
	// TODO: Implement in Phase 4
	return nil
}

// Resize updates the view dimensions
func (v *WorkersView) Resize(width, height int) {
	v.width = width
	v.height = height
}

// View renders the workers view
func (v *WorkersView) View() string {
	// TODO: Implement in Phase 4
	// For now, show a placeholder
	title := workersTitleStyle.Render("👷 Workers Control")
	content := "\n\nWorkers view will be implemented in Phase 4.\n\n"
	content += "Features:\n"
	content += "  • Adjust worker count with slider\n"
	content += "  • View real-time metrics (req/s, throughput)\n"
	content += "  • Monitor active workers\n\n"
	content += "Press Esc to go back."

	return workersAppStyle(title + content)
}
