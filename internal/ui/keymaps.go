package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/views"
)

type globalKeyMap struct{}

func (k globalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.ViewFiles, kbind.ViewLogs, kbind.ViewSettings, kbind.CancelOperation, kbind.Quit}
}

func (k globalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{kbind.ViewFiles, kbind.ViewLogs, kbind.ViewSettings, kbind.CancelOperation, kbind.Quit},
	}
}

func getGlobalKeyMap() help.KeyMap {
	return globalKeyMap{}
}

type filesViewKeyMap struct{}

func (k filesViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.Up, kbind.Down, kbind.Select}
}

func (k filesViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{kbind.Up, kbind.Down, kbind.Select}}
}

type logsViewKeyMap struct{}

func (k logsViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.Up, kbind.Down, kbind.Select, kbind.GotoBottom}
}

func (k logsViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{
		kbind.Up, kbind.Down, kbind.GotoTop, kbind.GotoBottom,
		kbind.PageUp, kbind.PageDown, kbind.Select,
	}}
}

func getViewSpecificKeyMap(view View, vm viewModel) help.KeyMap {
	switch view {
	case ViewFiles:
		return filesViewKeyMap{}
	case ViewLogs:
		return logsViewKeyMap{}
	case ViewSettings:
		if sv, ok := vm.(views.SettingsView); ok {
			return sv.KeyMap()
		}
		return globalKeyMap{}
	default:

		return globalKeyMap{}
	}
}
