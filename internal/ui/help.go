package ui

import (
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	ListUp       key.Binding
	ListDown     key.Binding
	Help         key.Binding
	Quit         key.Binding
	Select       key.Binding
	Cancel       key.Binding
	LogUp        key.Binding
	LogDown      key.Binding
	ViewFiles    key.Binding
	ViewLogs     key.Binding
	ViewSettings key.Binding
	ViewWorkers  key.Binding
	Save         key.Binding
	Profile      key.Binding
	WorkerInc    key.Binding
	WorkerDec    key.Binding
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

// filesKeyMap shows keys available in Files view
type filesKeyMap struct {
	Select       key.Binding
	Cancel       key.Binding
	ViewFiles    key.Binding
	ViewLogs     key.Binding
	ViewSettings key.Binding
	ViewWorkers  key.Binding
	Quit         key.Binding
}

func (k filesKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Cancel, k.ViewFiles, k.ViewLogs, k.ViewSettings, k.ViewWorkers, k.Quit}
}

func (k filesKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Select, k.Cancel},
		{k.ViewFiles, k.ViewLogs, k.ViewSettings, k.ViewWorkers},
		{k.Quit},
	}
}

// logsKeyMap shows keys available in Logs view
type logsKeyMap struct {
	LogUp        key.Binding
	LogDown      key.Binding
	ViewFiles    key.Binding
	ViewLogs     key.Binding
	ViewSettings key.Binding
	ViewWorkers  key.Binding
	Quit         key.Binding
}

func (k logsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.LogUp, k.LogDown, k.ViewFiles, k.ViewLogs, k.ViewSettings, k.ViewWorkers, k.Quit}
}

func (k logsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.LogUp, k.LogDown},
		{k.ViewFiles, k.ViewLogs, k.ViewSettings, k.ViewWorkers},
		{k.Quit},
	}
}

// settingsKeyMap shows keys available in Settings view
type settingsKeyMap struct {
	Save         key.Binding
	Profile      key.Binding
	Cancel       key.Binding
	ViewFiles    key.Binding
	ViewLogs     key.Binding
	ViewSettings key.Binding
	ViewWorkers  key.Binding
	Quit         key.Binding
}

func (k settingsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Save, k.Profile, k.Cancel, k.ViewFiles, k.ViewLogs, k.ViewSettings, k.ViewWorkers, k.Quit}
}

func (k settingsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Save, k.Profile, k.Cancel},
		{k.ViewFiles, k.ViewLogs, k.ViewSettings, k.ViewWorkers},
		{k.Quit},
	}
}

// workersKeyMap shows keys available in Workers view
type workersKeyMap struct {
	WorkerInc    key.Binding
	WorkerDec    key.Binding
	ViewFiles    key.Binding
	ViewLogs     key.Binding
	ViewSettings key.Binding
	ViewWorkers  key.Binding
	Quit         key.Binding
}

func (k workersKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.WorkerInc, k.WorkerDec, k.ViewFiles, k.ViewLogs, k.ViewSettings, k.ViewWorkers, k.Quit}
}

func (k workersKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.WorkerInc, k.WorkerDec},
		{k.ViewFiles, k.ViewLogs, k.ViewSettings, k.ViewWorkers},
		{k.Quit},
	}
}

func (this keyMap) ShortHelp() []key.Binding {
	return []key.Binding{this.Select, this.Cancel, this.ViewFiles, this.ViewLogs, this.ViewSettings, this.ViewWorkers, this.Quit}
}

func (this keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{this.ViewFiles, this.ViewLogs, this.ViewSettings, this.ViewWorkers},
		{this.ListUp, this.ListDown},
		{this.LogUp, this.LogDown},
		{this.Select, this.Cancel},
		{this.Help, this.Quit},
	}
}

var keys = keyMap{
	ListUp: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move file selection up"),
	),
	ListDown: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move file selection down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "run"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	LogUp: key.NewBinding(
		key.WithKeys("shift+up"),
		key.WithHelp("shift+↑", "scroll up"),
	),
	LogDown: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("shift+↓", "scroll down"),
	),
	ViewFiles: key.NewBinding(
		key.WithKeys("f1", "ctrl+f"),
		key.WithHelp("F1", "files"),
	),
	ViewLogs: key.NewBinding(
		key.WithKeys("f2", "ctrl+l"),
		key.WithHelp("F2", "logs"),
	),
	ViewSettings: key.NewBinding(
		key.WithKeys("f3", "ctrl+t"),
		key.WithHelp("F3", "settings"),
	),
	ViewWorkers: key.NewBinding(
		key.WithKeys("f4", "ctrl+w"),
		key.WithHelp("F4", "workers"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Profile: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "profile"),
	),
	WorkerInc: key.NewBinding(
		key.WithKeys("right", "+"),
		key.WithHelp("→/+", "increase"),
	),
	WorkerDec: key.NewBinding(
		key.WithKeys("left", "-"),
		key.WithHelp("←/-", "decrease"),
	),
}

// globalKeyMap shows only global navigation keys (for header)
type globalKeyMap struct {
	ViewFiles    key.Binding
	ViewLogs     key.Binding
	ViewSettings key.Binding
	ViewWorkers  key.Binding
	Quit         key.Binding
}

func (k globalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ViewFiles, k.ViewLogs, k.ViewSettings, k.ViewWorkers, k.Quit}
}

func (k globalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ViewFiles, k.ViewLogs, k.ViewSettings, k.ViewWorkers, k.Quit},
	}
}

// getGlobalKeyMap returns the global navigation keymap (for header)
func getGlobalKeyMap() help.KeyMap {
	return globalKeyMap{
		ViewFiles:    keys.ViewFiles,
		ViewLogs:     keys.ViewLogs,
		ViewSettings: keys.ViewSettings,
		ViewWorkers:  keys.ViewWorkers,
		Quit:         keys.Quit,
	}
}

// getContextualKeyMap returns the appropriate keymap based on the current view (for status bar)
func getContextualKeyMap(view View) help.KeyMap {
	switch view {
	case ViewFiles:
		return filesKeyMap{
			Select:       keys.Select,
			Cancel:       keys.Cancel,
			ViewFiles:    keys.ViewFiles,
			ViewLogs:     keys.ViewLogs,
			ViewSettings: keys.ViewSettings,
			ViewWorkers:  keys.ViewWorkers,
			Quit:         keys.Quit,
		}
	case ViewLogs:
		return logsKeyMap{
			LogUp:        keys.LogUp,
			LogDown:      keys.LogDown,
			ViewFiles:    keys.ViewFiles,
			ViewLogs:     keys.ViewLogs,
			ViewSettings: keys.ViewSettings,
			ViewWorkers:  keys.ViewWorkers,
			Quit:         keys.Quit,
		}
	case ViewSettings:
		return settingsKeyMap{
			Save:         keys.Save,
			Profile:      keys.Profile,
			Cancel:       keys.Cancel,
			ViewFiles:    keys.ViewFiles,
			ViewLogs:     keys.ViewLogs,
			ViewSettings: keys.ViewSettings,
			ViewWorkers:  keys.ViewWorkers,
			Quit:         keys.Quit,
		}
	case ViewWorkers:
		return workersKeyMap{
			WorkerInc:    keys.WorkerInc,
			WorkerDec:    keys.WorkerDec,
			ViewFiles:    keys.ViewFiles,
			ViewLogs:     keys.ViewLogs,
			ViewSettings: keys.ViewSettings,
			ViewWorkers:  keys.ViewWorkers,
			Quit:         keys.Quit,
		}
	default:
		return keys
	}
}

// filesViewKeyMap shows only files view specific keys
type filesViewKeyMap struct {
	Select key.Binding
	Cancel key.Binding
}

func (k filesViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Cancel}
}

func (k filesViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Select, k.Cancel}}
}

// logsViewKeyMap shows only logs view specific keys
type logsViewKeyMap struct {
	LogUp   key.Binding
	LogDown key.Binding
}

func (k logsViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.LogUp, k.LogDown}
}

func (k logsViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.LogUp, k.LogDown}}
}

// settingsViewKeyMap shows only settings view specific keys
type settingsViewKeyMap struct {
	Save    key.Binding
	Profile key.Binding
	Cancel  key.Binding
}

func (k settingsViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Save, k.Profile, k.Cancel}
}

func (k settingsViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Save, k.Profile, k.Cancel}}
}

// workersViewKeyMap shows only workers view specific keys
type workersViewKeyMap struct {
	WorkerInc key.Binding
	WorkerDec key.Binding
}

func (k workersViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.WorkerInc, k.WorkerDec}
}

func (k workersViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.WorkerInc, k.WorkerDec}}
}

// getViewSpecificKeyMap returns only view-specific keys (excluding global navigation)
func getViewSpecificKeyMap(view View) help.KeyMap {
	switch view {
	case ViewFiles:
		return filesViewKeyMap{
			Select: keys.Select,
			Cancel: keys.Cancel,
		}
	case ViewLogs:
		return logsViewKeyMap{
			LogUp:   keys.LogUp,
			LogDown: keys.LogDown,
		}
	case ViewSettings:
		return settingsViewKeyMap{
			Save:    keys.Save,
			Profile: keys.Profile,
			Cancel:  keys.Cancel,
		}
	case ViewWorkers:
		return workersViewKeyMap{
			WorkerInc: keys.WorkerInc,
			WorkerDec: keys.WorkerDec,
		}
	default:
		// Return empty keymap for unknown views
		return globalKeyMap{}
	}
}

