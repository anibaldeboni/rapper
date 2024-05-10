package logs

import (
	"os"
	"sync"

	"github.com/anibaldeboni/rapper/internal/styles"
)

var _ Logger = (*loggerImpl)(nil)

type Message interface {
	String() string
	WithIcon(string) Message
	WithKind(string) Message
	WithMessage(string) Message
}

//go:generate mockgen -destination mock/log_mock.go github.com/anibaldeboni/rapper/internal/logs Logger
type Logger interface {
	HasNewLogs() bool
	Add(Message)
	Get() []string
	Len() int
	WriteToFile(LogLiner)
}

type loggerImpl struct {
	sync.RWMutex
	logs    []Message
	logFile *os.File
	count   int
}

// NewLoggger creates a new instance of Loggger.
func NewLoggger(filePath string) Logger {
	var logger loggerImpl
	if filePath != "" {
		if file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660); err != nil {
			logger.Add(errorMessage(err.Error()))
		} else {
			logger.logFile = file
		}
	}
	return &logger
}

// HasNewLogs checks if there are new logs available.
// It returns true if there are new logs, otherwise false.
func (this *loggerImpl) HasNewLogs() bool {
	this.RLock()
	defer this.RUnlock()
	if this.count < len(this.logs) {
		this.count = len(this.logs)
		return true
	}
	return false
}

// Add appends a log message to the log manager's logs.
func (this *loggerImpl) Add(log Message) {
	this.Lock()
	defer this.Unlock()
	this.logs = append(this.logs, log)
}

// Get returns all logs as a slice of strings.
func (this *loggerImpl) Get() []string {
	this.RLock()
	defer this.RUnlock()
	var logs []string
	for _, log := range this.logs {
		logs = append(logs, log.String())
	}
	return logs
}

// Len returns the number of logs.
func (this *loggerImpl) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.logs)
}

type LogLiner interface {
	String() []byte
}

// WriteToFile writes a line to the file logger.
// If the logger is enabled, it acquires a lock, writes the line to the file,
// and adds an error message to the log manager if there was an error writing to the file.
func (this *loggerImpl) WriteToFile(line LogLiner) {
	if this.logFile != nil {
		if err := write(this.logFile, line); err != nil {
			this.Add(errorMessage(err.Error()))
		}
	}
}

func errorMessage(message string) Message {
	return NewMessage().
		WithIcon(styles.IconSkull).
		WithKind("Output").
		WithMessage(message)
}

func write(file *os.File, line LogLiner) error {
	if _, err := file.Write(append(line.String(), '\n')); err != nil {
		return err
	}
	return nil
}
