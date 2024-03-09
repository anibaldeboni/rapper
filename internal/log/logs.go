package log

import (
	"sync"
)

type LogMessage interface {
	String() string
}

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

func NewLogManager() LogManager {
	return &logManagerImpl{}
}

func (l *logManagerImpl) HasNewLogs() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.count < len(l.logs) {
		l.count = len(l.logs)
		return true
	}
	return false
}

func (l *logManagerImpl) Add(log LogMessage) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, log)
}
func (l *logManagerImpl) Get() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var logs []string
	for _, log := range l.logs {
		logs = append(logs, log.String())
	}
	return logs
}
func (l *logManagerImpl) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.logs)
}
