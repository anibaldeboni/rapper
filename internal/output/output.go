package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/anibaldeboni/rapper/internal/log"
	"github.com/anibaldeboni/rapper/internal/styles"
)

var _ Stream = (*streamImpl)(nil)

//go:generate mockgen -destination mock/output_mock.go github.com/anibaldeboni/rapper/internal/output Stream
type Stream interface {
	Close()
	Enabled() bool
	Send(Message)
}

type streamImpl struct {
	ch       chan Message
	filePath string
	logs     log.LogManager
}

type Message struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
	Error  error  `json:"error"`
	Body   []byte `json:"body"`
}

type message struct {
	message string
}

func (this *message) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("Output"), this.message)
}

// NewMessage creates a new Message with the specified URL, status code, error, and body.
func NewMessage(url string, status int, err error, body []byte) Message {
	return Message{URL: url, Status: status, Error: err, Body: body}
}

// New creates a new Stream instance with the specified file path and log manager.
// It returns the created Stream.
func New(filePath string, logs log.LogManager) Stream {
	o := &streamImpl{filePath: filePath, ch: make(chan Message), logs: logs}
	go listen(o)
	return o
}

func (this *streamImpl) Close() {
	if this.ch != nil {
		close(this.ch)
	}
}

func (this *streamImpl) Enabled() bool {
	return this.filePath != ""
}

// Send sends the given log message to the output channel if the output is enabled.
func (this *streamImpl) Send(log Message) {
	if this.Enabled() {
		this.ch <- log
	}
}

func listen(this *streamImpl) {
	if !this.Enabled() {
		return
	}
	file, err := os.OpenFile(this.filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		this.logs.Add(&message{message: err.Error()})
	}
	defer file.Close()
	for log := range this.ch {
		m, _ := json.Marshal(log)
		if _, err := file.Write(append(m, '\n')); err != nil {
			this.logs.Add(&message{message: err.Error()})
		}
	}
}
