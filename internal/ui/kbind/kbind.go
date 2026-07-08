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
	Save = key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	)
	Profile = key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "profile"),
	)
	// SliderInc/SliderDec accept both the shifted text ("+", "-") and the
	// shifted keystroke ("shift+=", "shift+-") representations.
	//
	// Why both: Bubble Tea v2 enables the Kitty keyboard protocol
	// (KeyboardEnhancements.ReportEventTypes) so terminals like Kitty,
	// WezTerm, foot, and recent xterm builds send shifted characters as
	// {Text: "", Code: unshifted-rune, Mod: ModShift} — not the legacy
	// {Text: shifted-rune}. Key.String() then renders as "shift+=" rather
	// than "+", and the bubbles Matches helper does plain string equality
	// against the binding keys, so a single WithKeys("+") entry silently
	// swallows the keypress.
	//
	// We keep the unshifted/legacy forms ("+", "-") so terminals that
	// still send the old format (or keypads) keep working, and we add
	// the Kitty keystroke forms so the modern format works too.
	SliderInc = key.NewBinding(
		key.WithKeys("+", "shift+="),
		key.WithHelp("+", "increase workers"),
	)
	SliderDec = key.NewBinding(
		key.WithKeys("-", "shift+-"),
		key.WithHelp("-", "decrease workers"),
	)
	NextField = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	)
	// PaneToggle is the help-text label for Tab when it acts as a
	// pane toggle (the settings view's two-pane model). The binding
	// itself is the same key as NextField (tab); the help text
	// clarifies its role. The settings view re-uses NextField for
	// the actual key match — PaneToggle is documentation only.
	PaneToggle = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "pane"),
	)
	PrevField = key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	)
	CancelOperation = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "kill operation"),
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
