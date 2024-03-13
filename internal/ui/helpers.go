package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anibaldeboni/rapper/internal/styles"
)

func mapListOptions(filePaths []string) []Option[string] {
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

	return opts
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
		fmt.Println(styles.QuitTextStyle(fmt.Sprintf(message, arg...)))
		os.Exit(0)
	case error:
		fmt.Println(styles.QuitTextStyle(fmt.Sprintf(message.Error()+"\n", arg...)))
		os.Exit(1)
	case nil:
		os.Exit(0)
	default:
		fmt.Println(styles.QuitTextStyle(fmt.Sprintf("%v\n", message)))
		os.Exit(1)
	}
}
