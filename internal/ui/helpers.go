package ui

import (
	"path/filepath"
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
