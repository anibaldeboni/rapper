package components

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToastManager_Layers_NoToastsReturnsNil covers the nil/empty case
// for the new Layers() factory: with no active toasts the manager must
// return nil so the caller can skip the compositor entirely.
func TestToastManager_Layers_NoToastsReturnsNil(t *testing.T) {
	tm := NewToastManager()

	layers := tm.Layers(120)

	assert.Nil(t, layers, "Layers must return nil when there are no active toasts")
}

// TestToastManager_Layers_NonPositiveWidthReturnsNil covers the terminal
// edge cases: when terminalWidth is 0 or negative the function must
// short-circuit and return nil instead of producing layers with
// negative coordinates.
func TestToastManager_Layers_NonPositiveWidthReturnsNil(t *testing.T) {
	tm := NewToastManager()
	tm.Success("hi")

	for _, width := range []int{0, -1} {
		t.Run("width="+itoa(width), func(t *testing.T) {
			layers := tm.Layers(width)
			assert.Nil(t, layers, "Layers must return nil for non-positive terminal width (got width=%d)", width)
		})
	}
}

// TestToastManager_Layers_SingleToastPositionedTopRight covers the
// happy path for a single active toast: X is anchored to the right edge
// (terminalWidth - toastVisualWidth - 2), Y starts at 1 (skip header),
// Z is 1 (above the bg layer which uses Z=0).
func TestToastManager_Layers_SingleToastPositionedTopRight(t *testing.T) {
	tm := NewToastManager()
	tm.Success("saved")

	layers := tm.Layers(120)

	require.Len(t, layers, 1, "Layers must return exactly one layer for one active toast")
	layer := layers[0]

	// Visible content is constrained to toastOverlayWidth (40) columns.
	assert.Equal(t, toastOverlayWidth, layer.Width(),
		"layer rendered width must equal the reserved toast column footprint")
	assert.Equal(t, toastOverlayWidth, lipgloss.Width(layer.GetContent()),
		"layer content visible width must equal the reserved toast column footprint")

	// X = terminalWidth - width(content) - 2 = 120 - 40 - 2 = 78.
	assert.Equal(t, 78, layer.GetX(),
		"X must be anchored to the right edge with a 2-col margin")

	// Y starts at 1 to skip the 1-line global header.
	assert.Equal(t, 1, layer.GetY(), "Y must start at 1 (below the header)")

	// Z is 1 so the toast sits above the bg layer (Z=0).
	assert.Equal(t, 1, layer.GetZ(), "Z must be 1 (above the bg layer)")
}

// TestToastManager_Layers_SingleToastContentIncludesIconAndMessage
// locks in the contract that the toast icon and message text are
// preserved inside the layer's rendered content.
func TestToastManager_Layers_SingleToastContentIncludesIconAndMessage(t *testing.T) {
	tm := NewToastManager()
	tm.Success("saved")

	layers := tm.Layers(120)
	require.Len(t, layers, 1)

	content := layers[0].GetContent()
	assert.Contains(t, content, "✓", "toast content must include the success icon")
	assert.Contains(t, content, "saved", "toast content must include the message text")
}

// TestToastManager_Layers_MultipleToastsStackByHeight covers the
// stacking invariant: the i+1-th toast must start at Y = Y[i] +
// layer[i].Height() so multi-line toasts (Padding(1, 2) = 3 lines)
// stack without overlap. This test is driven by Layer.Height() rather
// than a hardcoded "+3" so it stays correct if a future toast style
// changes the per-toast height.
func TestToastManager_Layers_MultipleToastsStackByHeight(t *testing.T) {
	tm := NewToastManager()
	tm.Success("a")
	tm.Error("b")
	tm.Info("c")

	layers := tm.Layers(120)
	require.Len(t, layers, 3, "Layers must return one layer per active toast")

	// All toast layers share the same X anchor and Z.
	for i, l := range layers {
		assert.Equal(t, 78, l.GetX(), "toast %d X must match the right-edge anchor", i)
		assert.Equal(t, 1, l.GetZ(), "toast %d Z must be 1", i)
	}

	// Stacking: Y[i+1] = Y[i] + layer[i].Height().
	expectedY := 1
	for i, l := range layers {
		assert.Equal(t, expectedY, l.GetY(),
			"toast %d Y must equal previous-Y + previous-height; got %d, want %d",
			i, l.GetY(), expectedY)
		expectedY += l.Height()
	}
}

// TestToastManager_Layers_NarrowTerminalFallsBackToXZero covers the
// edge case where the terminal is narrower than the toast footprint.
// A negative targetX would otherwise underflow; the implementation
// must clamp to 0 so the toast remains visible at the left edge.
func TestToastManager_Layers_NarrowTerminalFallsBackToXZero(t *testing.T) {
	tm := NewToastManager()
	tm.Success("hi")

	// 30-col terminal is narrower than 40-col toast + 2-col margin, so
	// 30 - 40 - 2 = -12 would be negative without the clamp.
	layers := tm.Layers(30)
	require.Len(t, layers, 1)

	assert.Equal(t, 0, layers[0].GetX(),
		"X must be clamped to 0 when terminalWidth < toastOverlayWidth + 2")
}

// TestToastManager_Layers_ExpiredToastExcluded covers the expiry
// path: a toast whose lifetime is past Duration must not appear in
// the result. Update() is the public mechanism for expiring toasts;
// we call it before Layers() and verify the slice is nil.
func TestToastManager_Layers_ExpiredToastExcluded(t *testing.T) {
	tm := NewToastManager()
	// Inject an already-expired toast (created 5s ago with a 4s
	// duration) directly via the internal Add path. We use the public
	// Add() and then mutate CreatedAt backwards to simulate expiry
	// without sleeping.
	tm.Success("expired-soon")
	require.Len(t, tm.GetActive(), 1, "sanity: toast is active before expiry")

	// Rewind CreatedAt past Duration so Update() drops it.
	tm.toasts[0].CreatedAt = time.Now().Add(-10 * time.Second)
	tm.Update()

	assert.Empty(t, tm.GetActive(), "Update() must drop expired toasts")
	layers := tm.Layers(120)
	assert.Nil(t, layers, "Layers must return nil after the only active toast has expired")
}

// itoa is a tiny helper used by the table-driven non-positive-width
// test. We avoid strconv here to keep the test file free of an
// otherwise-unused import.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
