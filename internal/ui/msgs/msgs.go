package msgs

import (
	"time"
)

type TickMsg time.Time

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
