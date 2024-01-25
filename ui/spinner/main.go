package spinner

import (
	"rapper/ui"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	spinnerStyle = ui.SpinnerStyle
	helpStyle    = ui.SpinnerHelpStyle
	dotStyle     = ui.DotStyle
	appStyle     = ui.AppStyle
)

type UpdateLabel string
type Error string
type Done string

type model struct {
	spinner  spinner.Model
	label    string
	quitting bool
	errors   []Error
}

func newModel() model {
	s := spinner.New()
	s.Style = spinnerStyle
	s.Spinner = spinner.Points
	return model{
		spinner: s,
		label:   "Initializing",
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case UpdateLabel:
		m.label = string(msg)
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case Error:
		m.errors = append(m.errors, msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case Done:
		m.label = string(msg)
		m.quitting = true
		return m, tea.Quit
	default:
		return m, nil
	}
}

func New() *tea.Program {
	return tea.NewProgram(newModel())
}

func (m model) View() string {
	var s string
	if !m.quitting {
		s += m.spinner.View() + " "
	}
	s += string(m.label)

	if len(m.errors) > 0 {
		s += "\n\n"
		for _, e := range m.errors {
			s += string(e) + "\n"
		}
	}

	if !m.quitting {
		s += helpStyle.Render("Press q or ctrl+c key to exit")
	}

	if m.quitting {
		s += "\n"
	}

	return appStyle.Render(s)
}
