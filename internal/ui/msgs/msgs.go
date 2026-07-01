package msgs

import (
	"time"
)

type TickMsg time.Time

// MetricsTickMsg fires every 100ms while the Logs view is active and drives
// the real-time metrics panel refresh. Defined here so any view or component
// can subscribe to the same tick stream without depending on the deleted
// `views` workers package.
type MetricsTickMsg time.Time

// ViewportSizeMsg is the chrome-adjusted terminal dimensions the views
// receive in place of the raw tea.WindowSizeMsg. Width/Height are
// already reduced by the chrome margins/header/status-bar, so views are
// chrome-ignorant. Synthesized by AppModel on every tea.WindowSizeMsg
// and routed to every view in the views map.
type ViewportSizeMsg struct {
	Width  int
	Height int
}

// ItemSelectedMsg is emitted by FilesView when the user presses the
// Select key (Enter) on a focused file. AppModel handles the message
// and calls selectFile — this removes the need for AppModel to query
// the view's SelectedItem.
type ItemSelectedMsg struct {
	FilePath string
}

// ThemeAppliedMsg is broadcast to all views when the theme changes
// (initial theme application and on tea.BackgroundColorMsg). The
// message replaces the imperative SetTheme call on each view.
type ThemeAppliedMsg struct {
	IsDark bool
}

// MetricsVisibilityMsg starts (Visible=true) or stops (Visible=false)
// the metrics tick chain. Sent to all views on navigation changes;
// only LogsView actually handles it.
type MetricsVisibilityMsg struct {
	Visible bool
}

// ConfigSavedMsg is sent when configuration is successfully saved
type ConfigSavedMsg struct{}

// ConfigSaveErrorMsg is sent when configuration save fails
type ConfigSaveErrorMsg struct {
	Err error
}

// ProfileSwitchedMsg is sent when profile is successfully switched
type ProfileSwitchedMsg struct {
	ProfileName string
}

// ProfileSwitchErrorMsg is sent when profile switch fails
type ProfileSwitchErrorMsg struct {
	Err error
}

// ProcessingStartedMsg is sent when file processing begins
type ProcessingStartedMsg struct {
	FilePath string
}

// ProcessingStoppedMsg is sent when processing completes or is cancelled
type ProcessingStoppedMsg struct {
	Success bool
	Err     error
}

// ProcessingProgressMsg is sent periodically during processing with metrics
type ProcessingProgressMsg struct {
	TotalRequests   uint64
	SuccessRequests uint64
	ErrorRequests   uint64
	LinesProcessed  uint64
	ActiveWorkers   int
	RequestsPerSec  float64
	StartTime       time.Time
	IsProcessing    bool
}
