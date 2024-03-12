package messages

import (
	"fmt"
	"slices"

	"github.com/anibaldeboni/rapper/internal/styles"
	"golang.org/x/exp/maps"
)

type requestError struct {
	message string
}

func (e *requestError) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("Request"), e.message)
}

func NewRequestError(message string) *requestError {
	return &requestError{message: message}
}

type statusError struct {
	message string
}

func (e *statusError) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("Status"), e.message)
}

func NewStatusError(message string) *statusError {
	return &statusError{message: message}
}

type csvError struct {
	message string
}

func (e *csvError) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("CSV"), e.message)
}

func NewCsvError(message string) *csvError {
	return &csvError{message: message}
}

type cancelationError struct {
	message string
}

func (e *cancelationError) String() string {
	return fmt.Sprintf("%s [%s] %s", styles.IconSkull, styles.Bold("Cancelation"), e.message)
}
func NewCancelationError(message string) *cancelationError {
	return &cancelationError{message: message}
}

type httpStatusError struct {
	record map[string]string
	status int
}

func (e *httpStatusError) String() string {
	var result string
	keys := maps.Keys(e.record)
	slices.Sort(keys)
	for _, key := range keys {
		result += fmt.Sprintf("%s: %s ", styles.Bold(key), e.record[key])
	}
	result += fmt.Sprintf("status: %s", styles.Pink(fmt.Sprint(e.status)))

	return result
}
func NewHttpStatusError(record map[string]string, status int) *httpStatusError {
	return &httpStatusError{record: record, status: status}
}

type doneMessage struct {
	errs  uint64
	lines uint64
}

func (d *doneMessage) String() string {
	var errMsg string
	var icon string
	if d.errs > 0 {
		errMsg = fmt.Sprintf("%s errors", styles.Pink(fmt.Sprint(d.errs)))
		icon = styles.IconError
	} else {
		errMsg = styles.Green("no errors")
		icon = styles.IconTrophy
	}

	return fmt.Sprintf("%s Read %d lines and got %s\n", icon, d.lines, errMsg)
}
func NewDoneMessage(errs uint64, lines uint64) *doneMessage {
	return &doneMessage{errs: errs, lines: lines}
}

type operationError struct{}

func (e *operationError) String() string {
	return fmt.Sprintf("\n%s  %s\n", styles.IconInformation, "Please wait the current operation to finish or cancel pressing ESC")
}

func NewOperationError() *operationError {
	return &operationError{}
}

type processingMessage struct {
	file string
}

func (p *processingMessage) String() string {
	return fmt.Sprintf("%s Processing file %s", styles.IconWomanDancing, styles.Green(p.file))
}

func NewProcessingMessage(file string) *processingMessage {
	return &processingMessage{file: file}
}
