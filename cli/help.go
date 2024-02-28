package cli

import (
	"github.com/anibaldeboni/rapper/cli/ui"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	ListUp   key.Binding
	ListDown key.Binding
	Help     key.Binding
	Quit     key.Binding
	Select   key.Binding
	Cancel   key.Binding
	LogUp    key.Binding
	LogDown  key.Binding
}

func createHelp() help.Model {
	help := help.New()
	help.Styles.ShortDesc = ui.HelpStyle
	help.Styles.FullDesc = ui.HelpStyle
	return help
}
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Cancel, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ListUp, k.ListDown},
		{k.LogUp, k.LogDown},
		{k.Select, k.Cancel},
		{k.Help, k.Quit},
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
		key.WithHelp("esc", "cancel"),
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
		key.WithHelp("shift + ↑", "move log up"),
	),
	LogDown: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("shift + ↓", "move log down"),
	),
}
