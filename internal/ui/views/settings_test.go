package views

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// settingsKeyMsg builds a tea.KeyPressMsg for the given key text so the
// settings view can be driven from unit tests without depending on the
// bubbles key.Matches helper's specific string format.
func settingsKeyMsg(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Text: text, Code: rune(text[0])}
}

func TestSettingsView_SliderFocusedOnConstruction_PlusKeyIncrements(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	// The processor is seeded at 4 workers on startup, so the slider is
	// constructed with min=1, max=4, current=4. The user must first drop
	// the count with - to free a slot, then + to confirm the keys dispatch
	// to the slider. Before the fix, + was swallowed by the URL input
	// because focus defaulted to urlField.
	const current = 4
	proc.EXPECT().GetWorkerCount().Return(current).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()

	v := NewSettingsView(configMgr, proc)

	// Step 1: press - to lower the count, confirming the slider received
	// the key and the processor was updated.
	proc.EXPECT().SetWorkers(current - 1).Times(1)
	_ = v.Update(settingsKeyMsg(kbind.SliderDec.Keys()[0]))
	assert.Equal(t, current-1, v.slider.Value, "slider value should drop after -")

	// Step 2: press + to raise the count back. This is the actual bug
	// scenario: the user lands in Settings, presses +, and expects the
	// worker count to grow.
	proc.EXPECT().SetWorkers(current).Times(1)
	_ = v.Update(settingsKeyMsg(kbind.SliderInc.Keys()[0]))
	assert.Equal(t, current, v.slider.Value, "slider value should rise after +")
}

func TestSettingsView_SliderFocusedOnConstruction_RendersFocusIndicator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	proc.EXPECT().GetWorkerCount().Return(2).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()

	v := NewSettingsView(configMgr, proc)

	assert.True(t, v.slider.Focused, "slider must be focused on construction so + / - work immediately")
}
