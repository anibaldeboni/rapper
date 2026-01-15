package views

import (
	"strings"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	logTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1).Bold(true)
)

// LogsView displays execution logs
type LogsView struct {
	viewport     viewport.Model
	logger       logs.Logger
	title        string
	width        int
	height       int
	isProcessing bool
	autoScroll   bool
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// User scrolled manually - disable auto-scroll
		if key.Matches(msg, key.NewBinding(key.WithKeys("up", "down", "pgup", "pgdown", "home"))) {
			v.autoScroll = false
		}
		// Re-enable auto-scroll if user explicitly goes to bottom
		if key.Matches(msg, key.NewBinding(key.WithKeys("end"))) {
			v.autoScroll = true
		}

		if key.Matches(msg, key.NewBinding(key.WithKeys("up"))) {
			v.ScrollUp(1)
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("down"))) {
			v.ScrollDown(1)
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("pgup"))) {
			v.ScrollUp(v.height)
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("pgdown"))) {
			v.ScrollDown(v.height)
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("home"))) {
			v.viewport.GotoTop()
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("end"))) {
			v.viewport.GotoBottom()
		}
	}
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

// Resize updates the view dimensions
func (v *LogsView) Resize(width, height int) {
	v.width = width
	v.height = height
	v.viewport.Width = width
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
				"",
				v.viewport.View(),
			),
		)
}

// UpdateLogs updates the viewport content with latest logs and auto-scrolls if processing
func (v *LogsView) UpdateLogs() {
	content := strings.Join(v.logger.Get(), "\n")
	v.viewport.SetContent(content)
	if v.isProcessing && v.autoScroll {
		v.viewport.GotoBottom()
	}
}

// SetProcessing updates the processing state
func (v *LogsView) SetProcessing(isProcessing bool) {
	v.isProcessing = isProcessing
	if isProcessing {
		v.autoScroll = true
	}
}

// ScrollUp scrolls the viewport up
func (v *LogsView) ScrollUp(lines int) {
	v.viewport.ScrollUp(lines)
}

// ScrollDown scrolls the viewport down
func (v *LogsView) ScrollDown(lines int) {
	v.viewport.ScrollDown(lines)
}
