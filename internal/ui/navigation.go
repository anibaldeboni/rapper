package ui

// View represents different views in the application
type View int

const (
	ViewFiles    View = iota // File selection view
	ViewLogs                 // Execution logs view
	ViewSettings             // Configuration settings view
	ViewWorkers              // Workers control view
)

// Navigation manages view navigation with history
type Navigation struct {
	current View
	history []View
}

// NewNavigation creates a new Navigation instance
func NewNavigation() *Navigation {
	return &Navigation{
		current: ViewFiles,
		history: make([]View, 0),
	}
}

// Current returns the current active view
func (n *Navigation) Current() View {
	return n.current
}

// Push switches to a new view and adds current to history
func (n *Navigation) Push(view View) {
	if view != n.current {
		n.history = append(n.history, n.current)
		n.current = view
	}
}

// Back returns to the previous view in history
func (n *Navigation) Back() View {
	if len(n.history) > 0 {
		// Pop last view from history
		n.current = n.history[len(n.history)-1]
		n.history = n.history[:len(n.history)-1]
	}
	return n.current
}

// Set directly sets the current view without affecting history
func (n *Navigation) Set(view View) {
	n.current = view
}

// CanGoBack returns true if there is history to go back to
func (n *Navigation) CanGoBack() bool {
	return len(n.history) > 0
}

// ClearHistory clears the navigation history
func (n *Navigation) ClearHistory() {
	n.history = make([]View, 0)
}

// String returns the name of the view
func (v View) String() string {
	switch v {
	case ViewFiles:
		return "Files"
	case ViewLogs:
		return "Logs"
	case ViewSettings:
		return "Settings"
	case ViewWorkers:
		return "Workers"
	default:
		return "Unknown"
	}
}
