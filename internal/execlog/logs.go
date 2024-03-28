package execlog

import (
	"sync"
)

var _ Manager = (*logManagerImpl)(nil)

type Message interface {
	String() string
	WithIcon(string) Message
	WithKind(string) Message
	WithMessage(string) Message
}

//go:generate mockgen -destination mock/log_mock.go github.com/anibaldeboni/rapper/internal/execlog Manager
type Manager interface {
	HasNewLogs() bool
	Add(Message)
	Get() []string
	Len() int
}

type logManagerImpl struct {
	sync.RWMutex
	logs  []Message
	count int
}

// NewLogManager creates a new instance of LogManager.
func NewLogManager() Manager {
	return &logManagerImpl{}
}

// HasNewLogs checks if there are new logs available.
// It returns true if there are new logs, otherwise false.
func (this *logManagerImpl) HasNewLogs() bool {
	this.RLock()
	defer this.RUnlock()
	if this.count < len(this.logs) {
		this.count = len(this.logs)
		return true
	}
	return false
}

// Add appends a log message to the log manager's logs.
func (this *logManagerImpl) Add(log Message) {
	this.Lock()
	defer this.Unlock()
	this.logs = append(this.logs, log)
}

// Get returns all logs as a slice of strings.
func (this *logManagerImpl) Get() []string {
	this.RLock()
	defer this.RUnlock()
	var logs []string
	for _, log := range this.logs {
		logs = append(logs, log.String())
	}
	return logs
}

// Len returns the number of logs.
func (this *logManagerImpl) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.logs)
}
