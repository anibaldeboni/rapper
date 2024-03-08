package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/anibaldeboni/rapper/cli/log"
	"github.com/anibaldeboni/rapper/cli/ui"
)

type Output interface {
	Close()
	Enabled() bool
	Send(OutputMessage)
	Listen()
}

type output struct {
	ch       chan OutputMessage
	filePath string
	logs     *log.Logs
}

type OutputMessage struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
	Error  error  `json:"error"`
	Body   []byte `json:"body"`
}

func NewMessage(url string, status int, err error, body []byte) OutputMessage {
	return OutputMessage{URL: url, Status: status, Error: err, Body: body}
}

func New(filePath string, logs *log.Logs) Output {
	return &output{filePath: filePath, ch: make(chan OutputMessage), logs: logs}
}

func (o *output) Close() {
	if o.ch != nil {
		close(o.ch)
	}
}

func (o *output) Enabled() bool {
	return o.filePath != ""
}

func (o *output) Send(log OutputMessage) {
	if o.Enabled() {
		o.ch <- log
	}
}

func (o *output) Listen() {
	if !o.Enabled() {
		return
	}
	file, err := os.OpenFile(o.filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		o.logs.Add(&outputMessage{message: err.Error()})
	}
	defer file.Close()
	for log := range o.ch {
		m, _ := json.Marshal(log)
		if _, err := file.Write(append(m, '\n')); err != nil {
			o.logs.Add(&outputMessage{message: err.Error()})
		}
	}
}

type outputMessage struct {
	message string
}

func (o *outputMessage) String() string {
	return fmt.Sprintf("%s [%s] %s", ui.IconSkull, ui.Bold("Output"), o.message)
}
