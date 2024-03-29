package ui

import (
	"context"

	"github.com/anibaldeboni/rapper/internal/execlog"
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
	AppVersion    = "2.5.2"
	viewPortTitle = styles.TitleStyle.Render("Execution logs")
	logs          execlog.Manager
	ctx           context.Context
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

func New(csvFiles []string, csvProc processor.Processor, logManager execlog.Manager) *Model {
	state.Set(SelectFile)
	csvProcessor = csvProc
	logs = logManager

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

func stop() {
	cancel()
	state.Set(Stale)
}

func operationError() execlog.Message {
	return execlog.NewMessage().
		WithIcon(styles.IconInformation).
		WithMessage("Please wait the current operation to finish or cancel pressing ESC")
}

func setContext() {
	ctx, cancel = context.WithCancel(context.Background())
	state.Set(Running)
}

func (this Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tickCmd(), this.spinner.Tick)
}

func (this Model) selectItem(item Option[string]) Model {
	if state.Get() != Running {
		setContext()
		csvProcessor.Do(ctx, stop, item.Value)
	} else {
		logs.Add(operationError())
	}
	return this
}

func (this Model) resizeElements(width int, height int) Model {
	this.width = width
	this.filesList.SetHeight(height - 4)

	logViewWidth := width - lipgloss.Width(this.filesList.View())
	headerHeight := lipgloss.Height(viewPortTitle)

	this.viewport.Height = height - headerHeight - 6
	this.viewport.Width = logViewWidth - 2

	return this
}
