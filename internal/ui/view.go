package ui

// View represents different views in the application
type View int

const (
	ViewFiles    View = iota // File selection view
	ViewLogs                 // Execution logs view
	ViewSettings             // Configuration settings view
)

// String returns the name of the view
func (v View) String() string {
	switch v {
	case ViewFiles:
		return "Files"
	case ViewLogs:
		return "Logs"
	case ViewSettings:
		return "Settings"
	default:
		return "Unknown"
	}
}
