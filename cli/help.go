package cli

import (
	"github.com/anibaldeboni/rapper/internal/styles"
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

	help.Styles.ShortKey = styles.HelpKeyStyle
	help.Styles.ShortDesc = styles.HelpDescStyle
	help.Styles.ShortSeparator = styles.HelpSepStyle
	help.Styles.Ellipsis = styles.HelpSepStyle.Copy()
	help.Styles.FullKey = styles.HelpKeyStyle.Copy()
	help.Styles.FullDesc = styles.HelpDescStyle.Copy()
	help.Styles.FullSeparator = styles.HelpSepStyle.Copy()

	return help
}
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Cancel, k.Quit, k.Help}
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
