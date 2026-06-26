package ui

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui/components"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	"github.com/anibaldeboni/rapper/internal/ui/views"
)

const AppName = "rapper"

// version can be set via ldflags during build time (e.g., -X 'github.com/anibaldeboni/rapper/internal/ui.version=1.0.0')
var version string

var AppVersion = getVersion()

// getVersion extracts version information from build metadata.
// It first checks if version was set via ldflags (e.g., by goreleaser),
// then tries to use the module version (works with tagged releases),
// and finally falls back to VCS revision for dev builds.
func getVersion() string {
	// If version was set via ldflags (by goreleaser), use it
	if version != "" {
		return normalizeVersion(version)
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	// Extract VCS information
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

	// Check if we have a proper semantic version (not pseudo-version)
	version := info.Main.Version
	if version != "" && version != "(devel)" && !isPseudoVersion(version) {
		// Clean version (e.g., "v2.6.0" -> "2.6.0")
		return normalizeVersion(version)
	}

	// Fall back to VCS revision for dev builds
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

// isPseudoVersion checks if a version string is a Go pseudo-version.
// Pseudo-versions have format like "v0.0.0-20191109021931-daa7c04131f5"
func isPseudoVersion(version string) bool {
	// Pseudo-versions contain a timestamp (e.g., "20191109021931")
	return strings.Contains(version, "-0.") ||
		strings.Contains(version, "+incompatible") ||
		(strings.Count(version, "-") >= 2 && len(version) > 25)
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
	logger    ports.LogService
	processor ports.ProcessorController
	configMgr ports.ConfigManager

	// Views
	filesView    *views.FilesView
	logsView     *views.LogsView
	settingsView *views.SettingsView
	workersView  *views.WorkersView

	// Common UI elements
	help             help.Model
	spinner          spinner.Model
	toastMgr         *components.ToastManager
	width            int
	height           int
	isDark           bool
	hasKeyEventTypes bool

	// Context for cancellation
	cancel   context.CancelFunc
	cancelMu *sync.RWMutex
}

// NewApp creates a new AppModel with multi-view support
func NewApp(csvFiles []string, fileProcessor ports.ProcessorController, log ports.LogService, configMgr ports.ConfigManager) *AppModel {
	// Convert CSV files to list items
	items := make([]list.Item, 0, len(csvFiles))
	for _, file := range csvFiles {
		items = append(items, mapFileToOption(file))
	}

	app := &AppModel{
		nav:          NewNavigation(),
		logger:       log,
		processor:    fileProcessor,
		configMgr:    configMgr,
		filesView:    views.NewFilesView(items),
		logsView:     views.NewLogsView(log, fileProcessor),
		settingsView: views.NewSettingsView(configMgr, fileProcessor),
		workersView:  views.NewWorkersView(fileProcessor),
		help:         createHelp(),
		spinner:      createSpinner(),
		toastMgr:     components.NewToastManager(),
		cancelMu:     &sync.RWMutex{},
		isDark:       true,
	}

	app.applyTheme(true)

	return app
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
	return tea.Batch(tea.RequestBackgroundColor, tickCmd(), m.spinner.Tick)
}

func (m *AppModel) applyTheme(isDark bool) {
	m.isDark = isDark
	styles.ApplyTheme(isDark)
	m.filesView.SetTheme(isDark)
	m.help = createHelp()
}

func operationError() logs.Message {
	return logs.NewMessage().
		WithIcon(IconInformation).
		WithMessage("Please wait the current operation to finish or cancel pressing ESC")
}
