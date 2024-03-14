package processor

import (
	"fmt"
	"slices"

	"github.com/anibaldeboni/rapper/internal/execlog"
	"github.com/anibaldeboni/rapper/internal/styles"
	"golang.org/x/exp/maps"
)

func cancelationMsg() execlog.Message {
	return execlog.NewMessage().
		WithIcon(styles.IconSkull).
		WithKind("Cancelation").
		WithMessage(fmt.Sprintf("Read %d lines and executed %d requests", linesCount.Load(), reqCount.Load()))
}

func requestError(message string) execlog.Message {
	return execlog.NewMessage().
		WithIcon(styles.IconSkull).
		WithKind("Request").
		WithMessage(message)
}

func csvError(message string) execlog.Message {
	return execlog.NewMessage().
		WithIcon(styles.IconSkull).
		WithKind("CSV").
		WithMessage(message)
}

func mapResponse(record map[string]string, status int) string {
	var result string
	keys := maps.Keys(record)
	slices.Sort(keys)
	for _, key := range keys {
		result += fmt.Sprintf("%s: %s ", styles.Bold(key), record[key])
	}
	result += fmt.Sprintf("status: %s", styles.Pink(fmt.Sprint(status)))

	return result
}
func httpStatusError(record map[string]string, status int) execlog.Message {
	return execlog.NewMessage().
		WithIcon(styles.IconWarning).
		WithMessage(mapResponse(record, status))
}

func doneMessage(errs uint64) execlog.Message {
	var errMsg string
	var icon string
	if errs > 0 {
		errMsg = fmt.Sprintf("%s errors", styles.Pink(fmt.Sprint(errs)))
		icon = styles.IconError
	} else {
		errMsg = styles.Green("no errors")
		icon = styles.IconTrophy
	}
	return execlog.NewMessage().
		WithIcon(icon).
		WithMessage(fmt.Sprintf("Finished with %s\n", errMsg))

}

func processingMessage(file string, workers int) execlog.Message {
	return execlog.NewMessage().
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
