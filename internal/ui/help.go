package ui

import (
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	ListUp              key.Binding
	ListDown            key.Binding
	Quit                key.Binding
	Select              key.Binding
	Cancel              key.Binding
	LogUp               key.Binding
	LogDown             key.Binding
	ViewFiles           key.Binding
	ViewLogs            key.Binding
	ViewSettings        key.Binding
	ViewWorkers         key.Binding
	Save                key.Binding
	Profile             key.Binding
	WorkerInc           key.Binding
	WorkerDec           key.Binding
	SwitchFieldForward  key.Binding
	SwitchFieldBackward key.Binding
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
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	LogUp: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "scroll up"),
	),
	LogDown: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "scroll down"),
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
	SwitchFieldForward: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	SwitchFieldBackward: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
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
	Save                key.Binding
	Profile             key.Binding
	Cancel              key.Binding
	SwitchFieldForward  key.Binding
	SwitchFieldBackward key.Binding
}

func (k settingsViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.SwitchFieldForward, k.SwitchFieldBackward, k.Save, k.Profile, k.Cancel}
}

func (k settingsViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.SwitchFieldForward, k.SwitchFieldBackward, k.Save, k.Profile, k.Cancel}}
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
			Save:                keys.Save,
			Profile:             keys.Profile,
			SwitchFieldForward:  keys.SwitchFieldForward,
			SwitchFieldBackward: keys.SwitchFieldBackward,
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
