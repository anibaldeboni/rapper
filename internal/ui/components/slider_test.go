package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/stretchr/testify/assert"
)

func TestSlider_View_RendersTrackHandleAndCount(t *testing.T) {
	t.Run("renders handle and current/max text for mid value", func(t *testing.T) {
		s := *NewSlider("Worker Count", 1, 8, 3)

		out := s.View()

		// Track characters: must contain both track and handle runes
		assert.Contains(t, out, "●", "handle should be visible")
		assert.True(t, strings.ContainsAny(out, "─━"), "track should be visible")
		// Label and current/max text must be present
		assert.Contains(t, out, "Worker Count")
		assert.Contains(t, out, "3 / 8")
	})

	t.Run("renders at min value", func(t *testing.T) {
		s := *NewSlider("W", 1, 4, 1)

		out := s.View()

		assert.Contains(t, out, "●", "handle visible at min")
		assert.Contains(t, out, "1 / 4")
	})

	t.Run("renders at max value", func(t *testing.T) {
		s := *NewSlider("W", 1, 4, 4)

		out := s.View()

		assert.Contains(t, out, "●", "handle visible at max")
		assert.Contains(t, out, "4 / 4")
	})
}

func TestSlider_Update_IncrementsAndDecrements(t *testing.T) {
	t.Run("SliderInc increments value by 1 and clamps at max", func(t *testing.T) {
		s := *NewSlider("W", 1, 5, 4)

		// Increment
		s, _ = s.Update(keyPressMsg(kbind.SliderInc.Keys()[0]))
		assert.Equal(t, 5, s.Value, "value should be 5 after increment")

		// At max, further increment should be clamped
		s, _ = s.Update(keyPressMsg(kbind.SliderInc.Keys()[0]))
		assert.Equal(t, 5, s.Value, "value should stay at max")
	})

	t.Run("SliderDec decrements value by 1 and clamps at min", func(t *testing.T) {
		s := *NewSlider("W", 1, 5, 2)

		// Decrement
		s, _ = s.Update(keyPressMsg(kbind.SliderDec.Keys()[0]))
		assert.Equal(t, 1, s.Value, "value should be 1 after decrement")

		// At min, further decrement should be clamped
		s, _ = s.Update(keyPressMsg(kbind.SliderDec.Keys()[0]))
		assert.Equal(t, 1, s.Value, "value should stay at min")
	})

	t.Run("unrelated key is ignored", func(t *testing.T) {
		s := *NewSlider("W", 1, 5, 3)

		s, _ = s.Update(keyPressMsg("x"))

		assert.Equal(t, 3, s.Value, "value should remain unchanged for unrelated keys")
	})
}

func TestSlider_Focused_RendersIndicator(t *testing.T) {
	t.Run("focused slider shows focus indicator prefix", func(t *testing.T) {
		s := *NewSlider("Worker Count", 1, 8, 3)
		s.Focused = true

		out := s.View()

		assert.Contains(t, out, "▶", "focused slider should show focus indicator")
	})
}

// keyPressMsg builds a tea.KeyPressMsg with the given key text.
// For printable keys (like "+" and "-"), the matching is done by string compare
// against the keybinding keys via the bubbles key.Matches helper.
func keyPressMsg(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Text: text, Code: rune(text[0])}
}
