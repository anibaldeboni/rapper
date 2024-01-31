package spinner

import (
	"fmt"
	"rapper/ui"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	spinnerStyle = ui.SpinnerStyle
	helpStyle    = ui.SpinnerHelpStyle
	appStyle     = ui.AppStyle
)

type updateLabel string
type errorMsg string
type doneMsg string

type model struct {
	spinner  spinner.Model
	label    string
	quitting bool
	errors   []errorMsg
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
	case updateLabel:
		m.label = string(msg)
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case errorMsg:
		m.errors = append(m.errors, msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case doneMsg:
		m.label = string(msg)
		m.quitting = true
		return m, tea.Quit
	default:
		return m, nil
	}
}

type Spinner interface {
	Run() (tea.Model, error)
	Update(string)
	Error(string)
	Done(string)
}
type SpinnerImpl struct {
	program *tea.Program
}

func New() Spinner {
	return &SpinnerImpl{
		program: tea.NewProgram(newModel()),
	}
}
func (s *SpinnerImpl) Run() (tea.Model, error) {
	return s.program.Run()
}
func (s *SpinnerImpl) Update(label string) {
	s.program.Send(updateLabel(label))
}
func (s *SpinnerImpl) Error(err string) {
	s.program.Send(errorMsg(err))
}
func (s *SpinnerImpl) Done(done string) {
	s.program.Send(doneMsg(done))
}

func (m model) View() string {
	var help string
	var spinner string
	var errors string
	var label = m.label + "\n\n"
	if !m.quitting {
		spinner = m.spinner.View() + " "
		help = helpStyle.Render(
			fmt.Sprintf("Press %s or %s to exit", ui.Bold("q"), ui.Bold("ctrl+c")),
		)
	}

	if len(m.errors) > 0 {
		for _, e := range m.errors {
			errors += string(e) + "\n"
		}
	}

	if m.quitting {
		errors += "\n"
	}

	return appStyle.Render(spinner + label + errors + help)
}
