package log

import (
	"sync"
)

var _ LogManager = (*logManagerImpl)(nil)

type LogMessage interface {
	String() string
}

//go:generate mockgen -destination mock/log_mock.go github.com/anibaldeboni/rapper/internal/log LogManager
type LogManager interface {
	HasNewLogs() bool
	Add(LogMessage)
	Get() []string
	Len() int
}

type logManagerImpl struct {
	mu    sync.RWMutex
	logs  []LogMessage
	count int
}

// NewLogManager creates a new instance of LogManager.
func NewLogManager() LogManager {
	return &logManagerImpl{}
}

// HasNewLogs checks if there are new logs available.
// It returns true if there are new logs, otherwise false.
func (this *logManagerImpl) HasNewLogs() bool {
	this.mu.RLock()
	defer this.mu.RUnlock()
	if this.count < len(this.logs) {
		this.count = len(this.logs)
		return true
	}
	return false
}

// Add appends a log message to the log manager's logs.
func (this *logManagerImpl) Add(log LogMessage) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.logs = append(this.logs, log)
}

// Get returns all logs as a slice of strings.
func (this *logManagerImpl) Get() []string {
	this.mu.RLock()
	defer this.mu.RUnlock()
	var logs []string
	for _, log := range this.logs {
		logs = append(logs, log.String())
	}
	return logs
}

// Len returns the number of logs.
func (this *logManagerImpl) Len() int {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return len(this.logs)
}
