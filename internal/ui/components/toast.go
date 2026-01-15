package components

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ToastType defines the type of toast notification
type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
	ToastWarning
)

// Toast represents a temporary notification message
type Toast struct {
	Message   string
	Type      ToastType
	CreatedAt time.Time
	Duration  time.Duration
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
		maxToasts: 3,
	}
}

// Add adds a new toast notification
func (tm *ToastManager) Add(message string, toastType ToastType) {
	toast := Toast{
		Message:   message,
		Type:      toastType,
		CreatedAt: time.Now(),
		Duration:  3 * time.Second,
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

// Render renders all active toasts
func (tm *ToastManager) Render() string {
	if !tm.HasActive() {
		return ""
	}

	var rendered strings.Builder

	for _, toast := range tm.toasts {
		rendered.WriteString(renderToast(toast))
	}

	return rendered.String()
}

// renderToast renders a single toast notification
func renderToast(toast Toast) string {
	var icon string
	var style lipgloss.Style

	baseStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(1).
		MarginTop(1)

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

	// Calculate remaining time percentage for fade effect
	elapsed := time.Since(toast.CreatedAt)
	remaining := toast.Duration - elapsed
	fadeThreshold := toast.Duration / 4 // Start fading in last 25%

	// Apply fade effect if nearing expiration
	if remaining < fadeThreshold {
		// Reduce opacity by adjusting style (simplified approach)
		style = style.Faint(true)
	}

	message := icon + " " + toast.Message

	return style.Render(message)
}
