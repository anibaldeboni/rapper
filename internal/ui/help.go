package ui

import (
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	ListUp      key.Binding
	ListDown    key.Binding
	Help        key.Binding
	Quit        key.Binding
	Select      key.Binding
	Cancel      key.Binding
	LogUp       key.Binding
	LogDown     key.Binding
	ViewFiles   key.Binding
	ViewLogs    key.Binding
	ViewSettings key.Binding
	ViewWorkers key.Binding
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
	ViewFiles: key.NewBinding(
		key.WithKeys("f1", "ctrl+f"),
		key.WithHelp("F1", "files"),
	),
	ViewLogs: key.NewBinding(
		key.WithKeys("f2", "ctrl+l"),
		key.WithHelp("F2", "logs"),
	),
	ViewSettings: key.NewBinding(
		key.WithKeys("f3", "ctrl+s"),
		key.WithHelp("F3", "settings"),
	),
	ViewWorkers: key.NewBinding(
		key.WithKeys("f4", "ctrl+w"),
		key.WithHelp("F4", "workers"),
	),
}
