package cli

import (
	"sync"
)

type Logs struct {
	mu   sync.RWMutex
	logs []string
}

func (l *Logs) Add(log string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, log)
}
func (l *Logs) Get() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.logs
}
func (l *Logs) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.logs)
}
