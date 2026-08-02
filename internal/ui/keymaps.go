package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
)

type globalKeyMap struct{}

func (k globalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{kbind.ViewRequests, kbind.ViewFiles, kbind.ViewLogs, kbind.CancelOperation, kbind.Quit}
}

func (k globalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{kbind.ViewRequests, kbind.ViewFiles, kbind.ViewLogs, kbind.CancelOperation, kbind.Quit},
	}
}

func getGlobalKeyMap() help.KeyMap {
	return globalKeyMap{}
}
