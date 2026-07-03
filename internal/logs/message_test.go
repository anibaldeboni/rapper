package logs_test

import (
	"testing"
	"time"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/web"
	"github.com/stretchr/testify/assert"
)

// TestLogMessage_NewHTTPMessage_ClassifiesByStatusCode proves that
// NewHTTPMessage picks the right LogType from the response status code:
// 2xx → Success, 4xx → ClientError, 5xx → ServerError, anything else
// (including 0/transport errors) → General.
func TestLogMessage_NewHTTPMessage_ClassifiesByStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantType   logs.LogType
	}{
		{name: "200 OK is success", statusCode: 200, wantType: logs.LogTypeSuccess},
		{name: "201 Created is success", statusCode: 201, wantType: logs.LogTypeSuccess},
		{name: "299 edge is success", statusCode: 299, wantType: logs.LogTypeSuccess},
		{name: "400 Bad Request is client error", statusCode: 400, wantType: logs.LogTypeClientError},
		{name: "404 Not Found is client error", statusCode: 404, wantType: logs.LogTypeClientError},
		{name: "499 edge is client error", statusCode: 499, wantType: logs.LogTypeClientError},
		{name: "500 Internal Server Error is server error", statusCode: 500, wantType: logs.LogTypeServerError},
		{name: "503 Service Unavailable is server error", statusCode: 503, wantType: logs.LogTypeServerError},
		{name: "0 (no response) is general", statusCode: 0, wantType: logs.LogTypeGeneral},
		{name: "1xx informational is general", statusCode: 100, wantType: logs.LogTypeGeneral},
		{name: "3xx redirect is general", statusCode: 301, wantType: logs.LogTypeGeneral},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := web.Response{
				URL:        "https://example.com/users/1",
				StatusCode: tt.statusCode,
				Method:     "GET",
				Body:       []byte(`{"id":1}`),
			}
			msg := logs.NewHTTPMessage(res)
			assert.Equal(t, tt.wantType, msg.Type, "LogType for status %d", tt.statusCode)
			assert.Equal(t, tt.statusCode, msg.StatusCode)
			assert.Equal(t, res.URL, msg.URL)
			assert.Equal(t, res.Method, msg.Method)
			assert.Equal(t, res.Body, msg.Body)
		})
	}
}

// TestLogMessage_NewGeneralMessage_PopulatesFields proves the
// free-form constructor sets Icon/Kind/Text and defaults the Type to
// General (no HTTP response context).
func TestLogMessage_NewGeneralMessage_PopulatesFields(t *testing.T) {
	msg := logs.NewGeneralMessage("⚠️", "Request", "something went wrong")

	assert.Equal(t, logs.LogTypeGeneral, msg.Type)
	assert.Equal(t, "⚠️", msg.Icon)
	assert.Equal(t, "Request", msg.Kind)
	assert.Equal(t, "something went wrong", msg.Text)
	assert.Empty(t, msg.URL)
	assert.Empty(t, msg.Method)
	assert.Zero(t, msg.StatusCode)
}

// TestLogMessage_String_GeneralIncludesIconKindAndText — the legacy
// String() method must produce a stable, human-readable form for general
// messages so log consumers and golden tests keep working.
func TestLogMessage_String_GeneralIncludesIconKindAndText(t *testing.T) {
	tests := []struct {
		name     string
		icon     string
		kind     string
		text     string
		contains []string
	}{
		{
			name:     "all fields set",
			icon:     "💀",
			kind:     "Cancelation",
			text:     "Read 5 lines",
			contains: []string{"💀", "Cancelation", "Read 5 lines"},
		},
		{
			name:     "icon only",
			icon:     "ℹ️",
			text:     "please wait",
			contains: []string{"ℹ️", "please wait"},
		},
		{
			name:     "text only",
			text:     "no extras",
			contains: []string{"no extras"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := logs.NewGeneralMessage(tt.icon, tt.kind, tt.text)
			out := msg.String()
			for _, want := range tt.contains {
				assert.Contains(t, out, want, "String()=%q must contain %q", out, want)
			}
		})
	}
}

// TestLogMessage_TimestampSetOnConstruct proves the constructors
// stamp a Timestamp so downstream consumers can sort or filter logs by
// recency. The exact value is non-deterministic so we just check it is
// recent (within the last minute).
func TestLogMessage_TimestampSetOnConstruct(t *testing.T) {
	before := time.Now()
	msg := logs.NewGeneralMessage("", "", "hi")
	after := time.Now()

	assert.False(t, msg.Timestamp.Before(before), "Timestamp must be >= before")
	assert.False(t, msg.Timestamp.After(after), "Timestamp must be <= after")
}
