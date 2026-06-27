package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
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

// TestSpliceRight_OverlayAnchoredToRightEdgeNotBaseLineWidth reproduces the
// visual bug where the middle line of a multi-line toast appeared on the
// LEFT while the top and bottom lines were on the RIGHT. The cause:
// spliceRight compared the overlay width against the BASE line width and,
// when the base line was shorter than the overlay, replaced the base line
// entirely — leaving the overlay at column 0 instead of the right edge.
//
// The fix must anchor the overlay to a fixed x-position computed from the
// terminal width (the `width` parameter), independent of the visible
// width of each base line. All overlay lines must end at the same column
// from the right, regardless of how short the underlying base line is.
func TestSpliceRight_OverlayAnchoredToRightEdgeNotBaseLineWidth(t *testing.T) {
	// Build a base where the middle content line is intentionally SHORT
	// (only 20 visible cols) while the top/bottom content lines are full
	// width (80 visible cols). The header and status bar are full width.
	header := strings.Repeat("H", 80)
	shortMiddle := "SHORT MIDDLE LINE"
	fullWidthBottom := strings.Repeat("B", 80)
	status := strings.Repeat("S", 80)

	base := strings.Join([]string{header, fullWidthBottom, shortMiddle, fullWidthBottom, status}, "\n")

	// Overlay lines are 40 visible cols each. With a 120-col terminal and
	// a 2-col right margin, the expected targetX is 120 - 40 - 2 = 78.
	// All overlay lines must end at column 117 (right edge - 2 margin).
	overlay := strings.Join([]string{
		strings.Repeat("X", 40),
		strings.Repeat("Y", 40),
		strings.Repeat("Z", 40),
	}, "\n")

	terminalWidth := 120
	got := spliceRight(base, overlay, terminalWidth)
	lines := strings.Split(got, "\n")

	// The overlay lands on lines 1, 2, 3 (between header at 0 and
	// status at len-1). Inspect the visible width of each spliced line
	// to confirm the overlay was placed at the same x-position from
	// the right, regardless of base line width.
	for i := 1; i <= 3; i++ {
		gotWidth := lipgloss.Width(lines[i])
		// Each spliced line must contain exactly the base-line prefix
		// (truncated to targetX) plus the 40-col overlay.
		// targetX = 120 - 40 - 2 = 78 visible columns of base kept.
		expectedMinWidth := 78 + 40 // = 118
		assert.GreaterOrEqual(t, gotWidth, expectedMinWidth,
			"line %d must contain the full overlay (40 cols) starting at x=78, not collapsed to the left; got visible width %d",
			i, gotWidth)
	}

	// The middle line (index 2) is the one that USED to break: the base
	// was only 17 visible cols, so the old code replaced it with the
	// overlay entirely, putting the overlay at column 0 (visible width
	// = 40). The fix must ensure the middle line is at least 118 visible
	// cols wide — proving the overlay is anchored to the right edge.
	middleWidth := lipgloss.Width(lines[2])
	assert.GreaterOrEqual(t, middleWidth, 118,
		"middle line (short base) must have overlay anchored to right edge; got width %d (this is the bug — overlay was at column 0)",
		middleWidth)

	// The overlay must be present in the middle line at all (regression
	// guard against the overlay being dropped).
	assert.Contains(t, lines[2], strings.Repeat("Y", 40),
		"middle line must contain the second overlay line")
}

// TestSpliceRight_WidthParameterControlsRightEdge verifies the
// triangulation: the `width` parameter actually drives the right-edge
// anchor. Two different widths must place the overlay at two different
// x-positions from the right, proving the fix is not relying on a
// hardcoded column.
func TestSpliceRight_WidthParameterControlsRightEdge(t *testing.T) {
	base := strings.Join([]string{
		strings.Repeat("H", 100), // header
		strings.Repeat("B", 100), // content 1
		strings.Repeat("B", 100), // content 2
		strings.Repeat("S", 100), // status
	}, "\n")
	overlay := strings.Repeat("X", 20) // 20-col overlay

	// Wide terminal: overlay should be far from the left.
	wideResult := spliceRight(base, overlay, 200)
	wideLines := strings.Split(wideResult, "\n")
	wideLine := wideLines[1] // first content line

	// Narrow terminal: overlay should be closer to the left.
	narrowResult := spliceRight(base, overlay, 80)
	narrowLines := strings.Split(narrowResult, "\n")
	narrowLine := narrowLines[1]

	// In the wide case, the first 'X' must appear later (further right)
	// than in the narrow case.
	wideXPos := strings.Index(wideLine, "X")
	narrowXPos := strings.Index(narrowLine, "X")
	assert.Greater(t, wideXPos, narrowXPos,
		"wider terminal must place overlay further from the left: wide=%d narrow=%d",
		wideXPos, narrowXPos)

	// The overlay must still be anchored: in a 200-col terminal with a
	// 20-col overlay and 2-col margin, targetX = 200-20-2 = 178. The
	// 100-col base is padded to 178, so the first X is at column 178.
	assert.Equal(t, 178, wideXPos,
		"wide terminal: overlay first X must be at column 178 (200-20-2)")

	// In an 80-col terminal: targetX = 80-20-2 = 58.
	assert.Equal(t, 58, narrowXPos,
		"narrow terminal: overlay first X must be at column 58 (80-20-2)")
}

// TestSpliceRight_OverlayWiderThanTerminalFallsBackToZero verifies the
// edge case: when the overlay is wider than the available terminal
// width, the function places it at column 0 instead of producing a
// negative targetX. This prevents arithmetic underflow and keeps the
// overlay visible.
func TestSpliceRight_OverlayWiderThanTerminalFallsBackToZero(t *testing.T) {
	base := strings.Join([]string{
		strings.Repeat("H", 50),
		strings.Repeat("B", 50),
		strings.Repeat("S", 50),
	}, "\n")
	overlay := strings.Repeat("X", 60) // wider than terminal (40)

	got := spliceRight(base, overlay, 40) // terminal narrower than overlay
	lines := strings.Split(got, "\n")

	// Overlay must land on the content line (line 1) and start at x=0
	// (the fallback when targetX would be negative).
	assert.Contains(t, lines[1], strings.Repeat("X", 60),
		"overlay must be present in the content line even when it is wider than the terminal")

	// Header and status must still be untouched.
	assert.Equal(t, strings.Repeat("H", 50), lines[0])
	assert.Equal(t, strings.Repeat("S", 50), lines[len(lines)-1])
}
