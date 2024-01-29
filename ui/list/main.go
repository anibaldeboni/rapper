package list

import (
	"fmt"
	"io"
	"rapper/ui"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const listHeight = 14

var (
	titleStyle        = ui.TitleStyle
	itemStyle         = ui.ItemStyle
	selectedItemStyle = ui.SelectedItemStyle
	paginationStyle   = ui.PaginationStyle
	helpStyle         = ui.HelpStyle
	quitTextStyle     = ui.QuitTextStyle
)

type Option struct {
	Title string
	Value string
}

func (i Option) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Option)
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

type model struct {
	list     list.Model
	Choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			i, ok := m.list.SelectedItem().(Option)
			if ok {
				m.Choice = i.Value
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.Choice != "" {
		return m.Choice
	}
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

func Ask(options []Option, title string) string {
	p := tea.NewProgram(build(options, title))
	m, err := p.Run()
	if err != nil {
		return ""
	}
	return m.(model).Choice
}

func build(options []Option, title string) model {
	listItems := make([]list.Item, 0)

	for _, option := range options {
		listItems = append(listItems, option)
	}

	const defaultWidth = 20

	l := list.New(listItems, itemDelegate{}, defaultWidth, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return model{list: l}
}
