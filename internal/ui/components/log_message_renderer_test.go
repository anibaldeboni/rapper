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
// renderer produces a compact, scannable title for HTTP responses:
// "[STATUS_BADGE] URL". The status code appears inside an inline
// coloured pill; the URL follows it. The HTTP method is intentionally
// omitted from the title — the URL already carries enough context.
func TestLogMessageRenderer_Title_FormatsHTTPRequestLine(t *testing.T) {
	r := components.LogMessageRenderer{}
	msg := logs.NewHTTPMessage(web.Response{
		Method:     "POST",
		URL:        "https://api.example.com/users",
		StatusCode: 201,
	})

	got := r.Title(msg)

	// The title must contain the status code and the URL.
	// Method is no longer part of the title (badge + URL format).
	assert.Contains(t, got, "201")
	assert.Contains(t, got, "https://api.example.com/users")
	// Verify the HTTP method is NOT in the title (design decision).
	assert.NotContains(t, got, "POST")
}

// TestLogMessageRenderer_Title_FreeFormIncludesIconAndKind proves the
// legacy free-form rendering still works for non-HTTP messages. The
// renderer must not impose a method/URL layout on general messages —
// the icon + kind + text form is the format the rest of the app
// expects.
func TestLogMessageRenderer_Title_FreeFormIncludesIconAndKind(t *testing.T) {
	r := components.LogMessageRenderer{}
	msg := logs.NewGeneralMessage("💀", "Cancelation", "read 5 lines")

	got := r.Title(msg)

	assert.Contains(t, got, "💀")
	assert.Contains(t, got, "Cancelation")
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

// TestLogMessageRenderer_Style_ColorsByLogType — each LogType
// produces a distinct style. The renderer owns the color palette so
// the generic DetailedList component stays domain-free.
func TestLogMessageRenderer_Style_ColorsByLogType(t *testing.T) {
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

	// Pairwise: every pair of styles must differ. The renderer's
	// contract is that the LogType changes the color; if two types
	// produce the same style, the user cannot tell them apart.
	for i, a := range cases {
		for _, b := range cases[i+1:] {
			assert.NotEqual(t, styles[a.name], styles[b.name],
				"style for %s must differ from %s", a.name, b.name)
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
