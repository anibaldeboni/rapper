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
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	"github.com/anibaldeboni/rapper/internal/ui/views"
)

const AppName = "rapper"

var version string

var AppVersion = getVersion()

func getVersion() string {

	if version != "" {
		return normalizeVersion(version)
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

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

	version := info.Main.Version
	if version != "" && version != "(devel)" && !isPseudoVersion(version) {

		return normalizeVersion(version)
	}

	if revision != "" {

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

func isPseudoVersion(version string) bool {

	return strings.Contains(version, "-0.") ||
		strings.Contains(version, "+incompatible") ||
		(strings.Count(version, "-") >= 2 && len(version) > 25)
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

type viewModel = tea.Model

type AppModel struct {
	currentView         View
	currentStatusKeymap help.KeyMap

	logger    ports.LogService
	processor ports.ProcessorController
	configMgr ports.ConfigManager

	views map[View]viewModel

	help             help.Model
	spinner          spinner.Model
	toastMgr         *components.ToastManager
	width            int
	height           int
	isDark           bool
	hasKeyEventTypes bool

	chrome ChromeLayout

	cancel   context.CancelFunc
	cancelMu *sync.RWMutex
}

func NewApp(csvFiles []string, fileProcessor ports.ProcessorController, log ports.LogService, configMgr ports.ConfigManager) *AppModel {

	items := make([]list.Item, 0, len(csvFiles))
	for _, file := range csvFiles {
		items = append(items, mapFileToOption(file))
	}

	initialFilesView := views.NewFilesView(items)
	app := &AppModel{
		currentView:         ViewFiles,
		currentStatusKeymap: initialFilesView.KeyMap(),
		logger:              log,
		processor:           fileProcessor,
		configMgr:           configMgr,
		views: map[View]viewModel{
			ViewFiles:    initialFilesView,
			ViewLogs:     views.NewLogsView(log, fileProcessor),
			ViewSettings: views.NewSettingsView(configMgr),
		},
		help:     createHelp(),
		spinner:  createSpinner(),
		toastMgr: components.NewToastManager(),
		cancelMu: &sync.RWMutex{},
		chrome:   NewChromeLayout(),
		isDark:   true,
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

	for k, v := range m.views {
		next, _ := v.Update(msgs.ThemeAppliedMsg{IsDark: isDark})
		m.views[k] = next
	}
	m.help = createHelp()
}

func operationError() logs.LogMessage {
	return logs.NewMessage("Please wait the current operation to finish or cancel pressing ESC", logs.WithIcon(styles.IconWarning), logs.AsGeneral())
}
