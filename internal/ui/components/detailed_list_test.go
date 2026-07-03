package components_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/ui/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubRenderer is the test renderer for DetailedList[string]. The
// behaviour is configurable per item via the Title/Detail fields —
// tests can flip Detail to "" to exercise the Enter-no-op branch.
type stubRenderer struct {
	titleOf  func(string) string
	detailOf func(string) string
}

func (r stubRenderer) Title(s string) string {
	if r.titleOf == nil {
		return s
	}
	return r.titleOf(s)
}

func (r stubRenderer) Detail(s string) string {
	if r.detailOf == nil {
		return "detail-of:" + s
	}
	return r.detailOf(s)
}

func (r stubRenderer) Style(s string) lipgloss.Style {
	return lipgloss.NewStyle()
}

func (r stubRenderer) SelectedStyle(s string) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

// newStubList builds a DetailedList with the supplied items. The
// initial autoScroll=true state is overridden to false so the
// cursor lands at index 0 after Append — most tests want a known
// starting cursor to assert navigation, not the autoScroll
// behaviour.
func newStubList(items []string) components.DetailedList[string] {
	l := components.NewDetailedList[string](stubRenderer{}).Append(items)
	l = l.WithAutoScroll(false)
	return l
}

func press(k rune, l components.DetailedList[string]) components.DetailedList[string] {
	next, _ := l.Update(tea.KeyPressMsg{Code: k})
	return next.(components.DetailedList[string])
}

// TestDetailedList_Down_MovesCursor advances the cursor one row and
// flips autoScroll off. This is the core navigation invariant: every
// directional key must set autoScroll=false so a running log does
// not snap back to the bottom under the user's feet.
func TestDetailedList_Down_MovesCursor(t *testing.T) {
	l := newStubList([]string{"a", "b", "c"})
	require.Equal(t, 0, l.Cursor(), "cursor starts at 0")

	l = press(tea.KeyDown, l)

	assert.Equal(t, 1, l.Cursor(), "Down must advance the cursor by 1")
	assert.False(t, l.AutoScroll(), "any directional key clears autoScroll")
}

// TestDetailedList_Up_ClampsAtZero — Up at cursor=0 is a no-op (the
// cursor does not wrap or go negative). Without the clamp the
// subsequent View() call would index items[-1].
func TestDetailedList_Up_ClampsAtZero(t *testing.T) {
	l := newStubList([]string{"a", "b"})

	l = press(tea.KeyUp, l)

	assert.Equal(t, 0, l.Cursor(), "Up at cursor=0 must stay at 0")
}

// TestDetailedList_Down_ClampsAtLast — Down at the last item does
// not advance further. This is the "bottom of the list" case.
func TestDetailedList_Down_ClampsAtLast(t *testing.T) {
	l := newStubList([]string{"a", "b"})
	l = press(tea.KeyDown, l) // cursor=1
	l = press(tea.KeyDown, l) // would overflow

	assert.Equal(t, 1, l.Cursor(), "Down at the last item must stay at the last index")
}

// TestDetailedList_PageDown_JumpsByPageSize — PgDn moves the cursor
// by the configured pageSize, not by 1.
func TestDetailedList_PageDown_JumpsByPageSize(t *testing.T) {
	l := components.NewDetailedList[string](stubRenderer{}).
		WithPageSize(2).
		Append([]string{"a", "b", "c", "d", "e"})
	l = l.WithAutoScroll(false)

	l = press(tea.KeyPgDown, l)

	assert.Equal(t, 2, l.Cursor(), "PgDn from 0 with pageSize=2 must land on 2")
	assert.False(t, l.AutoScroll(), "PgDn clears autoScroll")
}

// TestDetailedList_PageUp_ClampsAtZero — PgUp at cursor=0 stays at 0.
func TestDetailedList_PageUp_ClampsAtZero(t *testing.T) {
	l := components.NewDetailedList[string](stubRenderer{}).
		WithPageSize(2).
		Append([]string{"a", "b", "c", "d"})
	l = l.WithAutoScroll(false)

	l = press(tea.KeyPgUp, l)

	assert.Equal(t, 0, l.Cursor(), "PgUp at 0 must stay at 0")
}

// TestDetailedList_Home_JumpsToFirst — Home is the explicit "jump to
// the top" key; cursor=0, autoScroll=false (we are now at the head).
func TestDetailedList_Home_JumpsToFirst(t *testing.T) {
	l := newStubList([]string{"a", "b", "c"})
	l = press(tea.KeyDown, l) // cursor=1

	l = press(tea.KeyHome, l)

	assert.Equal(t, 0, l.Cursor(), "Home must move the cursor to 0")
	assert.False(t, l.AutoScroll(), "Home clears autoScroll")
}

// TestDetailedList_End_JumpsToLastAndEnablesAutoScroll — End is the
// "go to tail" key; the cursor lands on the last item and autoScroll
// re-enables. This matches the legacy viewport behavior where
// GotoBottom also re-armed auto-scroll.
func TestDetailedList_End_JumpsToLastAndEnablesAutoScroll(t *testing.T) {
	l := newStubList([]string{"a", "b", "c", "d"})

	l = press(tea.KeyEnd, l)

	assert.Equal(t, 3, l.Cursor(), "End must move the cursor to the last index")
	assert.True(t, l.AutoScroll(), "End must re-enable autoScroll")
}

// TestDetailedList_Enter_TogglesExpand — Enter expands the row, a
// second Enter collapses it.
func TestDetailedList_Enter_TogglesExpand(t *testing.T) {
	l := newStubList([]string{"a", "b", "c"})

	l = press(tea.KeyEnter, l)
	assert.Equal(t, 0, l.Expanded(), "first Enter expands cursor row")

	l = press(tea.KeyEnter, l)
	assert.Equal(t, -1, l.Expanded(), "second Enter collapses back to -1")
}

// TestDetailedList_Enter_NoOpOnEmptyDetail — if the renderer returns
// an empty Detail for the cursor row, Enter is a silent no-op (the
// expand index stays at -1 and the View does not include extra
// detail text).
func TestDetailedList_Enter_NoOpOnEmptyDetail(t *testing.T) {
	l := components.NewDetailedList[string](stubRenderer{
		detailOf: func(s string) string { return "" },
	}).Append([]string{"no-detail-row"})
	l = l.WithAutoScroll(false)

	l = press(tea.KeyEnter, l)

	assert.Equal(t, -1, l.Expanded(), "Enter on an item with empty Detail must be a no-op")
}

// TestDetailedList_Reset_ClearsItemsAndEnablesAutoScroll — Reset is
// called by LogsView on ProcessingStartedMsg so each run starts from
// a clean slate with the cursor at the tail.
func TestDetailedList_Reset_ClearsItemsAndEnablesAutoScroll(t *testing.T) {
	l := newStubList([]string{"a", "b"})
	l = press(tea.KeyUp, l) // autoScroll=false

	l = l.Reset()

	assert.Equal(t, 0, l.Cursor(), "Reset must put the cursor at 0")
	assert.True(t, l.AutoScroll(), "Reset must re-enable autoScroll")
	assert.Equal(t, 0, l.Len(), "Reset must drop all items")
}

// TestDetailedList_Append_GrowsItems — Append adds the supplied items
// to the buffer; on the first call the buffer goes from empty to
// non-empty.
func TestDetailedList_Append_GrowsItems(t *testing.T) {
	l := components.NewDetailedList[string](stubRenderer{}).WithAutoScroll(false)
	require.Equal(t, 0, l.Len())

	l = l.Append([]string{"a", "b"})

	assert.Equal(t, 2, l.Len())
}

// TestDetailedList_View_RendersAllItems — the rendered output
// contains every title (one per item). The implementation must not
// silently drop rows.
func TestDetailedList_View_RendersAllItems(t *testing.T) {
	l := newStubList([]string{"alpha", "beta", "gamma"})

	out := l.View().Content

	for _, want := range []string{"alpha", "beta", "gamma"} {
		assert.True(t, strings.Contains(out, want), "View must render %q; got: %q", want, out)
	}
}

// TestDetailedList_View_RendersSelectedItemBold — the cursor row
// gets the SelectedStyle (bold) so the user can see which entry is
// focused. The check is loose because lipgloss may emit ANSI codes
// that vary across terminals.
func TestDetailedList_View_RendersSelectedItemBold(t *testing.T) {
	l := newStubList([]string{"a", "b"})

	out := l.View().Content
	lines := strings.Split(out, "\n")
	require.GreaterOrEqual(t, len(lines), 2)

	assert.NotEqual(t, lines[0], lines[1],
		"selected row must differ in style from unselected; got identical lines")
}

// TestDetailedList_View_ExpandedRow_AppendsDetail — when the cursor
// is on a row with non-empty Detail, the rendered output includes
// the detail text on the line below the title. This is the
// "expandable" behaviour the spec calls out.
func TestDetailedList_View_ExpandedRow_AppendsDetail(t *testing.T) {
	l := newStubList([]string{"a", "b"})

	l = press(tea.KeyEnter, l) // expand row 0 — cursor is at 0 (autoScroll off)

	out := l.View().Content

	assert.Contains(t, out, "detail-of:a", "expanded row must render its Detail() result")
}

// TestDetailedList_HandlesUnrelatedKeys_NoOp — keys that are not
// navigation keys (e.g. arbitrary letters) are no-ops. This guards
// against accidentally swallowing global keypresses.
func TestDetailedList_HandlesUnrelatedKeys_NoOp(t *testing.T) {
	l := newStubList([]string{"a", "b"})

	next, cmd := l.Update(tea.KeyPressMsg{Code: 'x'})

	assert.Nil(t, cmd, "unknown keys must not produce a command")
	l2 := next.(components.DetailedList[string])
	assert.Equal(t, 0, l2.Cursor(), "unknown key must not move the cursor")
}

// TestDetailedList_Init_ReturnsNil — Init is a no-op cmd.
func TestDetailedList_Init_ReturnsNil(t *testing.T) {
	l := components.NewDetailedList[string](stubRenderer{})
	assert.Nil(t, l.Init())
}
