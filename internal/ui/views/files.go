package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Option is a generic option for lists
type Option[T comparable] struct {
	Value T
	Title string
}

var fileTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1).Bold(true).Render

func (Option[T]) FilterValue() string { return "" }

// FilesView handles CSV file selection
type FilesView struct {
	list   list.Model
	width  int
	height int
	title  string
}

// NewFilesView creates a new FilesView
func NewFilesView(csvFiles []list.Item) *FilesView {
	l := createFileList(csvFiles)

	return &FilesView{
		list:  l,
		title: "👀 Select a CSV file to process",
	}
}

// Update handles messages for the files view
func (v *FilesView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return cmd
}

// Resize updates the view dimensions
func (v *FilesView) Resize(width, height int) {
	v.width = width
	v.height = height
	v.list.SetHeight(height - 5)
}

// View renders the files view
func (v *FilesView) View() string {
	return lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				fileTitleStyle(v.title),
				v.list.View(),
			),
		)
}

// SelectedItem returns the currently selected file
func (v *FilesView) SelectedItem() list.Item {
	return v.list.SelectedItem()
}

// Styles for file list
var (
	bullet            = "⦿"
	inactiveDot       = "⦁"
	titleStyle        = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("255"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d6acff"))
	paginationStyle   = lipgloss.NewStyle().PaddingLeft(2)
	activeDot         = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#d3d3d3", Dark: "#d3d3d3"}).SetString(bullet).Bold(true)
	inactiveDotStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8d8d8d", Dark: "#8d8d8d"}).SetString(inactiveDot).Bold(true)
	titleBarStyle     = lipgloss.NewStyle().PaddingBottom(1)
)

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
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// createFileList creates a list model for file selection
func createFileList(items []list.Item) list.Model {
	l := list.New(items, fileItemDelegate{}, 0, 0)
	l.InfiniteScrolling = true
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.KeyMap.CursorUp = key.NewBinding(key.WithKeys("up"))
	l.KeyMap.CursorDown = key.NewBinding(key.WithKeys("down"))
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.ActivePaginationDot = activeDot
	l.Styles.InactivePaginationDot = inactiveDotStyle
	l.Styles.TitleBar = titleBarStyle

	return l
}
