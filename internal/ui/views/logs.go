package views

import (
	"strings"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	logTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1).Bold(true)
)

// LogsView displays execution logs
type LogsView struct {
	viewport viewport.Model
	logger   logs.Logger
	title    string
	width    int
	height   int
}

// NewLogsView creates a new LogsView
func NewLogsView(logger logs.Logger) *LogsView {
	vp := viewport.New(60, 20)

	return &LogsView{
		viewport: vp,
		logger:   logger,
		title:    "Execution logs",
	}
}

// Update handles messages for the logs view
func (v *LogsView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

// Resize updates the view dimensions
func (v *LogsView) Resize(width, height int) {
	v.width = width
	v.height = height
	v.viewport.Width = width - 4
	v.viewport.Height = height - 6
}

// View renders the logs view
func (v *LogsView) View() string {
	titleBar := logTitleStyle.Render(v.title)

	return lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				titleBar,
				v.viewport.View(),
			),
		)
}

// UpdateContent updates the viewport content with latest logs
func (v *LogsView) UpdateContent() {
	content := strings.Join(v.logger.Get(), "\n")
	v.viewport.SetContent(content)
	v.viewport.GotoBottom()
}

// ScrollUp scrolls the viewport up
func (v *LogsView) ScrollUp(lines int) {
	v.viewport.ScrollUp(lines)
}

// ScrollDown scrolls the viewport down
func (v *LogsView) ScrollDown(lines int) {
	v.viewport.ScrollDown(lines)
}
