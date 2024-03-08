package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anibaldeboni/rapper/cli/ui"
	"github.com/anibaldeboni/rapper/files"
)

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
				Title: trimFilename(filePath, 17),
				Value: filePath,
			},
		)
	}

	return opts, nil
}
func trimFilename(filename string, length int) string {
	f := filepath.Base(filename)
	if len(f) < length {
		return f
	}
	return f[:length] + "..."

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
