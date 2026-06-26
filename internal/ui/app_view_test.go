package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSpliceRight_StacksOverlayAtTopOfContentArea verifies that the
// toast overlay is stacked in the top-right of the content area instead
// of being scattered across the header, content, and status bar lines.
//
// The base string models the global app layout built by View():
//
//	line 0 → header
//	line 1 → first content line
//	line 2 → second content line
//	line 3 → status bar
//
// A 3-line overlay was previously spliced onto lines 0, 1, 2, scattering
// the toasts vertically. The fix: splice onto lines 1..len-2 (the
// content area only), so the toasts appear stacked at the top-right of
// the content.
func TestSpliceRight_StacksOverlayAtTopOfContentArea(t *testing.T) {
	header := "=== HEADER ==="
	content1 := "first content line"
	content2 := "second content line"
	status := "=== STATUS ==="
	base := strings.Join([]string{header, content1, content2, status}, "\n")

	overlay := strings.Join([]string{
		"TOAST ONE",
		"TOAST TWO",
		"TOAST THREE",
	}, "\n")

	got := spliceRight(base, overlay, 80)
	lines := strings.Split(got, "\n")

	// Header must be untouched.
	assert.Equal(t, header, lines[0], "header line must not be overwritten by an overlay line")

	// Status bar must be untouched.
	assert.Equal(t, status, lines[len(lines)-1], "status bar line must not be overwritten by an overlay line")

	// The first overlay line (TOAST ONE) must appear in the first
	// content line, not the header.
	assert.Contains(t, lines[1], "TOAST ONE", "first toast must land on the first content line")
	assert.Contains(t, lines[2], "TOAST TWO", "second toast must land on the second content line")

	// No toast should appear on the header or status bar.
	assert.NotContains(t, lines[0], "TOAST", "header must not contain any toast")
	assert.NotContains(t, lines[len(lines)-1], "TOAST", "status bar must not contain any toast")
}

// TestSpliceRight_EmptyOverlayReturnsBase guards the early-exit path
// because it is the most common no-op (no active toasts) and a
// regression there would break the empty-state layout.
func TestSpliceRight_EmptyOverlayReturnsBase(t *testing.T) {
	base := "header\ncontent\nstatus"
	assert.Equal(t, base, spliceRight(base, "", 80))
}

// TestSpliceRight_ShortBaseIgnoresOverlay verifies that when the base
// has no content area (only header + status, or fewer lines) the
// overlay is dropped instead of corrupting the header or status bar.
// Without this guard the toasts would still scatter onto whatever lines
// the base has.
func TestSpliceRight_ShortBaseIgnoresOverlay(t *testing.T) {
	cases := []struct {
		name string
		base string
	}{
		{"only header", "header"},
		{"header and status", "header\nstatus"},
		{"single line", "x"},
	}
	overlay := "TOAST"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := spliceRight(tc.base, overlay, 80)
			assert.Equal(t, tc.base, got, "overlay must be dropped when base has no content area")
			assert.NotContains(t, got, "TOAST", "no toast may land on the header or status bar")
		})
	}
}

// TestSpliceRight_MoreOverlayThanContentLinesStops verifies the fix
// does not write past the last content line. With a 3-line content
// area and a 5-line overlay, the extra overlay lines must be dropped
// and the status bar must stay untouched.
func TestSpliceRight_MoreOverlayThanContentLinesStops(t *testing.T) {
	header := "HEADER"
	c1 := "c1"
	c2 := "c2"
	c3 := "c3"
	status := "STATUS"
	base := strings.Join([]string{header, c1, c2, c3, status}, "\n")

	overlay := strings.Join([]string{
		"TOAST A",
		"TOAST B",
		"TOAST C",
		"TOAST D", // extra — must be dropped
		"TOAST E", // extra — must be dropped
	}, "\n")

	got := spliceRight(base, overlay, 80)
	lines := strings.Split(got, "\n")

	// Status bar must be untouched even with extra overlay lines.
	assert.Equal(t, status, lines[len(lines)-1], "status bar must stay untouched when overlay is longer than content")
	assert.NotContains(t, lines[len(lines)-1], "TOAST", "status bar must not contain any toast")
}
