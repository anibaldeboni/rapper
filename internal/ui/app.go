package ui

import (
	"context"

	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/anibaldeboni/rapper/internal/ui/views"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	AppName    = "rapper"
	AppVersion = "2.6.0"
)

// AppModel is the new multi-view model
type AppModel struct {
	// Navigation
	nav *Navigation

	// State
	state     *State
	logger    logs.Logger
	processor processor.Processor
	configMgr config.Manager

	// Views
	filesView    *views.FilesView
	logsView     *views.LogsView
	settingsView *views.SettingsView
	workersView  *views.WorkersView

	// Common UI elements
	help    help.Model
	spinner spinner.Model
	width   int
	height  int

	// Context for cancellation
	cancel context.CancelFunc
}

// NewApp creates a new AppModel with multi-view support
func NewApp(csvFiles []string, fileProcessor processor.Processor, log logs.Logger, configMgr config.Manager) *AppModel {
	// Convert CSV files to list items
	items := make([]list.Item, 0, len(csvFiles))
	for _, file := range csvFiles {
		items = append(items, mapFileToOption(file))
	}

	return &AppModel{
		nav:          NewNavigation(),
		state:        &State{},
		logger:       log,
		processor:    fileProcessor,
		configMgr:    configMgr,
		filesView:    views.NewFilesView(items, fileProcessor, log),
		logsView:     views.NewLogsView(log),
		settingsView: views.NewSettingsView(configMgr),
		workersView:  views.NewWorkersView(fileProcessor),
		help:         createHelp(),
		spinner:      createSpinner(),
	}
}

func createSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tickCmd(), m.spinner.Tick)
}

func operationError() logs.Message {
	return logs.NewMessage().
		WithIcon(IconInformation).
		WithMessage("Please wait the current operation to finish or cancel pressing ESC")
}
