package ui

// View represents different views in the application
type View int

const (
	ViewFiles    View = iota // File selection view
	ViewLogs                 // Execution logs view
	ViewRequests             // Configuration settings view
)

// String returns the name of the view
func (v View) String() string {
	switch v {
	case ViewFiles:
		return "Files"
	case ViewLogs:
		return "Logs"
	case ViewRequests:
		return "Settings"
	default:
		return "Unknown"
	}
}
