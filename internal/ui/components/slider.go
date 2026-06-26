package components

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
)

// trackWidth is the fixed character width of the slider's interactive track.
// Chosen to match the original Workers-view slider so visual rhythm is
// preserved when the component replaces the inline renderer.
const trackWidth = 30

var (
	sliderTrackIdleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sliderTrackFocusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	sliderFillStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	sliderHandleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("205")).Bold(true)
	sliderLabelFocusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	sliderLabelIdleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
)

// Slider is a reusable horizontal control that lets the user pick a value in
// [Min, Max] using the + and - keys. It owns its own state so callers should
// treat the returned value as authoritative and dispatch SetWorkers on
// changes via Update's returned command (or by polling Value).
type Slider struct {
	Value   int
	Min     int
	Max     int
	Label   string
	Width   int
	Focused bool
}

// NewSlider creates a slider with the given label, range, and initial value.
// The caller is responsible for clamping `initial` into [min, max] before
// passing it in; the constructor does not silently correct out-of-range
// values so misuse is visible at runtime.
func NewSlider(label string, min, max, initial int) *Slider {
	return &Slider{
		Value: initial,
		Min:   min,
		Max:   max,
		Label: label,
		Width: trackWidth,
	}
}

// View renders the slider as a labelled track with a handle and a "value/max"
// counter. The output never panics on degenerate inputs (max <= min renders
// an empty track with the label and counter).
func (s Slider) View() string {
	label := s.labelView()
	track := s.trackView()
	counter := fmt.Sprintf(" %d / %d", s.Value, s.Max)

	return label + " " + track + counter
}

// Update handles key events that drive the slider. +/SliderInc increments
// (clamped to Max), -/SliderDec decrements (clamped to Min). Any other
// message leaves the slider untouched. The returned command is always nil
// because the slider does not schedule work.
func (s Slider) Update(msg tea.Msg) (Slider, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}

	switch {
	case key.Matches(keyMsg, kbind.SliderInc):
		if s.Value < s.Max {
			s.Value++
		}
	case key.Matches(keyMsg, kbind.SliderDec):
		if s.Value > s.Min {
			s.Value--
		}
	}

	return s, nil
}

// labelView renders the label with a focus prefix when focused.
func (s Slider) labelView() string {
	if s.Focused {
		return sliderLabelFocusStyle.Render("▶ " + s.Label + ":")
	}
	return sliderLabelIdleStyle.Render(s.Label + ":")
}

// trackView renders the 30-column track with a handle. When the range is
// degenerate the track is filled with idle dashes and no handle is drawn.
func (s Slider) trackView() string {
	width := s.Width
	if width <= 0 {
		width = trackWidth
	}
	if s.Max <= s.Min {
		return sliderTrackIdleStyle.Render(strings.Repeat("─", width))
	}

	// Position is the index where the handle sits. Clamp defensively so a
	// out-of-range Value never indexes past the slice.
	pos := int(float64(s.Value-s.Min) / float64(s.Max-s.Min) * float64(width-1))
	if pos < 0 {
		pos = 0
	}
	if pos > width-1 {
		pos = width - 1
	}

	trackStyle := sliderTrackIdleStyle
	if s.Focused {
		trackStyle = sliderTrackFocusStyle
	}

	var b strings.Builder
	for i := range width {
		switch {
		case i == pos:
			b.WriteString(sliderHandleStyle.Render("●"))
		case i < pos:
			b.WriteString(sliderFillStyle.Render("━"))
		default:
			b.WriteString(trackStyle.Render("─"))
		}
	}
	return b.String()
}
