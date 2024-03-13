package ui

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	_ list.Item         = (*Option[any])(nil)
	_ list.ItemDelegate = (*itemDelegate[any])(nil)
)

type Option[T comparable] struct {
	Title string
	Value T
}

func (Option[T]) FilterValue() string { return "" }

type itemDelegate[T comparable] struct{}

func (itemDelegate[T]) Height() int                             { return 1 }
func (itemDelegate[T]) Spacing() int                            { return 0 }
func (itemDelegate[T]) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (itemDelegate[T]) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Option[T])
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title)

	fn := styles.ItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return styles.SelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func createList[T comparable](options []Option[T], title string) list.Model {
	listItems := make([]list.Item, 0, len(options))
	var maxItemTitleLength int
	for _, option := range options {
		listItems = append(listItems, option)
		if len(option.Title) > maxItemTitleLength {
			maxItemTitleLength = len(option.Title)
		}
	}

	defaultWidth := max(maxItemTitleLength, len(title)) + 2
	maxHeight := 20
	listHeight := min(len(listItems), maxHeight) + 4

	l := list.New(listItems, itemDelegate[T]{}, defaultWidth, listHeight)
	l.Title = title
	l.InfiniteScrolling = true
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.KeyMap.CursorUp = key.NewBinding(key.WithKeys("up"))
	l.KeyMap.CursorDown = key.NewBinding(key.WithKeys("down"))
	l.Styles.Title = styles.TitleStyle
	l.Styles.PaginationStyle = styles.PaginationStyle
	l.Styles.ActivePaginationDot = styles.ActivePaginationDot
	l.Styles.InactivePaginationDot = styles.InactivePaginationDot
	l.Styles.TitleBar = styles.TitleBarStyle

	return l
}

func max(v ...int) int {
	return slices.Max(v)
}

func min(v ...int) int {
	return slices.Min(v)
}
