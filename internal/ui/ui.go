package ui

import (
	"context"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/anibaldeboni/rapper/internal/styles"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	AppName       = "rapper"
	AppVersion    = "2.6.0"
	viewPortTitle = styles.TitleStyle.Render("Execution logs")
	logger        logs.Logger
	cancel        context.CancelFunc
	state         = &State{}
	csvProcessor  processor.Processor
	_             tea.Model = (*Model)(nil)
)

type Model struct {
	viewport  viewport.Model
	filesList list.Model
	help      help.Model
	spinner   spinner.Model
	width     int
}

func New(csvFiles []string, fileProcessor processor.Processor, log logs.Logger) *Model {
	state.Set(SelectFile)
	csvProcessor = fileProcessor
	logger = log

	return &Model{
		viewport:  viewport.New(20, 60),
		filesList: createList(mapListOptions(csvFiles), "Choose a file"),
		help:      createHelp(),
		spinner:   createSpinner(),
	}
}

func createSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	return s
}

func operationError() logs.Message {
	return logs.NewMessage().
		WithIcon(styles.IconInformation).
		WithMessage("Please wait the current operation to finish or cancel pressing ESC")
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tickCmd(), m.spinner.Tick)
}

func (m Model) selectItem(item Option[string]) Model {
	var ctx context.Context

	if state.Get() != Running {
		state.Set(Running)
		ctx, cancel = csvProcessor.Do(context.Background(), item.Value)
		go func(ctx context.Context) {
			<-ctx.Done()
			state.Set(Stale)
		}(ctx)
	} else {
		logger.Add(operationError())
	}
	return m
}

func (m Model) resizeElements(width int, height int) Model {
	m.width = width
	headerHeight := lipgloss.Height(viewPortTitle)
	m.filesList.SetHeight(height - headerHeight - 3)

	logViewWidth := width - lipgloss.Width(m.filesList.View())
	m.viewport.Height = height - headerHeight - 6
	m.viewport.Width = logViewWidth - 2

	return m
}
