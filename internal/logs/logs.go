package logs

import (
	"fmt"
	"os"
	"sync"

	"github.com/anibaldeboni/rapper/internal/styles"
)

// Line is the contract for anything that can be written to the output
// file. The processor's RequestLine is the canonical implementation;
// the abstraction exists so the logger does not import the processor
// package (which would create a cycle once the processor imports
// logs for the Add(LogMessage) call).
type Line interface {
	Bytes() []byte
}

// Logger is the in-memory + on-disk log sink. Processors call
// logger.Add(LogMessage) for every event they want to surface; the
// TUI polls logger.Get() to render the list. WriteToFile streams
// per-request records (RequestLine) to the on-disk log file.
type logger struct {
	file *os.File
	sync.RWMutex
	messages []LogMessage
}

// NewLogger creates a new Logger instance.
//
// If filePath is non-empty the file is opened with
// O_WRONLY|O_APPEND|O_CREATE and perms 0660, matching the
// pre-refactor behavior. A failure to open the file is captured as a
// LogMessage in the in-memory buffer so the user sees the error in
// the TUI.
func NewLogger(filePath string) *logger {
	var l logger
	if filePath != "" {
		if file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660); err != nil {
			l.Add(errorMessage(err.Error()))
		} else {
			l.file = file
		}
	}
	return &l
}

// Add appends a log message to the in-memory buffer.
func (l *logger) Add(log LogMessage) {
	l.Lock()
	defer l.Unlock()
	l.messages = append(l.messages, log)
}

// Get returns a snapshot of all log messages. The returned slice is
// a copy — callers can mutate it freely without affecting the
// internal state.
func (l *logger) Get() []LogMessage {
	l.RLock()
	defer l.RUnlock()
	out := make([]LogMessage, len(l.messages))
	copy(out, l.messages)
	return out
}

// Clear empties the in-memory log buffer. Called on ProcessingStartedMsg
// so each run starts from a clean slate.
func (l *logger) Clear() {
	l.Lock()
	defer l.Unlock()
	l.messages = nil
}

// WriteToFile writes the given log line to the log file, if it is open.
// If writing fails, the error is captured as a LogMessage in the
// in-memory buffer.
func (l *logger) WriteToFile(line Line) {
	if l.file != nil {
		if err := write(l.file, line); err != nil {
			l.Add(errorMessage(err.Error()))
		}
	}
}

func errorMessage(message string) LogMessage {
	return NewGeneralMessage(styles.IconSkull, "Output", message)
}

func write(file *os.File, line Line) error {
	if _, err := file.Write(append(line.Bytes(), '\n')); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}
