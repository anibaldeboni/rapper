package processor

import (
	"context"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/web"
)

// HttpGateway defines the interface for making HTTP requests.
// This interface is defined here (in the client package) following Go idioms.
//
//go:generate mockgen -destination mock/http_gateway_mock.go -package mock_processor github.com/anibaldeboni/rapper/internal/processor HttpGateway
type HttpGateway interface {
	// Exec executes an HTTP request with the given data map
	Exec(ctx context.Context, data map[string]string) (web.Response, error)

	// UpdateConfig updates the gateway configuration
	UpdateConfig(method, urlTemplate, bodyTemplate string, headers map[string]string) error
}

// RequestLogger defines the interface for logging HTTP requests.
// Only the methods actually used by processor are included.
//
//go:generate mockgen -destination mock/request_logger_mock.go -package mock_processor github.com/anibaldeboni/rapper/internal/processor RequestLogger
type RequestLogger interface {
	// Add adds a log message to the in-memory buffer
	Add(msg logs.Message)

	// WriteToFile writes a log line to the output file
	WriteToFile(line logs.Line)
}
