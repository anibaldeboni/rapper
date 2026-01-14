package views

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/anibaldeboni/rapper/internal/ui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// FilesView handles CSV file selection
type FilesView struct {
	list      list.Model
	processor processor.Processor
	logger    logs.Logger
	cancel    context.CancelFunc
	width     int
	height    int
}

// NewFilesView creates a new FilesView
func NewFilesView(csvFiles []list.Item, processor processor.Processor, logger logs.Logger) *FilesView {
	l := createFileList(csvFiles, "Choose a file")

	return &FilesView{
		list:      l,
		processor: processor,
		logger:    logger,
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
	return v.list.View()
}

// SelectedItem returns the currently selected file
func (v *FilesView) SelectedItem() list.Item {
	return v.list.SelectedItem()
}

// StartProcessing starts processing the selected file
func (v *FilesView) StartProcessing(ctx context.Context, filePath string) (context.Context, context.CancelFunc) {
	return v.processor.Do(ctx, filePath)
}

// SetCancel sets the cancel function for the current operation
func (v *FilesView) SetCancel(cancel context.CancelFunc) {
	v.cancel = cancel
}

// Cancel cancels the current operation
func (v *FilesView) Cancel() {
	if v.cancel != nil {
		v.cancel()
	}
}

// fileItemDelegate is the delegate for rendering file list items
type fileItemDelegate struct{}

func (d fileItemDelegate) Height() int                             { return 1 }
func (d fileItemDelegate) Spacing() int                            { return 0 }
func (d fileItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d fileItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ui.Option[string])
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title)

	fn := ui.ItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return ui.SelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// createFileList creates a list model for file selection
func createFileList(items []list.Item, title string) list.Model {
	l := list.New(items, fileItemDelegate{}, 0, 0)
	l.Title = title
	l.InfiniteScrolling = true
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.KeyMap.CursorUp = key.NewBinding(key.WithKeys("up"))
	l.KeyMap.CursorDown = key.NewBinding(key.WithKeys("down"))
	l.Styles.Title = ui.TitleStyle
	l.Styles.PaginationStyle = ui.PaginationStyle
	l.Styles.ActivePaginationDot = ui.ActivePaginationDot
	l.Styles.InactivePaginationDot = ui.InactivePaginationDot
	l.Styles.TitleBar = ui.TitleBarStyle

	return l
}
