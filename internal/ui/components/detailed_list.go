package components

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
)

// ItemRenderer is the adapter contract for DetailedList[T]. The
// generic list knows nothing about the underlying type — every
// presentation decision (row title, detail text, row style) is
// delegated to the renderer so the component can be reused with any
// T without leaking domain types into the components package.
//
// The contract is intentionally small (4 methods) so renderers stay
// easy to implement and test.
type ItemRenderer[T any] interface {
	// Title returns the single-line title for the row. Always shown.
	Title(item T) string
	// Detail returns the multi-line detail text for the row, or ""
	// to signal "this row is not expandable". The component treats
	// empty Detail as a no-op for Enter.
	Detail(item T) string
	// Style returns the style applied to the title of a non-selected
	// row. Implementations can use it to color the row by category.
	Style(item T) lipgloss.Style
	// SelectedStyle returns the style applied to the title of the
	// currently selected row.
	SelectedStyle(item T) lipgloss.Style
}

// defaultPageSize is the fallback page size when the view has not
// been sized yet (or when height / itemHeight is degenerate). It
// matches the original LogsView viewport behavior.
const defaultPageSize = 5

// DetailedList is a vertical, single-cursor, expandable list
// component. It owns the cursor position, the expanded-row index,
// the autoScroll flag, and the page size. The component does not
// import any domain types — every presentation decision lives on
// the ItemRenderer[T] passed at construction time.
//
// All Update methods operate on a value receiver and return a new
// DetailedList[T]; callers MUST capture the return so state
// changes survive across the framework's message dispatch.
type DetailedList[T any] struct {
	items      []T
	renderer   ItemRenderer[T]
	cursor     int
	expanded   int
	width      int
	height     int
	pageSize   int
	autoScroll bool
}

// Compile-time guard: DetailedList must satisfy tea.Model. The
// generic instantiation here is a placeholder; real uses
// (DetailedList[LogMessage], DetailedList[string]) are checked at
// their construction site.
var _ tea.Model = DetailedList[any]{}

// NewDetailedList returns an empty DetailedList wired to the
// supplied renderer. Items are added with Append; the empty
// constructor is the zero-state the framework sees on Init.
func NewDetailedList[T any](renderer ItemRenderer[T]) DetailedList[T] {
	return DetailedList[T]{
		renderer:   renderer,
		expanded:   -1,
		pageSize:   defaultPageSize,
		autoScroll: true,
	}
}

// WithPageSize overrides the page size used by PgUp/PgDn. The
// default is 5; tests use this to force deterministic jumps.
func (l DetailedList[T]) WithPageSize(n int) DetailedList[T] {
	if n > 0 {
		l.pageSize = n
	}
	return l
}

// WithAutoScroll overrides the autoScroll flag. The default is true
// (a fresh list pins the cursor to the tail so new entries are
// visible); tests use this to assert navigation from a known
// cursor position. Disabling autoScroll also parks the cursor at
// index 0 so the caller can start navigating from a deterministic
// position.
func (l DetailedList[T]) WithAutoScroll(enabled bool) DetailedList[T] {
	l.autoScroll = enabled
	if !enabled {
		l.cursor = 0
	}
	return l
}

// Init returns nil. The component does not schedule work; the
// parent view (LogsView) drives refreshes via Append on
// MetricsTickMsg.
func (l DetailedList[T]) Init() tea.Cmd { return nil }

// Append adds new items to the end of the buffer. Called by
// LogsView on every MetricsTickMsg when the in-memory log grew.
//
// When autoScroll is true (the user is at the tail) the cursor
// follows the last item so the next View() lands on the freshly
// added entry. This preserves the legacy "scroll-to-bottom on
// new content" behavior without requiring the parent to thread
// scroll state into the component.
func (l DetailedList[T]) Append(items []T) DetailedList[T] {
	wasEmpty := len(l.items) == 0
	l.items = append(l.items, items...)
	if wasEmpty {
		l.cursor = 0
	}
	if l.autoScroll {
		l.cursor = len(l.items) - 1
	}
	return l
}

// Reset empties the buffer, puts the cursor at 0, and re-enables
// autoScroll. Called by LogsView on ProcessingStartedMsg so each
// run starts from a clean slate.
func (l DetailedList[T]) Reset() DetailedList[T] {
	l.items = nil
	l.cursor = 0
	l.expanded = -1
	l.autoScroll = true
	return l
}

// Cursor returns the current cursor index. Diagnostic accessor for
// tests.
func (l DetailedList[T]) Cursor() int { return l.cursor }

// Expanded returns the index of the currently expanded row, or -1
// if no row is expanded.
func (l DetailedList[T]) Expanded() int { return l.expanded }

// AutoScroll reports whether the cursor is currently pinned to the
// tail. Parent views use this to decide whether to push a new
// MetricsTickMsg-driven Append.
func (l DetailedList[T]) AutoScroll() bool { return l.autoScroll }

// Len returns the number of items currently in the buffer.
func (l DetailedList[T]) Len() int { return len(l.items) }

// Items returns a copy of the current item slice. Useful for tests
// that want to inspect the buffer without poking private fields.
func (l DetailedList[T]) Items() []T {
	out := make([]T, len(l.items))
	copy(out, l.items)
	return out
}

// SetSize updates the rendered width/height. Called by the parent
// view on ViewportSizeMsg. The component does not paginate yet
// (height is not used for visible-row calculation); the field is
// reserved for a future viewport-based scroll implementation.
func (l DetailedList[T]) SetSize(width, height int) DetailedList[T] {
	l.width = width
	l.height = height
	return l
}

// Width returns the width previously set via SetSize. Diagnostic
// accessor used by views/logs_test.go to assert the partition
// invariant.
func (l DetailedList[T]) Width() int { return l.width }

// Height returns the height previously set via SetSize.
func (l DetailedList[T]) Height() int { return l.height }

// Update handles a single tea message. Only key presses are
// recognised; every other message is a no-op and the returned
// command is nil. Recognised keys (from the LogsView keymap):
//
//   - Up:    cursor--, autoScroll=false
//   - Down:  cursor++, autoScroll=false
//   - PgUp:  cursor -= pageSize, autoScroll=false
//   - PgDn:  cursor += pageSize, autoScroll=false
//   - Home:  cursor = 0,    autoScroll=false
//   - End:   cursor = last, autoScroll=true
//   - Enter: toggle expand; no-op if Detail is empty
//
// Left/Right are intentionally NOT handled — DetailedList is
// vertical-only.
func (l DetailedList[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return l, nil
	}

	switch {
	case key.Matches(keyMsg, kbind.Up):
		if l.cursor > 0 {
			l.cursor--
		}
		l.autoScroll = false
	case key.Matches(keyMsg, kbind.Down):
		if l.cursor < len(l.items)-1 {
			l.cursor++
		}
		l.autoScroll = false
	case key.Matches(keyMsg, kbind.PageUp):
		l.cursor -= l.pageSize
		if l.cursor < 0 {
			l.cursor = 0
		}
		l.autoScroll = false
	case key.Matches(keyMsg, kbind.PageDown):
		l.cursor += l.pageSize
		if maxIdx := len(l.items) - 1; maxIdx >= 0 && l.cursor > maxIdx {
			l.cursor = maxIdx
		}
		l.autoScroll = false
	case key.Matches(keyMsg, kbind.GotoTop):
		l.cursor = 0
		l.autoScroll = false
	case key.Matches(keyMsg, kbind.GotoBottom):
		l.cursor = len(l.items) - 1
		if l.cursor < 0 {
			l.cursor = 0
		}
		l.autoScroll = true
	case key.Matches(keyMsg, kbind.Select):
		if len(l.items) == 0 {
			return l, nil
		}
		if l.renderer.Detail(l.items[l.cursor]) == "" {
			return l, nil
		}
		if l.expanded == l.cursor {
			l.expanded = -1
		} else {
			l.expanded = l.cursor
		}
	}

	return l, nil
}

// View renders the list as a tea.View. Each row gets the renderer's
// Style (or SelectedStyle when the cursor is on it). The expanded
// row's Detail text is appended below the title. AutoScroll
// silently moves the cursor to the last item so the freshly added
// entry is visible after every Append.
func (l DetailedList[T]) View() tea.View {
	if l.autoScroll && len(l.items) > 0 {
		l.cursor = len(l.items) - 1
	}

	if len(l.items) == 0 {
		return tea.NewView("")
	}

	rows := make([]string, 0, len(l.items))
	for i, item := range l.items {
		style := l.renderer.Style(item)
		if i == l.cursor {
			style = l.renderer.SelectedStyle(item)
		}
		title := l.renderer.Title(item)
		rows = append(rows, style.Render(title))

		if l.expanded == i {
			detail := l.renderer.Detail(item)
			if detail != "" {
				rows = append(rows, style.Render(detail))
			}
		}
	}
	return tea.NewView(strings.Join(rows, "\n"))
}

// String renders the list as a plain string. Provided for callers
// that want a quick rendering without the tea.View wrapper (used
// by tests that don't care about the tea.View field set).
func (l DetailedList[T]) String() string {
	return l.View().Content
}
