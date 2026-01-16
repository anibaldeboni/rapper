package views

import (
	"strings"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	logTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).MarginBottom(1).Padding(0, 1).Bold(true)
)

// LogsView displays execution logs
type LogsView struct {
	viewport   viewport.Model
	logger     logs.Logger
	title      string
	width      int
	height     int
	autoScroll bool
}

// NewLogsView creates a new LogsView
func NewLogsView(logger logs.Logger) *LogsView {
	vp := viewport.New(0, 0)

	v := &LogsView{
		viewport:   vp,
		logger:     logger,
		title:      "📝 Execution logs",
		autoScroll: true,
	}
	// Load initial logs
	v.updateLogs()
	return v
}

// Update handles messages for the logs view
func (v *LogsView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// Handle any processing-related messages by checking their content
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
	default:
		// Check for processing messages using reflection
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

// Resize updates the view dimensions
func (v *LogsView) Resize(width, height int) {
	v.width = width
	v.height = height
	v.viewport.Width = (width / 2) - 2
	v.viewport.Height = height - 3
}

// View renders the logs view
func (v *LogsView) View() string {
	return lipgloss.NewStyle().
		MarginLeft(2).
		MarginTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				logTitleStyle.Render(v.title),
				v.viewport.View(),
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
