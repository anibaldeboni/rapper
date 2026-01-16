package kbind

import "github.com/charmbracelet/bubbles/key"

var (
	Up = key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	)
	Down = key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	)
	Right = key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "move right"),
	)
	Left = key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "move left"),
	)
	Select = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "run"),
	)
	Cancel = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	)
	Quit = key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	)
	ViewFiles = key.NewBinding(
		key.WithKeys("f1", "ctrl+f"),
		key.WithHelp("F1", "files"),
	)
	ViewLogs = key.NewBinding(
		key.WithKeys("f2", "ctrl+l"),
		key.WithHelp("F2", "logs"),
	)
	ViewSettings = key.NewBinding(
		key.WithKeys("f3", "ctrl+t"),
		key.WithHelp("F3", "settings"),
	)
	ViewWorkers = key.NewBinding(
		key.WithKeys("f4", "ctrl+w"),
		key.WithHelp("F4", "workers"),
	)
	Save = key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	)
	Profile = key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "profile"),
	)
	WorkerInc = key.NewBinding(
		key.WithKeys("right", "+"),
		key.WithHelp("→/+", "increase"),
	)
	WorkerDec = key.NewBinding(
		key.WithKeys("left", "-"),
		key.WithHelp("←/-", "decrease"),
	)
	NextField = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	)
	PrevField = key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	)
	CancelOperation = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "cancel operation"),
	)
)
