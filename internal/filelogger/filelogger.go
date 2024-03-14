package filelogger

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/anibaldeboni/rapper/internal/execlog"
	"github.com/anibaldeboni/rapper/internal/styles"
)

var _ FileLogger = (*fileLoggerImpl)(nil)

//go:generate mockgen -destination mock/output_mock.go github.com/anibaldeboni/rapper/internal/filelogger Stream
type FileLogger interface {
	Write(Line)
}

type fileLoggerImpl struct {
	mu         sync.Mutex
	logManager execlog.Manager
	file       *os.File
}

type Line struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
	Error  error  `json:"error"`
	Body   []byte `json:"body"`
}

// NewLine creates a new Line struct with the provided URL, status code, error, and body.
func NewLine(url string, status int, err error, body []byte) Line {
	return Line{URL: url, Status: status, Error: err, Body: body}
}

// New creates a new FileLogger instance with the specified file path and log manager.
// It returns the created FileLogger.
func New(filePath string, logManager execlog.Manager) FileLogger {
	var err error
	fileLogger := &fileLoggerImpl{logManager: logManager}

	if filePath == "" {
		return fileLogger
	}

	if fileLogger.file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660); err != nil {
		logManager.Add(errorMessage(err.Error()))
	}

	return fileLogger
}

func enabled(this *fileLoggerImpl) bool {
	return this.file != nil
}

// Write writes a line to the file logger.
// If the logger is enabled, it acquires a lock, writes the line to the file,
// and adds an error message to the log manager if there was an error writing to the file.
func (this *fileLoggerImpl) Write(line Line) {
	if enabled(this) {
		this.mu.Lock()
		defer this.mu.Unlock()
		if err := write(this.file, line); err != nil {
			this.logManager.Add(errorMessage(err.Error()))
		}
	}
}

func errorMessage(message string) execlog.Message {
	return execlog.NewMessage().
		WithIcon(styles.IconSkull).
		WithKind("Output").
		WithMessage(message)
}

func write(file *os.File, line Line) error {
	m, _ := json.Marshal(line)
	if _, err := file.Write(append(m, '\n')); err != nil {
		return err
	}
	return nil
}
