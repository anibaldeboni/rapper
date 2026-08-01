package kbind

import "charm.land/bubbles/v2/key"

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
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "quit"),
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
	Save = key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	)
	Profile = key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "profile"),
	)
	CancelOperation = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "kill operation"),
	)
	// Edit is the binding the Settings view uses to enter edit mode
	// (focus the raw-YAML textarea). The same `e` key is used
	// regardless of the pane state — the binding is documentation
	// only; the actual key match happens in SettingsView.Update.
	Edit = key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	)
	// ExitEdit is the binding the Settings view uses to leave edit
	// mode (blur the raw-YAML textarea and return to selection mode).
	// `esc` is chosen over `shift+e` because the textarea is a text
	// input: extracted text bindings (typing "e", shift+e produces
	// "E") would conflict with an alphanumeric exit key. Esc is a
	// non-printing key the editor won't insert into the buffer.
	ExitEdit = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "exit edit"),
	)
	GotoBottom = key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", "go to bottom"),
	)
	GotoTop = key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("home", "go to top"),
	)
	PageUp = key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	)
	PageDown = key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "page down"),
	)
)
