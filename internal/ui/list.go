package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	_ list.Item         = (*Option[any])(nil)
	_ list.ItemDelegate = (*itemDelegate[any])(nil)
)

type Option[T comparable] struct {
	Value T
	Title string
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
