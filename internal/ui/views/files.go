package views

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/csv"
	"github.com/anibaldeboni/rapper/internal/styles"
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
	list       list.Model
	width      int
	height     int
	separator  rune
	table      table.Model
	previewErr string
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

const (
	previewMaxRows        = 10
	previewPaneMarginTop  = 1
	previewPaneMarginLeft = 1
)

// NewFilesView creates a new FilesView. The returned value (not a
// pointer) is what AppModel stores in the views map; callers must
// capture the value returned by Update to preserve state.
func NewFilesView(csvFiles []list.Item, separator rune) FilesView {
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
	l.Title = "🗂️ Available files"

	v := FilesView{
		list:      l,
		separator: separator,
		table:     newPreviewTable(csv.Preview{}),
	}
	v.list.Styles.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d3d3d3")).
		SetString("⦿").
		Bold(true)
	v.list.Styles.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8d8d8d")).
		SetString("⦁").
		Bold(true)
	v = v.refreshPreview()
	return v
}

func previewTableStyles() table.Styles {
	cell := lipgloss.NewStyle().
		Align(lipgloss.Left)

	return table.Styles{
		Header:   styles.TableHeaderStyle,
		Cell:     cell,
		Selected: cell,
	}
}

func newPreviewTable(preview csv.Preview) table.Model {
	columns := make([]table.Column, len(preview.Headers))
	for i, header := range preview.Headers {
		columns[i] = table.Column{Title: header, Width: 1}
	}

	rows := make([]table.Row, len(preview.Rows))
	for i, row := range preview.Rows {
		rows[i] = table.Row(row)
	}

	model := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithKeyMap(table.KeyMap{}),
	)
	model.SetStyles(previewTableStyles())
	return model
}

type filesViewKeyMap struct{}

func (filesViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.Up, kbind.Down, kbind.RunBatch}
}

func (filesViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{kbind.Up, kbind.Down, kbind.RunBatch}}
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
//   - tea.KeyPressMsg on kbind.RunBatch: emit msgs.ItemSelectedMsg
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
		v.layout()
		return v, nil
	}

	previousIndex := v.list.Index()
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, kbind.RunBatch) {
		selected, ok := v.list.SelectedItem().(Option[string])
		if !ok {
			return v, nil
		}
		filePath := selected.Value
		var listCmd tea.Cmd
		v.list, listCmd = v.list.Update(keyMsg)
		v.refreshOnSelectionChange(previousIndex)
		return v, tea.Batch(listCmd, emitItemSelected(filePath))
	}

	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	v.refreshOnSelectionChange(previousIndex)
	return v, cmd
}

func (v *FilesView) refreshOnSelectionChange(previousIndex int) {
	if previousIndex != v.list.Index() {
		*v = v.refreshPreview()
	}
}

func (v FilesView) refreshPreview() FilesView {
	selected, ok := v.list.SelectedItem().(Option[string])
	if !ok {
		v.table = newPreviewTable(csv.Preview{})
		v.previewErr = ""
		v.layout()
		return v
	}

	preview, err := csv.ReadPreview(selected.Value, v.separator, previewMaxRows)
	if err != nil {
		v.table = newPreviewTable(csv.Preview{})
		v.previewErr = err.Error()
		v.layout()
		return v
	}

	v.table = newPreviewTable(preview)
	v.previewErr = ""
	v.layout()
	return v
}

func (v FilesView) WithSeparator(separator rune) FilesView {
	v.separator = separator
	return v.refreshPreview()
}

func (v *FilesView) layout() {
	if v.width == 0 && v.height == 0 {
		return
	}

	listWidth := max(v.width/4, 0)
	tableWidth := v.tableWidth()
	v.list.SetSize(listWidth, v.height)
	v.table.SetWidth(tableWidth)
	v.table.SetHeight(max(v.height, 0))
	v.setTableColumnWidths(tableWidth)
}

func (v FilesView) tableWidth() int {
	listWidth := max(v.width/4, 0)
	return max(v.width-listWidth-previewPaneMarginLeft, 0)
}

func (v *FilesView) setTableColumnWidths(width int) {
	columns := v.table.Columns()
	if len(columns) == 0 {
		return
	}

	contentWidth := max(width-2*len(columns), len(columns))
	columnWidth := max(contentWidth/len(columns), 1)
	for i := range columns {
		columns[i].Width = columnWidth
	}
	columns[len(columns)-1].Width += max(contentWidth-columnWidth*len(columns), 0)
	v.table.SetColumns(columns)
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
	previewContent := v.table.View()
	if v.previewErr != "" {
		previewContent = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			MaxWidth(v.tableWidth()).
			Render("⚠ " + v.previewErr)
	}
	previewPane := lipgloss.NewStyle().
		Width(v.tableWidth()).
		Height(v.height - 6).
		MarginTop(previewPaneMarginTop).
		MarginLeft(previewPaneMarginLeft).
		Render(lipgloss.JoinVertical(
			lipgloss.Top,
			styles.TitleStyle.MarginBottom(1).Render("📄 Preview"),
			previewContent,
		))
	listPane := ""
	if v.list.Width() > 0 {
		listPane = lipgloss.NewStyle().
			Width(v.list.Width()).
			MaxWidth(v.list.Width()).
			Height(v.height - 6).
			Render(v.list.View())
	}
	return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, listPane, previewPane))
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
