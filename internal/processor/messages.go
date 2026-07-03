package processor

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/styles"
)

func cancelationMsg() logs.LogMessage {
	return logs.NewGeneralMessage(
		styles.IconSkull,
		"Cancelation",
		fmt.Sprintf("Read %d lines and executed %d requests", linesCount.Load(), reqCount.Load()),
	)
}

func requestError(message string) logs.LogMessage {
	return logs.NewGeneralMessage(styles.IconSkull, "Request", message)
}

func csvError(message string) logs.LogMessage {
	return logs.NewGeneralMessage(styles.IconSkull, "CSV", message)
}

func doneMessage(errs uint64) logs.LogMessage {
	errMsg := styles.Green("no errors")
	icon := styles.IconTrophy

	if errs > 0 {
		errMsg = styles.Pink(strconv.FormatUint(errs, 10)) + " errors"
		icon = styles.IconError
	}

	return logs.NewGeneralMessage(icon, "Done", "Finished with "+errMsg)
}

func processingMessage(file string, workers int) logs.LogMessage {
	return logs.NewGeneralMessage(
		styles.IconWomanDancing,
		"Processing",
		fmt.Sprintf("Processing file %s using %s", styles.Green(file), workersMsg(workers)),
	)
}

func workersMsg(workers int) string {
	w := "worker"
	if workers > 1 {
		w += "s"
	}
	return fmt.Sprintf("%d %s", workers, w)
}

// RequestLine is the per-request record streamed to the on-disk
// output file. The body is included even on errors so the user can
// inspect what the server actually said; the field is omitted from
// the JSON when nil to keep success-only output compact.
type RequestLine struct {
	Error  error  `json:"error"`
	URL    string `json:"url"`
	Method string `json:"method"`
	Body   []byte `json:"body"`
	Status int    `json:"status"`
}

// Bytes serialises the request line as JSON. Used by
// logger.WriteToFile to stream the line to the configured output
// file.
func (r RequestLine) Bytes() []byte {
	m, _ := json.Marshal(r)
	return m
}
