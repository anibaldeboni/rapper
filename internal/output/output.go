package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/anibaldeboni/rapper/internal/log"
	"github.com/anibaldeboni/rapper/internal/styles"
)

type Stream interface {
	Close()
	Enabled() bool
	Send(Message)
}

type streamIpml struct {
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

func NewMessage(url string, status int, err error, body []byte) Message {
	return Message{URL: url, Status: status, Error: err, Body: body}
}

func New(filePath string, logs log.LogManager) Stream {
	o := &streamIpml{filePath: filePath, ch: make(chan Message), logs: logs}
	go listen(o)
	return o
}

func (o *streamIpml) Close() {
	if o.ch != nil {
		close(o.ch)
	}
}

func (o *streamIpml) Enabled() bool {
	return o.filePath != ""
}

func (o *streamIpml) Send(log Message) {
	if o.Enabled() {
		o.ch <- log
	}
}

func listen(o *streamIpml) {
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

type message struct {
	message string
}

func (o *message) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("Output"), o.message)
}
