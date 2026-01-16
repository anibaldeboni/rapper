package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Option is a generic option for lists
type Option[T comparable] struct {
	Value T
	Title string
}

func (Option[T]) FilterValue() string { return "" }

// FilesView handles CSV file selection
type FilesView struct {
	list   list.Model
	width  int
	height int
}

// Styles for file list
var (
	bullet            = "⦿"
	inactiveDot       = "⦁"
	titleStyle        = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Margin(1, 2)
	itemStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	selectedItemStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("#d6acff"))
	paginationStyle   = lipgloss.NewStyle().MarginLeft(2)
	activeDot         = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#d3d3d3", Dark: "#d3d3d3"}).SetString(bullet).Bold(true)
	inactiveDotStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8d8d8d", Dark: "#8d8d8d"}).SetString(inactiveDot).Bold(true)
)

// NewFilesView creates a new FilesView
func NewFilesView(csvFiles []list.Item) *FilesView {
	l := list.New(csvFiles, fileItemDelegate{}, 60, 0)
	l.InfiniteScrolling = true
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.KeyMap.CursorUp = kbind.Up
	l.KeyMap.CursorDown = kbind.Down
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.ActivePaginationDot = activeDot
	l.Styles.InactivePaginationDot = inactiveDotStyle
	l.Styles.TitleBar = titleStyle.Bold(true)
	l.Title = "👀 Select a CSV file to process"

	return &FilesView{
		list: l,
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
	v.list.SetWidth((width / 4) * 3)
}

// View renders the files view
func (v *FilesView) View() string {
	return v.list.View()
}

// SelectedItem returns the currently selected file
func (v *FilesView) SelectedItem() list.Item {
	return v.list.SelectedItem()
}

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
