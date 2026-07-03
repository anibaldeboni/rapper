package components

import (
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/tidwall/pretty"
)

// Row-color palette owned by the log renderer. Kept private so the
// palette cannot leak into other packages; tests verify the
// invariant (one color per LogType) without locking the exact
// ANSI codes.
var (
	logRowGeneralStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	logRowSuccessStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
	logRowClientErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	logRowServerErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	logRowSelectedStyle  = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("237"))
	logRowSelectedGenSt  = logRowSelectedStyle.Foreground(lipgloss.Color("245"))
	logRowSelectedSuccSt = logRowSelectedStyle.Foreground(lipgloss.Color("40"))
	logRowSelectedCliSt  = logRowSelectedStyle.Foreground(lipgloss.Color("214"))
	logRowSelectedServSt = logRowSelectedStyle.Foreground(lipgloss.Color("196"))
)

// LogMessageRenderer implements ItemRenderer[logs.LogMessage] so the
// generic DetailedList component can render the structured log
// payload without knowing anything about the logs domain. The
// renderer owns:
//
//   - Row title format (HTTP → "METHOD URL status"; free-form →
//     "icon [kind] text")
//   - Row detail (only HTTP responses; body is pretty-printed JSON
//     with ANSI coloring so the expanded view is readable in the
//     TUI)
//   - Row colors (one base style per LogType; selected rows are the
//     matching style + bold + background)
//
// Free-form messages (LogTypeGeneral with icon/kind/text) never
// produce a non-empty Detail — the DetailedList treats Enter on
// such a row as a no-op.
type LogMessageRenderer struct{}

// Compile-time check: LogMessageRenderer must implement
// ItemRenderer[LogMessage].
var _ ItemRenderer[logs.LogMessage] = LogMessageRenderer{}

// Title returns the single-line title for the log row.
//
//   - HTTP responses render as "METHOD URL statusCode" so the
//     user can scan status codes at a glance.
//   - Free-form messages render as the legacy "icon [Kind] text"
//     form, preserving the appearance every other part of the app
//     already expects.
func (LogMessageRenderer) Title(m logs.LogMessage) string {
	if m.Type == logs.LogTypeSuccess || m.Type == logs.LogTypeClientError || m.Type == logs.LogTypeServerError {
		return m.Method + " " + m.URL + " " + statusBadge(m.StatusCode)
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

// Style returns the per-row color for unselected rows. The choice
// is keyed off LogType so the user can scan a list and immediately
// see success vs error counts.
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

// SelectedStyle returns the per-row color for the currently focused
// row. Selected rows get the same foreground as the unselected
// style plus bold + a contrasting background so the cursor is easy
// to track in long lists.
func (r LogMessageRenderer) SelectedStyle(m logs.LogMessage) lipgloss.Style {
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

// statusBadge returns a short, color-free string for the HTTP
// status code. Kept as a separate helper so the test can assert
// the layout without depending on the renderer's color choices.
func statusBadge(code int) string {
	// Small, allocation-free formatting; lipgloss handles the color.
	return formatInt(code)
}

func formatInt(code int) string {
	// strconv.Itoa is the canonical choice; inlined to keep imports
	// stable across the file.
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
