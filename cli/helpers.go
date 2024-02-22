package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/anibaldeboni/rapper/cli/ui"
	"github.com/anibaldeboni/rapper/files"
	"golang.org/x/exp/maps"
)

type Error int

const (
	Request Error = iota
	Status
	CSV
	Cancelation
)

func fmtError(kind Error, err string) string {
	f := func(kind string, err string) string {
		return fmt.Sprintf("%s [%s] %s", ui.IconSkull, ui.Bold(kind), err)
	}
	switch kind {
	case Request:
		return f("Request", err)
	case Status:
		return f("Status", err)
	case CSV:
		return f("CSV", err)
	case Cancelation:
		return f("Cancelation", err)
	default:
		return f("Unknown", err)
	}
}

func fmtStatusError(record map[string]string, status int) string {
	var result string
	keys := maps.Keys(record)
	sort.Strings(keys)
	for _, key := range keys {
		result += fmt.Sprintf("%s: %s ", ui.Bold(key), record[key])
	}
	result += fmt.Sprintf("status: %s", ui.Pink(fmt.Sprint(status)))

	return result
}

func formatDoneMessage(errs int) string {
	if errs > 0 {
		return fmt.Sprintf("%s Processing finished with %s errors\n", ui.IconError, ui.Pink(fmt.Sprint(errs)))
	}
	return fmt.Sprintf("%s Processing finished with %s\n", ui.IconTrophy, ui.Green("no errors"))
}

func findCsv(path string) ([]Option[string], error) {
	filePaths, err := files.FindFiles(path, "*.csv")
	if len(err) > 0 {
		return nil, fmt.Errorf("Could not execute file scan in %s", ui.Bold(path))
	}
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("No CSV files found in %s", ui.Bold(path))
	}

	opts := make([]Option[string], 0, len(filePaths))
	for _, filePath := range filePaths {
		opts = append(
			opts,
			Option[string]{
				Title: filepath.Base(filePath),
				Value: filePath,
			},
		)
	}

	return opts, nil
}

func Exit(message any, arg ...any) {
	switch message := message.(type) {
	case string:
		if len(message) == 0 {
			os.Exit(0)
		}
		fmt.Println(ui.QuitTextStyle(fmt.Sprintf(message, arg...)))
		os.Exit(0)
	case error:
		fmt.Println(ui.QuitTextStyle(fmt.Sprintf(message.Error()+"\n", arg...)))
		os.Exit(1)
	case nil:
		os.Exit(0)
	default:
		fmt.Println(ui.QuitTextStyle(fmt.Sprintf("%v\n", message)))
		os.Exit(1)
	}
}
