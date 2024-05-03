package ui

import (
	"path/filepath"

	"github.com/anibaldeboni/rapper/internal/styles"
)

func mapListOptions(filePaths []string) []Option[string] {
	opts := make([]Option[string], 0, len(filePaths))
	maxWidth := int(float64(styles.TerminalWidth()) * 0.18)
	for _, filePath := range filePaths {
		opts = append(
			opts,
			Option[string]{
				Title: trimFilename(filePath, maxWidth),
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
