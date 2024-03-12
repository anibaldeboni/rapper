package log

import (
	"sync"
)

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
func (l *logManagerImpl) HasNewLogs() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.count < len(l.logs) {
		l.count = len(l.logs)
		return true
	}
	return false
}

// Add appends a log message to the log manager's logs.
func (l *logManagerImpl) Add(log LogMessage) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, log)
}

// Get returns all logs as a slice of strings.
func (l *logManagerImpl) Get() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var logs []string
	for _, log := range l.logs {
		logs = append(logs, log.String())
	}
	return logs
}

// Len returns the number of logs.
func (l *logManagerImpl) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.logs)
}
