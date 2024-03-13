package messages

import (
	"fmt"
	"slices"

	"github.com/anibaldeboni/rapper/internal/log"
	"github.com/anibaldeboni/rapper/internal/styles"
	"golang.org/x/exp/maps"
)

var (
	_ log.LogMessage = (*requestError)(nil)
	_ log.LogMessage = (*statusError)(nil)
	_ log.LogMessage = (*csvError)(nil)
	_ log.LogMessage = (*cancelationError)(nil)
	_ log.LogMessage = (*httpStatusError)(nil)
	_ log.LogMessage = (*doneMessage)(nil)
	_ log.LogMessage = (*operationError)(nil)
	_ log.LogMessage = (*processingMessage)(nil)
	_ log.LogMessage = (*genericMessage)(nil)
)

type requestError struct {
	message string
}

func (this *requestError) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("Request"), this.message)
}

func NewRequestError(message string) *requestError {
	return &requestError{message: message}
}

type statusError struct {
	message string
}

func (this *statusError) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("Status"), this.message)
}

func NewStatusError(message string) *statusError {
	return &statusError{message: message}
}

type csvError struct {
	message string
}

func (this *csvError) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("CSV"), this.message)
}

func NewCsvError(message string) *csvError {
	return &csvError{message: message}
}

type cancelationError struct {
	message string
}

func (this *cancelationError) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("Cancelation"), this.message)
}
func NewCancelationError(message string) *cancelationError {
	return &cancelationError{message: message}
}

type httpStatusError struct {
	record map[string]string
	status int
}

func (this *httpStatusError) String() string {
	var result string
	keys := maps.Keys(this.record)
	slices.Sort(keys)
	for _, key := range keys {
		result += fmt.Sprintf("%s: %s ", styles.Bold(key), this.record[key])
	}
	result += fmt.Sprintf("status: %s", styles.Pink(fmt.Sprint(this.status)))

	return result
}
func NewHttpStatusError(record map[string]string, status int) *httpStatusError {
	return &httpStatusError{record: record, status: status}
}

type doneMessage struct {
	errs uint64
}

func (this *doneMessage) String() string {
	var errMsg string
	var icon string
	if this.errs > 0 {
		errMsg = fmt.Sprintf("%s errors", styles.Pink(fmt.Sprint(this.errs)))
		icon = styles.IconError
	} else {
		errMsg = styles.Green("no errors")
		icon = styles.IconTrophy
	}

	return fmt.Sprintf("%s Finished with %s\n", icon, errMsg)
}
func NewDoneMessage(errs uint64) *doneMessage {
	return &doneMessage{errs: errs}
}

type operationError struct{}

func (this *operationError) String() string {
	return fmt.Sprintf("\n%s  %s\n", styles.IconInformation, "Please wait the current operation to finish or cancel pressing ESC")
}

func NewOperationError() *operationError {
	return &operationError{}
}

type processingMessage struct {
	file string
}

func (this *processingMessage) String() string {
	return fmt.Sprintf("%s Processing file %s", styles.IconWomanDancing, styles.Green(this.file))
}

func NewProcessingMessage(file string) *processingMessage {
	return &processingMessage{file: file}
}

type genericMessage struct {
	message string
	kind    string
	icon    string
}

func (this *genericMessage) String() string {
	return fmt.Sprintf("%s [%s] %s", this.icon, styles.Bold(this.kind), this.message)
}
func (this *genericMessage) WithIcon(icon string) *genericMessage {
	this.icon = icon
	return this
}
func (this *genericMessage) WithKind(kind string) *genericMessage {
	this.kind = kind
	return this
}
func (this *genericMessage) WithMessage(message string) *genericMessage {
	this.message = message
	return this
}
func NewGenericMessage() *genericMessage {
	return &genericMessage{icon: "🤷‍♂️"}
}
