package views

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	proc.EXPECT().GetMaxWorkers().Return(current).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()

	v := NewSettingsView(configMgr, proc)

	// Step 1: press - to lower the count, confirming the slider received
	// the key and the processor was updated.
	proc.EXPECT().SetWorkers(current - 1).Times(1)
	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.SliderDec.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, current-1, v.slider.Value, "slider value should drop after -")

	// Step 2: press + to raise the count back. This is the actual bug
	// scenario: the user lands in Settings, presses +, and expects the
	// worker count to grow.
	proc.EXPECT().SetWorkers(current).Times(1)
	next, _ = v.Update(settingsKeyMsg(kbind.SliderInc.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, current, v.slider.Value, "slider value should rise after +")
}

func TestSettingsView_SliderFocusedOnConstruction_RendersFocusIndicator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	proc.EXPECT().GetWorkerCount().Return(2).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(2).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()

	v := NewSettingsView(configMgr, proc)

	assert.True(t, v.slider.Focused, "slider must be focused on construction so + / - work immediately")
}

// TestSettingsView_SliderMaxEqualsProcessorMaxWorkers is the regression test
// for the bug where the worker-count slider was constructed with Max equal to
// the current worker count instead of the hardware maximum. The slider Max
// must reflect processor.MaxWorkers (runtime.NumCPU()) so that pressing `+`
// can raise the count beyond the current setting.
//
// Root cause: settings.go used `initial` for both the slider's Value and Max
// arguments to NewSlider. If the processor started at 2 workers (e.g. via
// the `-workers` flag), the slider was bounded to [1, 2] and the user could
// never grow it back up to NumCPU() from the Settings view.
//
// Fix: introduce GetMaxWorkers() on the ProcessorController port and pass
// its return value as the slider's Max. The hexagonal boundary is preserved
// because the views package continues to depend only on the port interface.
func TestSettingsView_SliderMaxEqualsProcessorMaxWorkers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	// Seed the processor at a worker count below NumCPU so that the
	// difference between current and max is visible. On an 8-core machine
	// the user should be able to climb from 2 up to 8.
	const current = 2
	const maxWorkers = 8
	proc.EXPECT().GetWorkerCount().Return(current).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(maxWorkers).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()

	v := NewSettingsView(configMgr, proc)

	assert.Equal(t, maxWorkers, v.slider.Max,
		"slider Max must equal processor.MaxWorkers, not the current worker count; "+
			"otherwise the user is trapped at the initial value and can never grow it")
	assert.Equal(t, current, v.slider.Value, "slider Value must be seeded at the current worker count")
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
	proc.EXPECT().GetMaxWorkers().Return(current).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()

	v := NewSettingsView(configMgr, proc)
	assert.Equal(t, current, v.slider.Value, "slider initialised at current worker count")
	assert.True(t, v.slider.Focused, "slider must be focused so + is dispatched to it")

	// Step 1: - key on a US keyboard is unshifted, so the terminal
	// sends the legacy {Text:"-", Code:'-'} form. The binding accepts
	// both "-" and "shift+-", so this works in any layout.
	proc.EXPECT().SetWorkers(current - 1).Times(1)
	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.SliderDec.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, current-1, v.slider.Value, "slider value drops after -")

	// Step 2: the real Kitty-protocol `+` keypress. Before the fix
	// Matches returns false (no binding key equals "shift+="), so
	// SetWorkers is never called and the value stays at current-1.
	// After the fix the binding includes "shift+=" and the value
	// grows back to current.
	proc.EXPECT().SetWorkers(current).Times(1)
	next, _ = v.Update(kittyPlusKeyMsg())
	v = next.(SettingsView)
	assert.Equal(t, current, v.slider.Value,
		"slider value must grow after Kitty-protocol `+` keypress; "+
			"the binding must match the shift+= keystroke representation")
}

// --------------------------------------------------------------------------------
// Settings view input dispatch — RED tests for the slider+modal bug.
//
// The Settings view's key dispatch chain in `Update` previously checked the
// focused-component blocks (slider, text fields) BEFORE the global shortcut
// switch (Tab/Shift+Tab/Ctrl+S/Ctrl+P). The slider block at
// settings.go:310-318 had a blanket `return nil` that swallowed every key —
// even though `Slider.Update` itself only acts on +/-. Since commit 5dbaada
// the slider is the default focus on construction, so the user lands in
// Settings with no keyboard path to change focus, save, or open the profile
// selector.
//
// The fix (see `sdd/.../design` in Engram) reorders the dispatch chain so
// globals run first, narrows the slider block to only intercept
// kbind.SliderInc/SliderDec, and extends the profile-selector modal with a
// Ctrl+P case. The tests below are the contract that the fix must satisfy.
// They are written BEFORE the fix and must fail in the red phase.
// --------------------------------------------------------------------------------

// TestSettingsView_SliderFocused_TabMovesFocus — pressing Tab when the slider
// is focused must advance focus to the next field. Before the fix the slider
// block at settings.go:310-318 returns nil for every key, so the global
// NextField handler at the bottom of Update is unreachable when the slider
// has focus. This test will fail in the red phase with `v.focused == sliderField`
// still true.
func TestSettingsView_SliderFocused_TabMovesFocus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	proc.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(1).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()

	v := NewSettingsView(configMgr, proc)
	require.Equal(t, sliderField, v.focused, "slider should be focused on construction")

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.NextField.Keys()[0]))
	v = next.(SettingsView)
	assert.NotEqual(t, sliderField, v.focused,
		"Tab must move focus away from the slider when the slider is focused; "+
			"the slider key-handling block must not swallow global navigation keys")
}

// TestSettingsView_SliderFocused_CtrlS_TriggersSave — pressing Ctrl+S when the
// slider is focused must trigger the save flow. The mock asserts that
// ConfigManager.Update and ConfigManager.Save are each called exactly once
// when the save command runs. Before the fix the slider block returns nil
// before the global Save handler runs, so neither Update nor Save is called
// and the test fails on the gomock expectation mismatch.
func TestSettingsView_SliderFocused_CtrlS_TriggersSave(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	proc.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(1).AnyTimes()
	// Get is called once by loadConfig during NewSettingsView and again by
	// saveConfig when the save command runs. Returning a fresh, empty Config
	// keeps the test realistic and avoids the loadConfig early-return on nil.
	configMgr.EXPECT().Get().Return(&config.Config{}).AnyTimes()
	configMgr.EXPECT().Update(gomock.Any()).Return(nil).Times(1)
	configMgr.EXPECT().Save().Return(nil).Times(1)

	v := NewSettingsView(configMgr, proc)
	require.Equal(t, sliderField, v.focused, "slider should be focused on construction")

	var next tea.Model
	cmd := tea.Cmd(nil)
	next, cmd = v.Update(settingsKeyMsg("ctrl+s"))
	v = next.(SettingsView)
	assert.NotNil(t, cmd,
		"Ctrl+S must trigger saveConfigCmd when the slider is focused; "+
			"the slider key-handling block must not swallow the Save global shortcut")
}

// TestSettingsView_SliderFocused_CtrlP_TogglesProfileSelector — pressing Ctrl+P
// when the slider is focused must open the profile selector. Before the fix
// the slider block returns nil, so the global Profile handler never toggles
// v.showProfileSelector. The mock stubs ListProfiles/GetActiveProfile because
// the Profile handler (settings.go:359-371) seeds profileListIndex from them.
func TestSettingsView_SliderFocused_CtrlP_TogglesProfileSelector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	proc.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(1).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()
	configMgr.EXPECT().ListProfiles().Return([]string{"default"}).AnyTimes()
	configMgr.EXPECT().GetActiveProfile().Return("default").AnyTimes()

	v := NewSettingsView(configMgr, proc)
	require.Equal(t, sliderField, v.focused, "slider should be focused on construction")
	require.False(t, v.showProfileSelector, "profile selector should be closed on construction")

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg("ctrl+p"))
	v = next.(SettingsView)
	assert.True(t, v.showProfileSelector,
		"Ctrl+P must toggle the profile selector open when the slider is focused; "+
			"the slider key-handling block must not swallow the Profile global shortcut")
}

// TestSettingsView_ModalOpen_CtrlP_ClosesSelector — pressing Ctrl+P while the
// profile-selector modal is open must close the modal without changing focus
// or producing a save command. Before the fix the modal block at
// settings.go:320-350 has no Ctrl+P case; the switch falls through to the
// bare `return nil` at line 350, so the modal stays open. After the fix the
// modal block handles Ctrl+P and sets v.showProfileSelector to false.
func TestSettingsView_ModalOpen_CtrlP_ClosesSelector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	proc.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(1).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()
	configMgr.EXPECT().ListProfiles().Return([]string{"default"}).AnyTimes()
	configMgr.EXPECT().GetActiveProfile().Return("default").AnyTimes()

	v := NewSettingsView(configMgr, proc)
	require.Equal(t, sliderField, v.focused, "slider should be focused on construction")

	// Pre-set the modal-open state the test is exercising. The view does
	// not expose a public setter; the test lives in the same package and
	// reaches in directly. This mirrors the real flow reached by the first
	// Ctrl+P press.
	v.showProfileSelector = true
	v.profileListIndex = 0

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg("ctrl+p"))
	v = next.(SettingsView)
	assert.False(t, v.showProfileSelector,
		"Ctrl+P must close the profile selector when the modal is open; "+
			"the modal block must handle the Profile key, not only Esc")
	assert.Equal(t, sliderField, v.focused,
		"Ctrl+P inside the modal must not change the focused field")
}

// TestSettingsView_CtrlP_TwiceIsIdempotent — pressing Ctrl+P twice in a row
// must return the view to its original state: modal closed, focus unchanged.
// The intermediate assertion (modal open after the first press) is what
// makes this test actually red: before the fix the first press is swallowed
// by the slider block, so the intermediate check fails. The end-state
// assertion (modal closed after the second press) is the re-entrancy
// guarantee: toggling twice must be a no-op on the modal state.
func TestSettingsView_CtrlP_TwiceIsIdempotent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	proc.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(1).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()
	configMgr.EXPECT().ListProfiles().Return([]string{"default"}).AnyTimes()
	configMgr.EXPECT().GetActiveProfile().Return("default").AnyTimes()

	v := NewSettingsView(configMgr, proc)
	require.Equal(t, sliderField, v.focused, "slider should be focused on construction")
	require.False(t, v.showProfileSelector, "profile selector should be closed on construction")

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg("ctrl+p"))
	v = next.(SettingsView)
	assert.True(t, v.showProfileSelector,
		"first Ctrl+P must open the profile selector when the slider is focused")

	next, _ = v.Update(settingsKeyMsg("ctrl+p"))
	v = next.(SettingsView)
	assert.False(t, v.showProfileSelector,
		"second Ctrl+P must close the profile selector (re-entrancy)")
	assert.Equal(t, sliderField, v.focused,
		"two consecutive Ctrl+P presses must leave focus unchanged")
}

// TestSettingsView_Resize_AccountsForOwnMargin is the regression test
// for BUG-002: SettingsView.Resize() subtracted height-4 from the
// viewport height, but the chrome (marginRows + headerHeight +
// statusBarHeight) is already deducted by the AppModel's WindowSizeMsg
// handler, so the `height` argument here is the post-chrome area. The
// view's own settingsAppStyle.Margin(1, 2) consumes 2 rows (1 top + 1
// bottom), so the viewport must be set to height-2 — not height-4. The
// bug produced 2 rows of dead space at the bottom of the form (the
// user reported "settings com muito espaço vazio em baixo").
//
// Root cause: Resize was double-subtracting chrome. The fix changes
// `height - 4` to `height - 2`. The width deduction (`width - 4`) is
// kept unchanged because the view-local Margin(1, 2) consumes 2
// columns on each side and the chrome already deducted 4 columns
// (marginCols), so the total horizontal margin is correctly 4.
func TestSettingsView_Resize_AccountsForOwnMargin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	proc.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(1).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()

	v := NewSettingsView(configMgr, proc)

	const width, height = 80, 20
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: width, Height: height})
	v = next.(SettingsView)

	assert.Equal(t, height-2, v.viewport.Height(),
		"viewport height must equal (height - 2) — the view's own Margin(1,2); "+
			"subtracting 4 double-counts the chrome and leaves 2 rows of dead space")
}

// TestSettingsView_Init_ReturnsNil — Init must return nil per R-6.
func TestSettingsView_Init_ReturnsNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(1).AnyTimes()
	configMgr.EXPECT().Get().Return(nil).AnyTimes()
	v := NewSettingsView(configMgr, proc)
	if cmd := v.Init(); cmd != nil {
		t.Fatalf("Init must return nil; got %T", cmd)
	}
}

// TestSettingsView_LoadConfig_PopulatesInputs is the regression test
// for the value-receiver mutation-loss bug in loadConfig. Before the
// fix the function populated the value-receiver copy's inputs and
// returned nothing, so the form started empty and stayed empty after
// a profile switch. The fix returns the modified SettingsView so the
// populated form state is captured at the call site.
//
// Test contract:
//   - The mock ConfigManager.Get() returns a fully-populated *config.Config.
//   - The returned SettingsView's form inputs reflect the config values.
//   - When Request.Method is empty, the method input defaults to "POST".
//   - loadConfig called directly (post-construction) returns a SettingsView
//     whose inputs also reflect the new config — this is the path the
//     spec's switchProfile scenario exercises.
func TestSettingsView_LoadConfig_PopulatesInputs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(1).AnyTimes()

	cfg := &config.Config{
		Request: config.RequestConfig{
			URLTemplate:  "http://example.com/api",
			Method:       "PUT",
			BodyTemplate: `{"name":"{{.name}}"}`,
			Headers:      map[string]string{"Authorization": "Bearer xyz", "Content-Type": "application/json"},
		},
		CSV: config.CSVConfig{Fields: []string{"id", "name"}},
	}
	configMgr.EXPECT().Get().Return(cfg).AnyTimes()

	v := NewSettingsView(configMgr, proc)
	require.Equal(t, sliderField, v.focused, "slider is the initial focus")

	assert.Equal(t, "http://example.com/api", v.urlInput.Value(),
		"urlInput must reflect config.Request.URLTemplate")
	assert.Equal(t, "PUT", v.methodInput.Value(),
		"methodInput must reflect config.Request.Method when set")
	assert.Equal(t, `{"name":"{{.name}}"}`, v.bodyInput.Value(),
		"bodyInput must reflect config.Request.BodyTemplate")

	// Headers are converted to "Key: Value" lines joined by newline.
	// map iteration order is not stable in Go, so we assert each line is
	// present in the textarea's value rather than asserting the exact
	// joined string.
	headersVal := v.headersInput.Value()
	assert.Contains(t, headersVal, "Authorization: Bearer xyz",
		"headersInput must contain the Authorization header line")
	assert.Contains(t, headersVal, "Content-Type: application/json",
		"headersInput must contain the Content-Type header line")
	assert.Equal(t, 2, strings.Count(headersVal, "\n")+1,
		"headersInput must contain exactly two header lines")

	assert.Equal(t, "id\nname", v.csvFieldsInput.Value(),
		"csvFieldsInput must reflect config.CSV.Fields joined by newline")
}

// TestSettingsView_LoadConfig_EmptyMethodDefaultsToPost is the
// regression test for the "empty method defaults to POST" scenario
// from the spec.
func TestSettingsView_LoadConfig_EmptyMethodDefaultsToPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(1).AnyTimes()

	cfg := &config.Config{
		Request: config.RequestConfig{
			URLTemplate: "http://x",
			Method:      "", // empty — must default to "POST"
		},
		CSV: config.CSVConfig{Fields: []string{"id"}},
	}
	configMgr.EXPECT().Get().Return(cfg).AnyTimes()

	v := NewSettingsView(configMgr, proc)

	assert.Equal(t, "POST", v.methodInput.Value(),
		"empty Request.Method must default to POST in the form")
}
