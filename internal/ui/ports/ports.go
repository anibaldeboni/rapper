package ports

import (
	"context"

	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/processor"
)

// ConfigManager defines the interface for managing configuration.
// Used by SettingsView which needs full config management capabilities.
//
//go:generate mockgen -destination ../mock/config_manager_mock.go -package mock_ui github.com/anibaldeboni/rapper/internal/ui/ports ConfigManager
type ConfigManager interface {
	// Get returns the active configuration
	Get() *config.Config

	// Update updates the active configuration in memory
	Update(cfg *config.Config) error

	// Save persists the configuration to disk
	Save() error

	// Profile management methods
	// ListProfiles returns names of all available profiles
	ListProfiles() []string

	// GetActiveProfile returns the name of the active profile
	GetActiveProfile() string

	// SetActiveProfile switches to the specified profile
	SetActiveProfile(name string) error
}

// ConfigProvider defines a read-only configuration interface.
// Used by components that only need to read config and watch for changes.
//
//go:generate mockgen -destination ../mock/config_provider_mock.go -package mock_ui github.com/anibaldeboni/rapper/internal/ui/ports ConfigProvider
type ConfigProvider interface {
	// Get returns the active configuration
	Get() *config.Config

	// OnChange registers a callback for configuration changes
	OnChange(callback func(*config.Config))
}

// ProcessorController defines the interface for controlling request processing.
// Used by AppModel and WorkersView to start processing and monitor status.
//
//go:generate mockgen -destination ../mock/processor_controller_mock.go -package mock_ui github.com/anibaldeboni/rapper/internal/ui/ports ProcessorController
type ProcessorController interface {
	// Do starts processing a CSV file
	Do(ctx context.Context, filePath string) (context.Context, context.CancelFunc)

	// GetMetrics returns current processing metrics
	GetMetrics() ProcessorMetrics

	// SetWorkers dynamically adjusts the number of workers
	SetWorkers(n int)

	// GetWorkerCount returns the current worker count
	GetWorkerCount() int
}

// ProcessorMetrics holds real-time processing metrics.
// This is an alias to processor.Metrics to avoid duplicating the type.
type ProcessorMetrics = processor.Metrics

// LogProvider defines the interface for reading logs.
// Used by LogsView to display execution logs.
//
//go:generate mockgen -destination ../mock/log_provider_mock.go -package mock_ui github.com/anibaldeboni/rapper/internal/ui/ports LogProvider
type LogProvider interface {
	// Get returns all log messages as strings
	Get() []string
}

// RequestLogger defines the interface for adding log messages.
// Used by AppModel to log errors and info messages.
//
//go:generate mockgen -destination ../mock/request_logger_mock.go -package mock_ui github.com/anibaldeboni/rapper/internal/ui/ports RequestLogger
type RequestLogger interface {
	// Add adds a log message
	Add(msg logs.Message)
}

// LogService combines log provider and logger behaviors.
//
//go:generate mockgen -destination ../mock/log_service_mock.go -package mock_ui github.com/anibaldeboni/rapper/internal/ui/ports LogService
type LogService interface {
	LogProvider
	RequestLogger
}

// LogMessage represents a displayable log message.
// Deprecated: prefer logs.Message directly via RequestLogger.
type LogMessage interface {
	String() string
}
