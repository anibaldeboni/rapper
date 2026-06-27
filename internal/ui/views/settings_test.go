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

// kittyPlusKeyMsg simulates a real `+` keypress as it arrives from terminals
// that have the Kitty keyboard protocol enabled (Kitty, WezTerm, foot,
// recent xterm builds). Bubble Tea v2 enables it via
// KeyboardEnhancements.ReportEventTypes, so the shifted character arrives
// as {Text: "", Code: '=', Mod: ModShift} — not as the legacy {Text: "+"}.
// The String() of such a message is "shift+=" (because Text is empty and
// Keystroke() walks the modifier+code path), which is what we have to
// match against in the keybinding.
func kittyPlusKeyMsg() tea.KeyPressMsg {
	return tea.KeyPressMsg{Text: "", Code: '=', Mod: tea.ModShift}
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

// TestSettingsView_SliderAcceptsKittyProtocolPlusKey is the regression test
// for the reported bug: pressing `+` on a real terminal that has the Kitty
// keyboard protocol enabled does not increment the worker count.
//
// Root cause: the SliderInc binding was WithKeys("+") only. The bubbles
// Matches helper compares the KeyPressMsg.String() to the binding keys via
// plain string equality. Kitty-protocol `+` arrives as
// {Text: "", Code: '=', Mod: ModShift} whose String() is "shift+=" — that
// does not equal "+", so the slider key handler never fires and the
// processor is never asked to grow the worker count.
//
// Fix: register the shifted keystroke representation alongside the
// shifted-text one in the binding.
//
// Test flow mirrors the real UX: the processor initialises at max (here
// 4 workers on a 4-core machine), so the slider is at 4/4. The user
// presses `-` (unshifted on US keyboards, so it arrives as
// {Text:"-", Code:'-'}) to free a slot, then `+` (shift+= in Kitty) to
// confirm the bug is fixed. Without the binding fix the second Update
// call silently no-ops and SetWorkers(4) is never invoked; with the
// fix SetWorkers(4) is called and v.slider.Value climbs back to 4.
func TestSettingsView_SliderAcceptsKittyProtocolPlusKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	// Seed the slider at max (processor init convention) so we can
	// exercise the same real-world sequence: - first, + second.
	const current = 4
	proc.EXPECT().GetWorkerCount().Return(current).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()

	v := NewSettingsView(configMgr, proc)
	assert.Equal(t, current, v.slider.Value, "slider initialised at current worker count")
	assert.True(t, v.slider.Focused, "slider must be focused so + is dispatched to it")

	// Step 1: - key on a US keyboard is unshifted, so the terminal
	// sends the legacy {Text:"-", Code:'-'} form. The binding accepts
	// both "-" and "shift+-", so this works in any layout.
	proc.EXPECT().SetWorkers(current - 1).Times(1)
	_ = v.Update(settingsKeyMsg(kbind.SliderDec.Keys()[0]))
	assert.Equal(t, current-1, v.slider.Value, "slider value drops after -")

	// Step 2: the real Kitty-protocol `+` keypress. Before the fix
	// Matches returns false (no binding key equals "shift+="), so
	// SetWorkers is never called and the value stays at current-1.
	// After the fix the binding includes "shift+=" and the value
	// grows back to current.
	proc.EXPECT().SetWorkers(current).Times(1)
	_ = v.Update(kittyPlusKeyMsg())
	assert.Equal(t, current, v.slider.Value,
		"slider value must grow after Kitty-protocol `+` keypress; "+
			"the binding must match the shift+= keystroke representation")
}
