package ui

import (
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

// globalKeyMap shows only global navigation keys (for header)
type globalKeyMap struct{}

func (k globalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.ViewFiles, kbind.ViewLogs, kbind.ViewSettings, kbind.ViewWorkers, kbind.CancelOperation, kbind.Quit}
}

func (k globalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{kbind.ViewFiles, kbind.ViewLogs, kbind.ViewSettings, kbind.ViewWorkers, kbind.CancelOperation, kbind.Quit},
	}
}

// getGlobalKeyMap returns the global navigation keymap (for header)
func getGlobalKeyMap() help.KeyMap {
	return globalKeyMap{}
}

// filesViewKeyMap shows only files view specific keys
type filesViewKeyMap struct{}

func (k filesViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.Up, kbind.Down, kbind.Select}
}

func (k filesViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{kbind.Up, kbind.Down, kbind.Select}}
}

// logsViewKeyMap shows only logs view specific keys
type logsViewKeyMap struct{}

func (k logsViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.Left, kbind.Up, kbind.Down, kbind.Right}
}

func (k logsViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{kbind.Left, kbind.Up, kbind.Down, kbind.Right}}
}

// settingsViewKeyMap shows only settings view specific keys
type settingsViewKeyMap struct{}

func (k settingsViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.NextField, kbind.PrevField, kbind.PageUp, kbind.PageDown, kbind.Save, kbind.Profile}
}

func (k settingsViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{kbind.NextField, kbind.PrevField, kbind.Save, kbind.Profile},
		{kbind.PageUp, kbind.PageDown, kbind.GotoTop, kbind.GotoBottom},
	}
}

// workersViewKeyMap shows only workers view specific keys
type workersViewKeyMap struct{}

func (k workersViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.WorkerDec, kbind.WorkerInc, kbind.PageUp, kbind.PageDown}
}

func (k workersViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{kbind.WorkerDec, kbind.WorkerInc},
		{kbind.Up, kbind.Down, kbind.PageUp, kbind.PageDown, kbind.GotoTop, kbind.GotoBottom},
	}
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
