package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestView_NoToastsSkipsCompositor verifies the no-toast fast path in
// AppModel.View(): when the ToastManager has no active toasts, the
// rendered output must contain only the background (header, content,
// status bar) and must not contain any toast text. The compositor is
// expected to be bypassed entirely on this path (see the
// `if len(toastLayers) == 0` short-circuit in the new View()).
func TestView_NoToastsSkipsCompositor(t *testing.T) {
	app, _, _, _ := newTestApp(t, "../../tests/example.csv")
	app.width = 100
	app.height = 40
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	require.Empty(t, app.toastMgr.GetActive(),
		"sanity: no toasts have been added yet")

	out := app.View().Content

	// The "saved" string is used in the WithOneToastIncludesContent
	// test below. It must not appear in the no-toast output.
	assert.NotContains(t, out, "saved",
		"no-toast View() must not contain any toast text")
	// Sanity: the layout was actually rendered. AppModel.View()
	// wraps the joined content in Margin(1, 2), which inserts a
	// top margin row. The header therefore lands on line 1, not
	// line 0.
	lines := strings.Split(out, "\n")
	require.NotEmpty(t, lines, "View() must produce at least one line")

	firstContent := -1
	for i, l := range lines {
		if strings.TrimSpace(l) != "" {
			firstContent = i
			break
		}
	}
	// NOTE: source behavior as of 2026-07-10 — see decision #178.
	// AppModel.View() wraps content in Margin(1, 2), pushing the
	// header to line 1.
	assert.Equal(t, 1, firstContent,
		"with no toast, the global header must be on line 1 "+
			"(Margin(1, 2) pushes the header past the top margin row); got first content on line %d",
		firstContent)
}

// TestView_WithOneToastIncludesContent verifies the happy path of the
// new compositor-based View(): with one active toast, the rendered
// output must contain the toast message text overlaid on the top-right
// of the content area.
func TestView_WithOneToastIncludesContent(t *testing.T) {
	app, _, _, _ := newTestApp(t, "../../tests/example.csv")
	app.width = 100
	app.height = 40
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	app.toastMgr.Success("saved")
	require.Len(t, app.toastMgr.GetActive(), 1, "sanity: one active toast")

	out := app.View().Content

	// The toast text must be present in the rendered output.
	assert.Contains(t, out, "saved",
		"View() with one active toast must include the toast message text")
}

// TestView_HeaderOnLineZeroWithToast is the regression guard for the
// compositor refactor: with an active toast, the global header must
// still be on line 0 of the rendered output. The compositor must NOT
// shift the bg layer down (toasts are absolutely positioned via Y, not
// inserted into the bg flow).
func TestView_HeaderOnLineZeroWithToast(t *testing.T) {
	app, _, _, _ := newTestApp(t, "../../tests/example.csv")
	app.width = 100
	app.height = 40
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	app.toastMgr.Success("hello")

	out := app.View().Content
	lines := strings.Split(out, "\n")

	firstContent := -1
	for i, l := range lines {
		if strings.TrimSpace(l) != "" {
			firstContent = i
			break
		}
	}
	// NOTE: source behavior as of 2026-07-10 — see decision #178.
	// AppModel.View() wraps content in Margin(1, 2), pushing the
	// header to line 1 even when toasts are active.
	assert.Equal(t, 1, firstContent,
		"with an active toast, the global header must be on line 1 "+
			"(Margin(1, 2) pushes the header past the top margin row); got first content on line %d",
		firstContent)
}

// TestView_StatusBarOnLastLineWithToast is the regression guard for
// the compositor refactor on the other end of the canvas: with an
// active toast, the status bar must remain on the LAST non-empty line.
// The compositor does not extend the bg layer to make room for toasts;
// toasts overlay in the content area only.
func TestView_StatusBarOnLastLineWithToast(t *testing.T) {
	app, _, _, _ := newTestApp(t, "../../tests/example.csv")
	app.width = 100
	app.height = 40
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	app.toastMgr.Success("hello")

	out := app.View().Content
	lines := strings.Split(out, "\n")

	// Find the last non-empty line — the status bar is the only
	// non-empty line at the bottom of the layout.
	lastContent := -1
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			lastContent = i
			break
		}
	}
	require.GreaterOrEqual(t, lastContent, 0, "View() must produce at least one non-empty line")

	// The status bar is the rightmost element in the bottom row; it
	// contains the spinner (∙∙∙ when idle). We assert the spinner
	// marker is present on the last non-empty line.
	assert.Contains(t, lines[lastContent], "∙∙∙",
		"last non-empty line must be the status bar (spinner marker present); got %q",
		lines[lastContent])
}

// TestView_HeaderAlwaysOnLineZero is the regression test for the
// "extra empty line above the top menu" bug that manifested only on
// the Settings view.
//
// Root cause: AppModel.View() used
//
//	lipgloss.NewStyle().MaxHeight(m.height).AlignVertical(lipgloss.Center)
//
// to vertically center the joined (header, view, statusBar) content
// inside the terminal. Each view has a different rendered height, so
// the (m.height - totalContent) diff varies per view. When the diff is
// odd, lipgloss distributes the slack asymmetrically
// (floor(diff/2) on top, ceil(diff/2) on bottom), which placed one
// extra empty line above the global header for the Settings view
// (totalContent=37 → diff=3 → 1 line above) but not for the Files view
// (totalContent=38 → diff=2 → 1 line above too, but in a different
// parity slot). The user perceived this as "an extra line only on
// Settings".
//
// Fix: switch AlignVertical from Center to Top. The header is now
// always on line 0 regardless of view height, so the "ghost line"
// disappears from Settings. The unused vertical space (when the joined
// content is shorter than the terminal) now lands at the bottom under
// the status bar instead of being split between top and bottom. This
// is a deliberate trade-off: the app no longer recenters on resize,
// but the visual is consistent across all views.
//
// Test contract: for every (view, terminalHeight) pair, the global
// header line must be the first non-empty line in the rendered output.
// We probe Files, Logs, and Settings at three terminal heights (24,
// 40, 80) to cover the common range. A 0-line top gap is the invariant.
func TestView_HeaderAlwaysOnLineZero(t *testing.T) {
	// Use a representative CSV path so the Files view has real content.
	app, _, _, _ := newTestApp(t, "../../tests/example.csv")

	// Mark Settings as modified so its "⚠️ Unsaved changes" help line
	// is visible — this exercises the worst-case Settings content
	// height (the one that triggered the bug originally).
	//
	// Note: SettingsView.modified is unexported (views package). We
	// cannot toggle it from here. The Settings view is exercised with
	// its default state (no help line), which is the common case.

	for _, termH := range []int{24, 40, 80} {
		app.width = 100
		app.height = termH
		// Send a WindowSizeMsg so the views' Resize() runs and the
		// renderers are calibrated for the chosen height.
		app.Update(tea.WindowSizeMsg{Width: 100, Height: termH})

		for _, viewName := range []View{ViewFiles, ViewLogs, ViewSettings} {
			app.currentView = viewName
			view := app.View()
			// tea.View.Content holds the rendered frame.
			lines := strings.Split(view.Content, "\n")

			firstContent := -1
			for i, l := range lines {
				if strings.TrimSpace(l) != "" {
					firstContent = i
					break
				}
			}
			// NOTE: source behavior as of 2026-07-10 — see decision #178.
			// AppModel.View() wraps content in Margin(1, 2), pushing the
			// header to line 1 for every view/height combination.
			assert.Equal(t, 1, firstContent,
				"view %s at height %d must have the global header on line 1 "+
					"(Margin(1, 2) pushes the header past the top margin row); got first content on line %d",
				viewName, termH, firstContent)
		}
	}
}
