package ui

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui/components"
	"github.com/anibaldeboni/rapper/internal/ui/views"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const AppName = "rapper"

var AppVersion = getVersion()

// getVersion extracts version information from build metadata.
// It tries to use the module version first (works with tagged releases),
// then falls back to VCS revision for dev builds.
func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	// Try module version first (e.g., "v2.6.0" or "(devel)")
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return normalizeVersion(info.Main.Version)
	}

	// Fall back to VCS revision for dev builds
	var revision string
	var modified bool
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}

	if revision != "" {
		// Use short commit hash (7 chars) + "-dev" suffix
		if len(revision) > 7 {
			revision = revision[:7]
		}
		if modified {
			return revision + "-dev-dirty"
		}
		return revision + "-dev"
	}

	return "dev"
}

// normalizeVersion removes the "v" prefix from version strings.
// E.g., "v2.6.0" -> "2.6.0"
func normalizeVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

// AppModel is the new multi-view model
type AppModel struct {
	// Navigation
	nav *Navigation

	// State
	logger    logs.Logger
	processor processor.Processor
	configMgr config.Manager

	// Views
	filesView    *views.FilesView
	logsView     *views.LogsView
	settingsView *views.SettingsView
	workersView  *views.WorkersView

	// Common UI elements
	help     help.Model
	spinner  spinner.Model
	toastMgr *components.ToastManager
	width    int
	height   int

	// Context for cancellation
	cancel   context.CancelFunc
	cancelMu *sync.RWMutex
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
		logger:       log,
		processor:    fileProcessor,
		configMgr:    configMgr,
		filesView:    views.NewFilesView(items),
		logsView:     views.NewLogsView(log),
		settingsView: views.NewSettingsView(configMgr),
		workersView:  views.NewWorkersView(fileProcessor),
		help:         createHelp(),
		spinner:      createSpinner(),
		toastMgr:     components.NewToastManager(),
		cancelMu:     &sync.RWMutex{},
	}
}

func createSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}

func createHelp() help.Model {
	help := help.New()

	help.Styles.ShortKey = styles.HelpKeyStyle
	help.Styles.ShortDesc = styles.HelpDescStyle
	help.Styles.ShortSeparator = styles.HelpSepStyle
	help.Styles.Ellipsis = styles.HelpSepStyle
	help.Styles.FullKey = styles.HelpKeyStyle
	help.Styles.FullDesc = styles.HelpDescStyle
	help.Styles.FullSeparator = styles.HelpSepStyle

	return help
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tickCmd(), m.spinner.Tick)
}

func operationError() logs.Message {
	return logs.NewMessage().
		WithIcon(IconInformation).
		WithMessage("Please wait the current operation to finish or cancel pressing ESC")
}
