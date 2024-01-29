package list

import (
	"fmt"
	"io"
	"rapper/ui"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	titleStyle        = ui.TitleStyle
	itemStyle         = ui.ItemStyle
	selectedItemStyle = ui.SelectedItemStyle
	paginationStyle   = ui.PaginationStyle
	helpStyle         = ui.HelpStyle
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

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model[T comparable] struct {
	list     list.Model
	choice   T
	quitting bool
}

func (m model[T]) Init() tea.Cmd {
	return nil
}

func (m model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(Option[T])
			if ok {
				m.quitting = true
				m.choice = i.Value
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model[T]) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

func Ask[T comparable](options []Option[T], title string) (T, error) {
	p := tea.NewProgram(build[T](options, title))
	m, err := p.Run()
	if err != nil {
		return *new(T), err
	}
	model, ok := m.(model[T])
	if !ok {
		return *new(T), fmt.Errorf("could not cast model to list model")
	}
	return model.choice, nil
}

func build[T comparable](options []Option[T], title string) model[T] {
	listItems := make([]list.Item, 0)

	for _, option := range options {
		listItems = append(listItems, option)
	}

	defaultWidth := 20
	listHeight := len(options) + 6

	l := list.New(listItems, itemDelegate[T]{}, defaultWidth, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return model[T]{list: l}
}
