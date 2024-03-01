package cli

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/anibaldeboni/rapper/cli/ui"

	"github.com/charmbracelet/bubbles/key"
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

func largestOptionTitleLength[T comparable](options []Option[T]) int {
	var max int
	for _, option := range options {
		if len(option.Title) > max {
			max = len(option.Title)
		}
	}
	return max
}

func createList[T comparable](options []Option[T]) list.Model {
	listItems := make([]list.Item, 0, len(options))

	for _, option := range options {
		listItems = append(listItems, option)
	}

	defaultWidth := largestOptionTitleLength(options) + 3
	maxHeight := 20
	listHeight := min(len(listItems), maxHeight) + 4

	l := list.New(listItems, itemDelegate[T]{}, defaultWidth, listHeight)
	l.Title = "Choose a file to process"
	l.InfiniteScrolling = true
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.KeyMap.CursorUp = key.NewBinding(key.WithKeys("up"))
	l.KeyMap.CursorDown = key.NewBinding(key.WithKeys("down"))
	l.Styles.Title = ui.TitleStyle
	l.Styles.PaginationStyle = ui.PaginationStyle
	l.Styles.TitleBar = ui.TitleBarStyle

	return l
}

func min(v ...int) int {
	return slices.Min(v)
}
