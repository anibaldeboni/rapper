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

const (
	Error  string = "error"
	Done   string = "done"
	Update string = "update"
)

type UpdateUI struct {
	Type    string
	Message string
}

type model struct {
	spinner  spinner.Model
	label    string
	quitting bool
	errors   []string
}

type Spinner interface {
	Run() (tea.Model, error)
	Update(UpdateUI)
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
func (s *SpinnerImpl) Update(u UpdateUI) {
	s.program.Send(u)
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

	case UpdateUI:
		return m.handleUIUpdate(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	default:
		return m, nil
	}

	return m, nil
}

func (m model) handleUIUpdate(msg UpdateUI) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case Error:
		m.errors = append(m.errors, msg.Message)
		return m, nil
	case Update:
		m.label = msg.Message
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case Done:
		m.label = msg.Message
		m.quitting = true
		return m, tea.Quit
	default:
		return m, nil
	}
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
			errors += e + "\n"
		}
	}

	if m.quitting {
		errors += "\n"
	}

	return appStyle.Render(spinner + label + errors + help)
}
