package logs

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/anibaldeboni/rapper/internal/web"
	"github.com/tidwall/pretty"
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

	// LogTypeWarning is for HTTP responses with a 4xx status code.
	LogTypeWarning

	// LogTypeError is for HTTP responses with a 5xx status code.
	LogTypeError
)

// LogMessage is the in-memory representation of a log entry. It
// carries enough structured data (status, method, URL, body) for the
// TUI renderer to show a color-coded, expandable list item, and
// enough free-form data (icon, kind, text) to keep the existing
// general-purpose log lines (cancellations, processing notices, etc.)
// looking the same as before.
type LogMessage struct {
	// Type drives the row color in the TUI renderer.
	Type      LogType
	BadgeIcon string
	Text      string
	Details   string
	Timestamp time.Time
}

// NewHTTPMessage builds a LogMessage from an HTTP response. The
// LogType is derived from the status code so the renderer can pick
// the right color without re-deriving the classification itself.
func NewHTTPMessage(res web.Response) LogMessage {
	body := string(pretty.Color(pretty.Pretty(res.Body), nil))
	return LogMessage{
		Type:      classifyStatus(res.StatusCode),
		BadgeIcon: strconv.Itoa(res.StatusCode),
		Text:      res.Method + " " + res.URL,
		Details:   body,
		Timestamp: time.Now(),
	}
}

// NewGeneralMessage builds a free-form LogMessage. Type is forced to
// LogTypeGeneral — the renderer cannot color a free-form line
// differently from any other free-form line based on the icon alone.
func NewGeneralMessage(icon, kind, title string, configs ...MessageConfig) LogMessage {
	msg := LogMessage{
		Type:      LogTypeGeneral,
		BadgeIcon: icon,
		Text:      title,
		Timestamp: time.Now(),
	}
	for _, config := range configs {
		config(&msg)
	}
	return msg
}

func NewMessage(title string, configs ...MessageConfig) LogMessage {
	msg := LogMessage{
		Type:      LogTypeGeneral,
		Text:      title,
		Timestamp: time.Now(),
	}
	for _, config := range configs {
		config(&msg)
	}
	return msg
}

type MessageConfig func(*LogMessage)

func WithDetail(detail string) MessageConfig {
	return func(m *LogMessage) {
		m.Details = detail
	}
}

func WithIcon(icon string) MessageConfig {
	return func(m *LogMessage) {
		m.BadgeIcon = icon
	}
}

func AsGeneral() MessageConfig {
	return func(m *LogMessage) {
		m.Type = LogTypeGeneral
	}
}

func AsError() MessageConfig {
	return func(m *LogMessage) {
		m.Type = LogTypeError
	}
}

func AsWarning() MessageConfig {
	return func(m *LogMessage) {
		m.Type = LogTypeWarning
	}
}

func AsSuccess() MessageConfig {
	return func(m *LogMessage) {
		m.Type = LogTypeSuccess
	}
}

func (m LogMessage) String() string {
	return collapseSpaces(m.BadgeIcon + m.Text + m.Details)
}

func collapseSpaces(s string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
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
		return LogTypeWarning
	case code >= 500 && code < 600:
		return LogTypeError
	default:
		return LogTypeGeneral
	}
}
