package logs

import (
	"fmt"
	"os"
	"sync"

	"github.com/anibaldeboni/rapper/internal/styles"
)

var _ Logger = (*loggerImpl)(nil)

//go:generate mockgen -destination mock/log_mock.go github.com/anibaldeboni/rapper/internal/logs Logger
type Logger interface {
	Add(Message)
	Get() []string
	WriteToFile(Line)
}

type Line interface {
	Bytes() []byte
}

type loggerImpl struct {
	sync.RWMutex
	messages []Message
	file     *os.File
}

// NewLoggger creates a new instance of Loggger.
func NewLoggger(filePath string) Logger {
	var logger loggerImpl
	if filePath != "" {
		if file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660); err != nil {
			logger.Add(errorMessage(err.Error()))
		} else {
			logger.file = file
		}
	}
	return &logger
}

// Add appends a log message to the log manager's logs.
func (l *loggerImpl) Add(log Message) {
	l.Lock()
	defer l.Unlock()
	l.messages = append(l.messages, log)
}

// Get returns all logs as a slice of strings.
func (l *loggerImpl) Get() []string {
	l.RLock()
	defer l.RUnlock()
	var logs []string
	for _, log := range l.messages {
		logs = append(logs, log.String())
	}
	return logs
}

// WriteToFile writes the given log line to the log file, if it is open.
// If there is an error while writing to the file, an error message is added to the logger.
func (l *loggerImpl) WriteToFile(line Line) {
	if l.file != nil {
		if err := write(l.file, line); err != nil {
			l.Add(errorMessage(err.Error()))
		}
	}
}

func errorMessage(message string) Message {
	return NewMessage().
		WithIcon(styles.IconSkull).
		WithKind("Output").
		WithMessage(message)
}

func write(file *os.File, line Line) error {
	if _, err := file.Write(append(line.Bytes(), '\n')); err != nil {
		return fmt.Errorf("Error writing to file: %w", err)
	}
	return nil
}
