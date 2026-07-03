package logs

import (
	"fmt"
	"time"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/web"
)

// LogType classifies a log message by source and severity. The
// renderer (LogMessageRenderer in internal/ui/components) maps each
// value to a color; the processor populates Type from the HTTP status
// code via NewHTTPMessage, and free-form messages default to General
// via NewGeneralMessage.
type LogType int

const (
	// LogTypeGeneral covers free-form log lines (cancellations, CSV
	// read errors, "processing file X" notifications, etc.). Anything
	// that is not an HTTP response lands here.
	LogTypeGeneral LogType = iota

	// LogTypeSuccess is for HTTP responses with a 2xx status code.
	LogTypeSuccess

	// LogTypeClientError is for HTTP responses with a 4xx status code.
	LogTypeClientError

	// LogTypeServerError is for HTTP responses with a 5xx status code.
	LogTypeServerError
)

// LogMessage is the in-memory representation of a log entry. It
// carries enough structured data (status, method, URL, body) for the
// TUI renderer to show a color-coded, expandable list item, and
// enough free-form data (icon, kind, text) to keep the existing
// general-purpose log lines (cancellations, processing notices, etc.)
// looking the same as before.
type LogMessage struct {
	// Type drives the row color in the TUI renderer.
	Type LogType

	// StatusCode, Method, URL, Body are populated for HTTP responses
	// (NewHTTPMessage). Free-form messages (NewGeneralMessage) leave
	// them at their zero values.
	StatusCode int
	Method     string
	URL        string
	Body       []byte

	// Icon, Kind, Text are populated for free-form messages. The HTTP
	// constructor leaves them empty; the renderer synthesises an
	// appropriate row title from Method+URL+StatusCode.
	Icon string
	Kind string
	Text string

	// Timestamp records when the message was constructed. Used by
	// downstream consumers for recency sorting; not part of the
	// rendered output.
	Timestamp time.Time
}

// NewHTTPMessage builds a LogMessage from an HTTP response. The
// LogType is derived from the status code so the renderer can pick
// the right color without re-deriving the classification itself.
func NewHTTPMessage(res web.Response) LogMessage {
	return LogMessage{
		Type:       classifyStatus(res.StatusCode),
		StatusCode: res.StatusCode,
		Method:     res.Method,
		URL:        res.URL,
		Body:       res.Body,
		Timestamp:  time.Now(),
	}
}

// NewGeneralMessage builds a free-form LogMessage. Type is forced to
// LogTypeGeneral — the renderer cannot color a free-form line
// differently from any other free-form line based on the icon alone.
func NewGeneralMessage(icon, kind, text string) LogMessage {
	return LogMessage{
		Type:      LogTypeGeneral,
		Icon:      icon,
		Kind:      kind,
		Text:      text,
		Timestamp: time.Now(),
	}
}

// String renders the message as a single human-readable line. The
// format matches the legacy builder output ("<icon> [<Kind>] <text>")
// for backward compatibility with log files, golden tests, and any
// downstream consumer that only sees the string form.
//
// HTTP responses use Method + URL + status code as the visible text —
// the body is shown on demand by the TUI renderer, not in the line
// itself.
func (m LogMessage) String() string {
	if m.Type == LogTypeSuccess || m.Type == LogTypeClientError || m.Type == LogTypeServerError {
		return fmt.Sprintf("%s %s %d", m.Method, m.URL, m.StatusCode)
	}
	var icon, kind string
	if m.Kind != "" {
		kind = fmt.Sprintf("[%s] ", styles.Bold(m.Kind))
	}
	if m.Icon != "" {
		icon = m.Icon + " "
	}
	return icon + kind + m.Text
}

// classifyStatus maps an HTTP status code to a LogType. 2xx is
// success, 4xx is a client error, 5xx is a server error; everything
// else (including 0, 1xx, 3xx) is general — transport failures and
// redirects don't carry enough semantic weight to deserve their own
// color in the logs view.
func classifyStatus(code int) LogType {
	switch {
	case code >= 200 && code < 300:
		return LogTypeSuccess
	case code >= 400 && code < 500:
		return LogTypeClientError
	case code >= 500 && code < 600:
		return LogTypeServerError
	default:
		return LogTypeGeneral
	}
}
