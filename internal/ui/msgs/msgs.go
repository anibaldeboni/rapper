package msgs

import (
	"time"

	"github.com/anibaldeboni/rapper/internal/processor"
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
	Metrics processor.Metrics
}
