package processor

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/styles"
)

func csvError(message string) logs.LogMessage {
	return logs.NewMessage("CSV error", logs.WithDetail(message), logs.WithIcon(styles.IconSkull), logs.AsError())
}

func doneMessage(errs uint64) logs.LogMessage {
	errMsg := "no errors"
	icon := styles.IconTrophy

	if errs > 0 {
		errMsg = strconv.FormatUint(errs, 10) + " errors"
		icon = styles.IconError
	}

	return logs.NewMessage("Finished with "+errMsg, logs.WithIcon(icon), logs.AsGeneral())
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
