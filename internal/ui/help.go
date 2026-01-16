package ui

import (
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Up                  key.Binding
	Down                key.Binding
	Right               key.Binding
	Left                key.Binding
	Quit                key.Binding
	Select              key.Binding
	Cancel              key.Binding
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
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "move right"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "move left"),
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
type globalKeyMap struct{}

func (k globalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{keys.ViewFiles, keys.ViewLogs, keys.ViewSettings, keys.ViewWorkers, keys.Quit}
}

func (k globalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{keys.ViewFiles, keys.ViewLogs, keys.ViewSettings, keys.ViewWorkers, keys.Quit},
	}
}

// getGlobalKeyMap returns the global navigation keymap (for header)
func getGlobalKeyMap() help.KeyMap {
	return globalKeyMap{}
}

// filesViewKeyMap shows only files view specific keys
type filesViewKeyMap struct{}

func (k filesViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{keys.Up, keys.Down, keys.Select, keys.Cancel}
}

func (k filesViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{keys.Up, keys.Down, keys.Select, keys.Cancel}}
}

// logsViewKeyMap shows only logs view specific keys
type logsViewKeyMap struct{}

func (k logsViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{keys.Up, keys.Right, keys.Down, keys.Left}
}

func (k logsViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{keys.Up, keys.Right, keys.Down, keys.Left}}
}

// settingsViewKeyMap shows only settings view specific keys
type settingsViewKeyMap struct{}

func (k settingsViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{keys.SwitchFieldForward, keys.SwitchFieldBackward, keys.Save, keys.Profile, keys.Cancel}
}

func (k settingsViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{keys.SwitchFieldForward, keys.SwitchFieldBackward, keys.Save, keys.Profile, keys.Cancel}}
}

// workersViewKeyMap shows only workers view specific keys
type workersViewKeyMap struct{}

func (k workersViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{keys.WorkerInc, keys.WorkerDec}
}

func (k workersViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{keys.WorkerInc, keys.WorkerDec}}
}

// getViewSpecificKeyMap returns only view-specific keys (excluding global navigation)
func getViewSpecificKeyMap(view View) help.KeyMap {
	switch view {
	case ViewFiles:
		return filesViewKeyMap{}
	case ViewLogs:
		return logsViewKeyMap{}
	case ViewSettings:
		return settingsViewKeyMap{}
	case ViewWorkers:
		return workersViewKeyMap{}
	default:
		// Return empty keymap for unknown views
		return globalKeyMap{}
	}
}
