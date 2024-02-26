package cli

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/anibaldeboni/rapper/cli/ui"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Option[T comparable] struct {
	Title string
	Value T
}

func (i Option[T]) FilterValue() string { return "" }

type itemDelegate[T comparable] struct{}

func (d itemDelegate[T]) Height() int                             { return 1 }
func (d itemDelegate[T]) Spacing() int                            { return 0 }
func (d itemDelegate[T]) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate[T]) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Option[T])
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

func createList[T comparable](options []Option[T]) list.Model {
	listItems := make([]list.Item, 0)

	for _, option := range options {
		listItems = append(listItems, option)
	}

	defaultWidth := 20
	maxHeight := 20
	listHeight := min(len(listItems), maxHeight) + 4

	l := list.New(listItems, itemDelegate[T]{}, defaultWidth, listHeight)
	l.Title = "Choose a file to process"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.Styles.Title = ui.TitleStyle
	l.Styles.PaginationStyle = ui.PaginationStyle
	l.Styles.TitleBar = ui.TitleBarStyle

	return l
}

func min(v ...int) int {
	return slices.Min(v)
}
