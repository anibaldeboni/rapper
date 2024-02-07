package cli

import (
	"fmt"
	"github.com/anibaldeboni/rapper/cli/ui"
	"github.com/anibaldeboni/rapper/files"
	"os"
	"path/filepath"
	"sort"
)

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

func formatError(name string, err string) string {
	return fmt.Sprintf("%s [%s] %s", ui.IconSkull, ui.Bold(name), err)
}

func formatStatusError(record map[string]string, status int) string {
	result := ui.IconWarning + "  "
	keys := make([]string, 0, len(record))
	for k := range record {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result += fmt.Sprintf("%s: %s ", ui.Bold(key), record[key])
	}
	result += fmt.Sprintf("status: %s", ui.Pink(fmt.Sprint(status)))

	return result
}

func findCsv(path string) ([]Option[string], error) {
	filePaths, err := files.FindFiles(path, "*.csv")
	if len(err) > 0 {
		return nil, fmt.Errorf("Could not execute file scan in %s", ui.Bold(path))
	}
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("No CSV files found in %s", ui.Bold(path))
	}

	opts := make([]Option[string], 0)
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
