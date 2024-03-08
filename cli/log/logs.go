package log

import (
	"sync"
)

type LogMessage interface {
	String() string
}

type Logs struct {
	mu    sync.RWMutex
	logs  []LogMessage
	count int
}

func (l *Logs) HasNewLogs() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.count < len(l.logs) {
		l.count = len(l.logs)
		return true
	}
	return false
}

func (l *Logs) Add(log LogMessage) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, log)
}
func (l *Logs) Get() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var logs []string
	for _, log := range l.logs {
		logs = append(logs, log.String())
	}
	return logs
}
func (l *Logs) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.logs)
}
