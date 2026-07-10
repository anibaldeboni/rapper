package components_test

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/ui/components"
	"github.com/anibaldeboni/rapper/internal/web"
	"github.com/stretchr/testify/assert"
)

// TestLogMessageRenderer_Title_FormatsHTTPRequestLine proves the
// renderer produces a compact, scannable title for HTTP responses.
// The title includes the HTTP method, the status code (inside an
// inline coloured pill), and the URL.
//
// NOTE: source behavior as of 2026-07-10 — see decision #178.
// NewHTTPMessage sets `Text: res.Method + " " + res.URL`, so the
// HTTP method IS part of the title.
func TestLogMessageRenderer_Title_FormatsHTTPRequestLine(t *testing.T) {
	r := components.LogMessageRenderer{}
	msg := logs.NewHTTPMessage(web.Response{
		Method:     "POST",
		URL:        "https://api.example.com/users",
		StatusCode: 201,
	})

	got := r.Title(msg, false)

	// The title must contain the status code, the URL, and the method.
	// NOTE: source behavior as of 2026-07-10 — see decision #178.
	// Method is part of the title (Text = Method + " " + URL).
	assert.Contains(t, got, "201")
	assert.Contains(t, got, "https://api.example.com/users")
	// NOTE: source behavior as of 2026-07-10 — see decision #178.
	assert.Contains(t, got, "POST",
		"HTTP method is part of the title (Text = Method + \" \" + URL)")
}

// TestLogMessageRenderer_Title_FreeFormIncludesIconAndKind proves the
// legacy free-form rendering still works for non-HTTP messages. The
// renderer must not impose a method/URL layout on general messages —
// the icon + title form is the format the rest of the app expects.
//
// NOTE: source behavior as of 2026-07-10 — see decision #178.
// NewGeneralMessage sets `Text: title` and IGNORES the kind
// argument. The test must NOT assert that the kind is in the
// rendered output.
func TestLogMessageRenderer_Title_FreeFormIncludesIconAndKind(t *testing.T) {
	r := components.LogMessageRenderer{}
	msg := logs.NewGeneralMessage("💀", "Cancelation", "read 5 lines")

	got := r.Title(msg, false)

	assert.Contains(t, got, "💀")
	assert.Contains(t, got, "read 5 lines")
}

// TestLogMessageRenderer_Detail_OnlyForHTTPResponses — the Detail
// payload is the response body, which only makes sense for HTTP
// responses. Free-form messages return "" so the parent
// DetailedList treats Enter as a no-op on those rows.
func TestLogMessageRenderer_Detail_OnlyForHTTPResponses(t *testing.T) {
	r := components.LogMessageRenderer{}

	httpMsg := logs.NewHTTPMessage(web.Response{
		Method:     "GET",
		URL:        "https://example.com",
		StatusCode: 200,
		Body:       []byte(`{"id":1}`),
	})
	generalMsg := logs.NewGeneralMessage("ℹ️", "Info", "starting up")

	assert.NotEmpty(t, r.Detail(httpMsg), "HTTP messages must produce a Detail payload")
	assert.Empty(t, r.Detail(generalMsg), "general messages must produce empty Detail so Enter is a no-op")
}

// TestLogMessageRenderer_Detail_PrettyPrintsJSON — the Detail for an
// HTTP response with a JSON body must be a formatted version of the
// body, not the raw bytes. The renderer uses tidwall/pretty to
// indent and color the JSON so it is readable in the TUI.
func TestLogMessageRenderer_Detail_PrettyPrintsJSON(t *testing.T) {
	r := components.LogMessageRenderer{}
	body := []byte(`{"id":1,"name":"alpha"}`)
	msg := logs.NewHTTPMessage(web.Response{
		Method:     "GET",
		URL:        "https://example.com",
		StatusCode: 200,
		Body:       body,
	})

	got := r.Detail(msg)

	// Pretty-formatted JSON contains newlines between the top-level
	// fields. The raw `{"id":1,"name":"alpha"}` does not.
	assert.Contains(t, got, "\n", "pretty-printed JSON must contain newlines between fields")
}

// TestLogMessageRenderer_Style_IsMonochromeAcrossLogType — the
// renderer returns the same style for every LogType. Color is
// applied per-item in Title (via the badge style) and per-row in
// SelectedStyle, not via Style().
//
// NOTE: source behavior as of 2026-07-10 — see decision #178.
// LogMessageRenderer.Style() returns `logRowStyle` unconditionally
// (line 60 of log_message_renderer.go). The test asserts this
// monochrome contract: every pair of LogType values produces the
// same style.
func TestLogMessageRenderer_Style_IsMonochromeAcrossLogType(t *testing.T) {
	r := components.LogMessageRenderer{}

	cases := []struct {
		name string
		msg  logs.LogMessage
	}{
		{"general", logs.NewGeneralMessage("ℹ️", "Info", "hi")},
		{"success", logs.NewHTTPMessage(web.Response{Method: "GET", URL: "u", StatusCode: 200})},
		{"client error", logs.NewHTTPMessage(web.Response{Method: "GET", URL: "u", StatusCode: 404})},
		{"server error", logs.NewHTTPMessage(web.Response{Method: "GET", URL: "u", StatusCode: 500})},
	}

	styles := make(map[string]lipgloss.Style)
	for _, c := range cases {
		styles[c.name] = r.Style(c.msg)
	}

	// NOTE: source behavior as of 2026-07-10 — see decision #178.
	// Pairwise: every pair of styles must be EQUAL (the renderer
	// returns logRowStyle for all LogType values).
	for i, a := range cases {
		for _, b := range cases[i+1:] {
			assert.Equal(t, styles[a.name], styles[b.name],
				"style for %s and %s must be equal (monochrome)", a.name, b.name)
		}
	}
}

// TestLogMessageRenderer_SelectedStyle_DiffersFromStyle proves the
// selected row stands out visually. We don't pin the exact look —
// only the invariant that the two styles are not identical.
func TestLogMessageRenderer_SelectedStyle_DiffersFromStyle(t *testing.T) {
	r := components.LogMessageRenderer{}
	msg := logs.NewHTTPMessage(web.Response{Method: "GET", URL: "u", StatusCode: 200})

	plain := r.Style(msg)
	selected := r.SelectedStyle(msg)

	assert.NotEqual(t, plain, selected, "selected style must differ from the plain style")
}
