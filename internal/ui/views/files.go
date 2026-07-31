package views

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
)

// Option is a generic option for lists
type Option[T comparable] struct {
	Value T
	Title string
}

func (Option[T]) FilterValue() string { return "" }

// FilesView handles CSV file selection. It is a value-type tea.Model
// (Init/Update/View) so AppModel can store it behind a uniform
// `map[View]viewModel` and broadcast messages to every view without
// re-dispatching on the concrete type.
type FilesView struct {
	list   list.Model
	width  int
	height int
}

// Compile-time guard: FilesView must satisfy tea.Model with a value
// receiver. Phase 1 converts the historical pointer-receiver surface
// to value receivers; this assertion fails at build time if the
// conversion regresses.
var _ tea.Model = FilesView{}

// Styles for file list
var (
	titleStyle        = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Margin(1, 2)
	itemStyle         = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("255"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d6acff"))
	paginationStyle   = lipgloss.NewStyle().MarginLeft(2)
)

// NewFilesView creates a new FilesView. The returned value (not a
// pointer) is what AppModel stores in the views map; callers must
// capture the value returned by Update to preserve state.
func NewFilesView(csvFiles []list.Item) FilesView {
	l := list.New(csvFiles, fileItemDelegate{}, 60, 5)
	l.InfiniteScrolling = true
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.KeyMap.CursorUp = kbind.Up
	l.KeyMap.CursorDown = kbind.Down
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.TitleBar = titleStyle.Bold(true)
	l.Title = "👀 Select a CSV file to process"

	v := FilesView{list: l}
	v.list.Styles.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d3d3d3")).
		SetString("⦿").
		Bold(true)
	v.list.Styles.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8d8d8d")).
		SetString("⦁").
		Bold(true)
	return v
}

type filesViewKeyMap struct{}

func (filesViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.Up, kbind.Down, kbind.Select}
}

func (filesViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{kbind.Up, kbind.Down, kbind.Select}}
}

func (FilesView) KeyMap() help.KeyMap {
	return filesViewKeyMap{}
}

// Init returns nil. Viewport size arrives through ViewportSizeMsg.
func (v FilesView) Init() tea.Cmd { return nil }

// Update handles messages for the files view. The view returns a
// value-receiver copy (FilesView) plus an optional command. Callers
// MUST capture the returned value to preserve state — the value
// receiver means the original struct is never mutated.
//
// Recognised messages:
//   - msgs.ViewportSizeMsg: resize the underlying list.
//   - tea.KeyPressMsg on kbind.Select: emit msgs.ItemSelectedMsg
//     carrying the focused file's path.
//
// The return type is (tea.Model, tea.Cmd) to satisfy the tea.Model
// interface; the concrete value is always a FilesView, so callers in
// the same package can type-assert without risk of panic.
func (v FilesView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case msgs.ViewActivatedMsg:
		return v, v.emitStatusKeymap()

	case msgs.ViewportSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.list.SetSize((msg.Width/4)*3, msg.Height)
		return v, nil

	case tea.KeyPressMsg:
		// Only intercept Select; let other keys fall through to the
		// list. The list handles Up/Down/PgUp/PgDn etc. on its own.
		if key.Matches(msg, kbind.Select) {
			// Capture the value BEFORE the list Update so the
			// selection is stable (list.Update may move focus).
			selected, ok := v.list.SelectedItem().(Option[string])
			if !ok {
				return v, nil
			}
			filePath := selected.Value
			// Let the list observe the keypress for visual feedback;
			// the list may produce its own cmd (e.g. a spinner tick)
			// which we compose with the ItemSelectedMsg.
			var listCmd tea.Cmd
			v.list, listCmd = v.list.Update(msg)
			return v, tea.Batch(listCmd, emitItemSelected(filePath))
		}
	}

	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v FilesView) emitStatusKeymap() tea.Cmd {
	return func() tea.Msg {
		return msgs.StatusKeymapMsg{Current: v.KeyMap()}
	}
}

// emitItemSelected returns a tea.Cmd that yields msgs.ItemSelectedMsg
// when run. Extracted so Update stays readable.
func emitItemSelected(filePath string) tea.Cmd {
	return func() tea.Msg {
		return msgs.ItemSelectedMsg{FilePath: filePath}
	}
}

// View renders the files view as a tea.View whose Content holds the
// bubbles list output.
func (v FilesView) View() tea.View {
	return tea.NewView(v.list.View())
}

// ListWidth returns the current width of the embedded list. Exposed
// for test assertions and AppModel state inspection.
func (v FilesView) ListWidth() int { return v.list.Width() }

// ListHeight returns the current height of the embedded list. Exposed
// for test assertions and AppModel state inspection.
func (v FilesView) ListHeight() int { return v.list.Height() }

// fileItemDelegate is the delegate for rendering file list items
type fileItemDelegate struct{}

func (d fileItemDelegate) Height() int                             { return 1 }
func (d fileItemDelegate) Spacing() int                            { return 0 }
func (d fileItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d fileItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Option[string])
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("▶ " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
