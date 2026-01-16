package ui

import (
	"path/filepath"

	"github.com/anibaldeboni/rapper/internal/ui/views"
	"golang.org/x/term"
)

func mapFileToOption(filePath string) views.Option[string] {
	maxWidth := int(float64(terminalWidth()) * 0.18)
	return views.Option[string]{
		Title: trimFilename(filePath, maxWidth),
		Value: filePath,
	}
}

func trimFilename(filename string, length int) string {
	f := filepath.Base(filename)
	if len(f) < length {
		return f
	}
	return f[:length] + "..."
}

func terminalWidth() int {
	width, _, err := term.GetSize(0)
	if err != nil {
		width = 80
	}
	return width
}
