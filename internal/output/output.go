package output

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/anibaldeboni/rapper/internal/log"
	"github.com/anibaldeboni/rapper/internal/styles"
)

var _ Stream = (*streamImpl)(nil)

//go:generate mockgen -destination mock/output_mock.go github.com/anibaldeboni/rapper/internal/output Stream
type Stream interface {
	Enabled() bool
	Send(Line)
}

type streamImpl struct {
	mu         sync.Mutex
	filePath   string
	logManager log.Manager
}

type Line struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
	Error  error  `json:"error"`
	Body   []byte `json:"body"`
}

// NewLine creates a new output line with the specified URL, status code, error, and body.
func NewLine(url string, status int, err error, body []byte) Line {
	return Line{URL: url, Status: status, Error: err, Body: body}
}

// New creates a new Stream instance with the specified file path and log manager.
// It returns the created Stream.
func New(filePath string, logManager log.Manager) Stream {
	return &streamImpl{
		filePath:   filePath,
		logManager: logManager,
	}
}

func (this *streamImpl) Enabled() bool {
	return this.filePath != ""
}

// Send sends the given log message to the output channel if the output is enabled.
func (this *streamImpl) Send(line Line) {
	if this.Enabled() {
		this.mu.Lock()
		defer this.mu.Unlock()
		if err := write(this.filePath, line); err != nil {
			this.logManager.Add(errorMessage(err.Error()))
		}
	}
}

func errorMessage(message string) log.Message {
	return log.NewMessage().
		WithIcon(styles.IconSkull).
		WithKind("Output").
		WithMessage(message)
}

func write(filePath string, line Line) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer file.Close()
	m, _ := json.Marshal(line)
	if _, err := file.Write(append(m, '\n')); err != nil {
		return err
	}
	return nil
}
