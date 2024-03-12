package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/anibaldeboni/rapper/internal/log"
	"github.com/anibaldeboni/rapper/internal/styles"
)

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

func (o *message) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("Output"), o.message)
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

func (o *streamImpl) Close() {
	if o.ch != nil {
		close(o.ch)
	}
}

func (o *streamImpl) Enabled() bool {
	return o.filePath != ""
}

// Send sends the given log message to the output channel if the output is enabled.
func (o *streamImpl) Send(log Message) {
	if o.Enabled() {
		o.ch <- log
	}
}

func listen(o *streamImpl) {
	if !o.Enabled() {
		return
	}
	file, err := os.OpenFile(o.filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		o.logs.Add(&message{message: err.Error()})
	}
	defer file.Close()
	for log := range o.ch {
		m, _ := json.Marshal(log)
		if _, err := file.Write(append(m, '\n')); err != nil {
			o.logs.Add(&message{message: err.Error()})
		}
	}
}
