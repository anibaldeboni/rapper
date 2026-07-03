package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
)

// globalKeyMap shows only global navigation keys (for header)
type globalKeyMap struct{}

func (k globalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.ViewFiles, kbind.ViewLogs, kbind.ViewSettings, kbind.CancelOperation, kbind.Quit}
}

func (k globalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{kbind.ViewFiles, kbind.ViewLogs, kbind.ViewSettings, kbind.CancelOperation, kbind.Quit},
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

// logsViewKeyMap shows only logs view specific keys. The list is
// vertical-only — Left/Right are not handled by the DetailedList
// anymore — so the keymap reflects the available navigation
// (Up/Down/PgUp/PgDn/Home/End) plus Enter to expand a row.
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

// settingsViewKeyMap shows only settings view specific keys.
// The slider keybindings are listed so users discover +/- while the slider
// has focus; the slider itself only intercepts +/- when focused.
type settingsViewKeyMap struct{}

func (k settingsViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.NextField, kbind.PrevField, kbind.PageUp, kbind.PageDown, kbind.Save, kbind.Profile, kbind.SliderInc, kbind.SliderDec}
}

func (k settingsViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{kbind.NextField, kbind.PrevField, kbind.Save, kbind.Profile},
		{kbind.SliderInc, kbind.SliderDec},
		{kbind.PageUp, kbind.PageDown, kbind.GotoTop, kbind.GotoBottom},
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
	default:
		// Return empty keymap for unknown views
		return globalKeyMap{}
	}
}
