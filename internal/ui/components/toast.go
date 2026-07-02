package components

import (
	"time"

	"charm.land/lipgloss/v2"
)

// ToastType defines the type of toast notification
type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
	ToastWarning
)

// toastOverlayWidth is the width reserved for each toast layer in
// columns. The metric panel in the Logs view uses ~36 columns by
// default, so 40 leaves a small visual gap between the two top-right
// elements. Kept package-private: the dimension is an internal detail
// of the toast rendering pipeline and is not part of any public
// contract.
const (
	toastOverlayWidth = 40

	// toastOverlayRightMargin is the number of blank columns kept between
	// the right edge of each toast and the right edge of the terminal.
	// Mirrors the 2-col margin that the previous spliceRight implementation
	// enforced so the visual stays identical.
	toastOverlayRightMargin = 2

	// toastOverlayHeaderOffset is the Y offset (in rows) of the first toast
	// from the top of the bg layer. Set to 1 to skip the global header
	// line, so toasts always land in the content area, never on top of the
	// navigation help bar.
	toastOverlayHeaderOffset = 1

	// toastOverlayZ is the Z-index used for every toast layer. The bg
	// layer is rendered at Z=0; toasts sit at Z=1 so they always appear
	// above the background content. Z is per-layer in
	// Compositor.flattenRecursive (it is NOT inherited parent→child), so
	// each toast sets this value explicitly.
	toastOverlayZ = 1
)

// Toast represents a temporary notification message
type Toast struct {
	Message   string
	Type      ToastType
	CreatedAt time.Time
	Duration  time.Duration
}

// isFading returns true when the toast is in the last 25% of its lifetime.
func (t Toast) isFading() bool {
	remaining := t.Duration - time.Since(t.CreatedAt)
	return remaining < t.Duration/4
}

// ToastManager manages a queue of toast notifications
type ToastManager struct {
	toasts    []Toast
	maxToasts int
}

// NewToastManager creates a new ToastManager
func NewToastManager() *ToastManager {
	return &ToastManager{
		toasts:    make([]Toast, 0),
		maxToasts: 5,
	}
}

// Add adds a new toast notification
func (tm *ToastManager) Add(message string, toastType ToastType) {
	toast := Toast{
		Message:   message,
		Type:      toastType,
		CreatedAt: time.Now(),
		Duration:  4 * time.Second,
	}

	tm.toasts = append(tm.toasts, toast)

	// Keep only the most recent toasts
	if len(tm.toasts) > tm.maxToasts {
		tm.toasts = tm.toasts[len(tm.toasts)-tm.maxToasts:]
	}
}

// Success adds a success toast
func (tm *ToastManager) Success(message string) {
	tm.Add(message, ToastSuccess)
}

// Error adds an error toast
func (tm *ToastManager) Error(message string) {
	tm.Add(message, ToastError)
}

// Info adds an info toast
func (tm *ToastManager) Info(message string) {
	tm.Add(message, ToastInfo)
}

// Warning adds a warning toast
func (tm *ToastManager) Warning(message string) {
	tm.Add(message, ToastWarning)
}

// Update removes expired toasts
func (tm *ToastManager) Update() {
	now := time.Now()
	activeToasts := make([]Toast, 0)

	for _, toast := range tm.toasts {
		if now.Sub(toast.CreatedAt) < toast.Duration {
			activeToasts = append(activeToasts, toast)
		}
	}

	tm.toasts = activeToasts
}

// GetActive returns all active toasts
func (tm *ToastManager) GetActive() []Toast {
	return tm.toasts
}

// HasActive returns true if there are active toasts
func (tm *ToastManager) HasActive() bool {
	return len(tm.toasts) > 0
}

// Layers returns one positioned *lipgloss.Layer per active toast,
// ready to be added to a lipgloss.Compositor alongside a background
// layer. Each toast is anchored to the right edge of the terminal with
// a 2-column margin and stacked vertically below the global header
// (Y starts at 1). The Z-index is 1, above the bg layer's Z=0.
//
// Returns nil when there are no active toasts or when terminalWidth is
// non-positive — the caller can use this to skip the compositor
// entirely on the no-toast path.
//
// Stacking uses Layer.Height() (not a hardcoded "+1") so multi-line
// toasts (currently 3 rows tall due to Padding(1, 2)) stack without
// overlap, and future style changes that change the per-toast height
// are picked up automatically.
func (tm *ToastManager) Layers(terminalWidth int) []*lipgloss.Layer {
	if !tm.HasActive() || terminalWidth <= 0 {
		return nil
	}

	layers := make([]*lipgloss.Layer, 0, len(tm.toasts))
	yOffset := toastOverlayHeaderOffset
	for _, toast := range tm.toasts {
		rs := lipgloss.NewStyle().Width(toastOverlayWidth).Align(lipgloss.Left)
		if toast.isFading() {
			rs = rs.Faint(true)
		}
		content := rs.Render(renderToast(toast))
		x := max(terminalWidth-lipgloss.Width(content)-toastOverlayRightMargin, 0)
		layer := lipgloss.NewLayer(content).X(x).Y(yOffset).Z(toastOverlayZ)
		layers = append(layers, layer)
		yOffset += layer.Height()
	}
	return layers
}

// renderToast renders a single toast notification
func renderToast(toast Toast) string {
	var icon string
	var style lipgloss.Style

	baseStyle := lipgloss.NewStyle().
		Padding(1, 2)

	switch toast.Type {
	case ToastSuccess:
		icon = "✓"
		style = baseStyle.
			Background(lipgloss.Color("40")).
			Foreground(lipgloss.Color("0")).
			Bold(true)

	case ToastError:
		icon = "✗"
		style = baseStyle.
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("15")).
			Bold(true)

	case ToastWarning:
		icon = "⚠"
		style = baseStyle.
			Background(lipgloss.Color("214")).
			Foreground(lipgloss.Color("0")).
			Bold(true)

	case ToastInfo:
		icon = "ℹ"
		style = baseStyle.
			Background(lipgloss.Color("39")).
			Foreground(lipgloss.Color("15")).
			Bold(true)
	}

	message := icon + " " + toast.Message

	return style.Render(message)
}
