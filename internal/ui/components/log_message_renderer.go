package components

import (
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/logs"
)

// Compile-time check: LogMessageRenderer must implement
// ItemRenderer[LogMessage].
var _ ItemRenderer[logs.LogMessage] = LogMessageRenderer{}

var (
	// Badge backgrounds — each is a solid color block with white text.
	logBadgeSuccessStyle   = lipgloss.NewStyle().Background(lipgloss.Color("34")).Foreground(lipgloss.Color("255")).Bold(true).Padding(0, 1)
	logBadgeClientErrStyle = lipgloss.NewStyle().Background(lipgloss.Color("166")).Foreground(lipgloss.Color("255")).Bold(true).Padding(0, 1)
	logBadgeServerErrStyle = lipgloss.NewStyle().Background(lipgloss.Color("124")).Foreground(lipgloss.Color("255")).Bold(true).Padding(0, 1)
	logBadgeGeneralStyle   = lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("255")).Bold(true).Padding(0, 1)

	logSelectedColor      = lipgloss.Color("#414141")
	logRowBackgroundColor = lipgloss.Color("#2e2f29")
	logRowStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#dbdbdb")).Background(logRowBackgroundColor).Padding(0, 1)
	logRowSelectedStyle   = logRowStyle.Bold(true).Background(logSelectedColor)
)

type LogMessageRenderer struct{}

// Title returns the single-line title for the log row.
func (LogMessageRenderer) Title(m logs.LogMessage, selected bool) string {
	titleStyle := logRowStyle
	if selected {
		titleStyle = logRowSelectedStyle
	}
	badgeStyle := logBadgeGeneralStyle
	switch m.Type {
	case logs.LogTypeSuccess:
		badgeStyle = logBadgeSuccessStyle
	case logs.LogTypeWarning:
		badgeStyle = logBadgeClientErrStyle
	case logs.LogTypeError:
		badgeStyle = logBadgeServerErrStyle
	}

	return badgeStyle.Render(m.BadgeIcon) + titleStyle.Render(m.Text)
}

func (lmr LogMessageRenderer) Detail(m logs.LogMessage) string {
	if len(m.Details) == 0 {
		return ""
	}
	width := lipgloss.Width(lmr.Title(m, false))
	return lipgloss.NewStyle().
		Background(logRowBackgroundColor).
		Width(width).
		MaxWidth(width).
		PaddingTop(1).
		PaddingLeft(2).
		Render(m.Details)
}

func (LogMessageRenderer) Style(m logs.LogMessage) lipgloss.Style {
	return logRowStyle
}

func (LogMessageRenderer) SelectedStyle(m logs.LogMessage) lipgloss.Style {
	return logRowSelectedStyle
}
