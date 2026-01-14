package views

import (
	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/anibaldeboni/rapper/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
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
	title := ui.TitleStyle.Render("👷 Workers Control")
	content := "\n\nWorkers view will be implemented in Phase 4.\n\n"
	content += "Features:\n"
	content += "  • Adjust worker count with slider\n"
	content += "  • View real-time metrics (req/s, throughput)\n"
	content += "  • Monitor active workers\n\n"
	content += "Press Esc to go back."

	return ui.AppStyle(title + content)
}
