package processor

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/styles"
	"golang.org/x/exp/maps"
)

func cancelationMsg() logs.Message {
	return logs.NewMessage().
		WithIcon(styles.IconSkull).
		WithKind("Cancelation").
		WithMessage(fmt.Sprintf("Read %d lines and executed %d requests", linesCount.Load(), reqCount.Load()))
}

func requestError(message string) logs.Message {
	return logs.NewMessage().
		WithIcon(styles.IconSkull).
		WithKind("Request").
		WithMessage(message)
}

func csvError(message string) logs.Message {
	return logs.NewMessage().
		WithIcon(styles.IconSkull).
		WithKind("CSV").
		WithMessage(message)
}

func mapResponse(record map[string]string, status int) string {
	keys := maps.Keys(record)
	slices.Sort(keys)

	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(styles.Bold(key))
		builder.WriteString(": ")
		builder.WriteString(record[key])
		builder.WriteString(" ")
	}
	builder.WriteString("status: ")
	builder.WriteString(styles.Pink(strconv.Itoa(status)))

	return builder.String()
}
func httpStatusError(record map[string]string, status int) logs.Message {
	return logs.NewMessage().
		WithIcon(styles.IconWarning).
		WithMessage(mapResponse(record, status))
}

func doneMessage(errs uint64) logs.Message {
	errMsg := styles.Green("no errors")
	icon := styles.IconTrophy

	if errs > 0 {
		errMsg = styles.Pink(strconv.FormatUint(errs, 10)) + " errors"
		icon = styles.IconError
	}

	return logs.NewMessage().
		WithIcon(icon).
		WithMessage(fmt.Sprintf("Finished with %s\n", errMsg))
}

func processingMessage(file string, workers int) logs.Message {
	return logs.NewMessage().
		WithIcon(styles.IconWomanDancing).
		WithMessage(fmt.Sprintf("Processing file %s using %s", styles.Green(file), workersMsg(workers)))
}

func workersMsg(workers int) string {
	w := "worker"
	if workers > 1 {
		w += "s"
	}
	return fmt.Sprintf("%d %s", workers, w)
}

type RequestLine struct {
	Error  error  `json:"error"`
	URL    string `json:"url"`
	Body   []byte `json:"body"`
	Status int    `json:"status"`
}

func (r RequestLine) Bytes() []byte {
	m, _ := json.Marshal(r)
	return m
}
