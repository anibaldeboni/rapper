package messages

import (
	"fmt"
	"slices"

	"github.com/anibaldeboni/rapper/cli/ui"
	"golang.org/x/exp/maps"
)

type requestError struct {
	message string
}

func (e *requestError) String() string {
	return fmt.Sprintf("%s [%s] %s", ui.IconSkull, ui.Bold("Request"), e.message)
}

func NewRequestError(message string) *requestError {
	return &requestError{message: message}
}

type statusError struct {
	message string
}

func (e *statusError) String() string {
	return fmt.Sprintf("%s [%s] %s", ui.IconSkull, ui.Bold("Status"), e.message)
}

func NewStatusError(message string) *statusError {
	return &statusError{message: message}
}

type csvError struct {
	message string
}

func (e *csvError) String() string {
	return fmt.Sprintf("%s [%s] %s", ui.IconSkull, ui.Bold("CSV"), e.message)
}

func NewCsvError(message string) *csvError {
	return &csvError{message: message}
}

type cancelationError struct {
	message string
}

func (e *cancelationError) String() string {
	return fmt.Sprintf("%s [%s] %s", ui.IconSkull, ui.Bold("Cancelation"), e.message)
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
		result += fmt.Sprintf("%s: %s ", ui.Bold(key), e.record[key])
	}
	result += fmt.Sprintf("status: %s", ui.Pink(fmt.Sprint(e.status)))

	return result
}
func NewHttpStatusError(record map[string]string, status int) *httpStatusError {
	return &httpStatusError{record: record, status: status}
}

type doneMessage struct {
	errs int
}

func (d *doneMessage) String() string {
	var errMsg string
	var icon string
	if d.errs > 0 {
		errMsg = fmt.Sprintf("%s errors", ui.Pink(fmt.Sprint(d.errs)))
		icon = ui.IconError
	} else {
		errMsg = ui.Green("no errors")
		icon = ui.IconTrophy
	}

	return fmt.Sprintf("%s Finished with %s\n", icon, errMsg)
}
func NewDoneMessage(errs int) *doneMessage {
	return &doneMessage{errs: errs}
}

type operationError struct{}

func (e *operationError) String() string {
	return fmt.Sprintf("\n%s  %s\n", ui.IconInformation, "Please wait the current operation to finish or cancel pressing ESC")
}

func NewOperationError() *operationError {
	return &operationError{}
}

type processingMessage struct {
	file string
}

func (p *processingMessage) String() string {
	return fmt.Sprintf("%s Processing file %s", ui.IconWomanDancing, ui.Green(p.file))
}

func NewProcessingMessage(file string) *processingMessage {
	return &processingMessage{file: file}
}
