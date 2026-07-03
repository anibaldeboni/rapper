package components

import (
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/tidwall/pretty"
)

// Badge and row-color palette owned by the log renderer.
//
// Badge styles: colored background pill on the left of each HTTP row.
// Row styles: foreground-only color applied to the URL text.
// Selected styles: type-tinted dark background so the cursor is easy
// to track without washing out the badge colors embedded in the title.
var (
	// Badge backgrounds — each is a solid color block with white text.
	logBadgeSuccessStyle   = lipgloss.NewStyle().Background(lipgloss.Color("34")).Foreground(lipgloss.Color("255")).Bold(true).Padding(0, 1)
	logBadgeClientErrStyle = lipgloss.NewStyle().Background(lipgloss.Color("166")).Foreground(lipgloss.Color("255")).Bold(true).Padding(0, 1)
	logBadgeServerErrStyle = lipgloss.NewStyle().Background(lipgloss.Color("124")).Foreground(lipgloss.Color("255")).Bold(true).Padding(0, 1)

	// Row foreground — applied to the URL text on unselected rows.
	logRowGeneralStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	logRowSuccessStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
	logRowClientErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	logRowServerErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	// Selected row — type-tinted dark background so the badge colours
	// remain readable inside the highlighted row.
	logRowSelectedGenSt  = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("237")).Foreground(lipgloss.Color("245"))
	logRowSelectedSuccSt = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("22")).Foreground(lipgloss.Color("40"))
	logRowSelectedCliSt  = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("58")).Foreground(lipgloss.Color("214"))
	logRowSelectedServSt = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("52")).Foreground(lipgloss.Color("196"))
)

// LogMessageRenderer implements ItemRenderer[logs.LogMessage] so the
// generic DetailedList component can render the structured log
// payload without knowing anything about the logs domain. The
// renderer owns:
//
//   - Row title format: HTTP rows render as "[BADGE] URL" where the
//     badge is a colored pill showing the status code; free-form rows
//     render as "icon [Kind] text".
//   - Row detail: only HTTP responses; body is pretty-printed JSON
//     with ANSI coloring so the expanded view is readable in the TUI.
//   - Row colors: one badge style + one row style per LogType.
//
// Free-form messages (LogTypeGeneral) never produce a non-empty
// Detail — the DetailedList treats Enter on such a row as a no-op.
type LogMessageRenderer struct{}

// Compile-time check: LogMessageRenderer must implement
// ItemRenderer[LogMessage].
var _ ItemRenderer[logs.LogMessage] = LogMessageRenderer{}

// Title returns the single-line title for the log row.
//
//   - HTTP responses render as "[STATUS_BADGE] URL" — a coloured pill
//     carrying the numeric status code, followed by the request URL.
//     The badge is rendered inline with its own ANSI escape codes so
//     it keeps its background colour even inside a selected row.
//   - Free-form messages render as the legacy "icon [Kind] text" form.
func (LogMessageRenderer) Title(m logs.LogMessage) string {
	switch m.Type {
	case logs.LogTypeSuccess:
		return logBadgeSuccessStyle.Render(formatInt(m.StatusCode)) + "  " + m.URL
	case logs.LogTypeClientError:
		return logBadgeClientErrStyle.Render(formatInt(m.StatusCode)) + "  " + m.URL
	case logs.LogTypeServerError:
		return logBadgeServerErrStyle.Render(formatInt(m.StatusCode)) + "  " + m.URL
	}
	return m.String()
}

// Detail returns the pretty-printed response body for HTTP rows, or
// "" for free-form rows. The empty-string signal is the contract
// the DetailedList component uses to decide whether Enter is a
// no-op (see DetailedList.Update).
//
// tidwall/pretty.Pretty reformats the body (indents, normalises
// spacing); pretty.Color adds ANSI colors so the JSON stands out
// in a terminal that supports them.
func (LogMessageRenderer) Detail(m logs.LogMessage) string {
	if m.Type != logs.LogTypeSuccess && m.Type != logs.LogTypeClientError && m.Type != logs.LogTypeServerError {
		return ""
	}
	if len(m.Body) == 0 {
		return ""
	}
	return string(pretty.Color(pretty.Pretty(m.Body), nil))
}

// Style returns the per-row foreground style for unselected rows.
// The badge inside the title carries its own background, so the row
// style only needs to set the foreground for the URL text.
func (LogMessageRenderer) Style(m logs.LogMessage) lipgloss.Style {
	switch m.Type {
	case logs.LogTypeSuccess:
		return logRowSuccessStyle
	case logs.LogTypeClientError:
		return logRowClientErrStyle
	case logs.LogTypeServerError:
		return logRowServerErrStyle
	default:
		return logRowGeneralStyle
	}
}

// SelectedStyle returns the per-row style for the currently focused
// row. Selected rows get a type-tinted dark background so the cursor
// is easy to track; the badge inside the title retains its own inline
// ANSI codes and therefore keeps its foreground/background regardless
// of the row background applied here.
func (LogMessageRenderer) SelectedStyle(m logs.LogMessage) lipgloss.Style {
	switch m.Type {
	case logs.LogTypeSuccess:
		return logRowSelectedSuccSt
	case logs.LogTypeClientError:
		return logRowSelectedCliSt
	case logs.LogTypeServerError:
		return logRowSelectedServSt
	default:
		return logRowSelectedGenSt
	}
}

// formatInt converts an integer to its decimal string representation.
// strconv.Itoa is the canonical choice; this version avoids importing
// strconv for a single call site.
func formatInt(code int) string {
	if code == 0 {
		return "0"
	}
	neg := code < 0
	if neg {
		code = -code
	}
	var buf [20]byte
	i := len(buf)
	for code > 0 {
		i--
		buf[i] = byte('0' + code%10)
		code /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
