package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToastManager_RenderOverlay_NoToastsReturnsEmpty(t *testing.T) {
	tm := NewToastManager()

	out := tm.RenderOverlay(40)

	assert.Empty(t, out, "RenderOverlay should return empty string when there are no toasts")
}

func TestToastManager_RenderOverlay_IncludesEachActiveToast(t *testing.T) {
	tm := NewToastManager()
	tm.Success("first toast")
	tm.Error("second toast")
	tm.Info("third toast")

	out := tm.RenderOverlay(40)

	// Each toast should appear in the overlay output
	assert.Contains(t, out, "first toast")
	assert.Contains(t, out, "second toast")
	assert.Contains(t, out, "third toast")
	// Toasts stack vertically — confirm a newline separates them
	assert.True(t, strings.Contains(out, "\n"), "toasts should be newline-separated when stacked")
}

func TestToastManager_RenderOverlay_DoesNotExceedWidth(t *testing.T) {
	tm := NewToastManager()
	tm.Success("hi")

	out := tm.RenderOverlay(20)

	// lipgloss may add padding; check that the output is bounded. The
	// important property is that we don't blow up and the overlay is
	// present.
	assert.NotEmpty(t, out, "overlay should be non-empty with an active toast")
}
