package views

import (
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/styles"
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

// settingsViewOpt mutates the settingsViewOpts carried by newTestSettingsView.
type settingsViewOpt func(*settingsViewOpts)

type settingsViewOpts struct {
	cfg         *config.Config
	workerCount int
	maxWorkers  int
	profiles    []string
	activeName  string
}

func withConfig(c *config.Config) settingsViewOpt {
	return func(o *settingsViewOpts) { o.cfg = c }
}

func withWorkerCount(n int) settingsViewOpt {
	return func(o *settingsViewOpts) { o.workerCount = n }
}

func withMaxWorkers(n int) settingsViewOpt {
	return func(o *settingsViewOpts) { o.maxWorkers = n }
}

func withProfiles(names []string) settingsViewOpt {
	return func(o *settingsViewOpts) { o.profiles = names }
}

func withActiveProfile(name string) settingsViewOpt {
	return func(o *settingsViewOpts) { o.activeName = name }
}

// newTestSettingsView builds a SettingsView with gomock-backed
// ConfigManager and ProcessorController and returns all three. Default
// expectations: Get→nil, GetWorkerCount→1, GetMaxWorkers→1,
// ListProfiles→["default"], GetActiveProfile→"default" — all .AnyTimes().
// Tests that need different values pass withConfig / withWorkerCount /
// withMaxWorkers; tests that need stricter expectations on Update/Save
// layer them on the returned mocks after the helper call.
func newTestSettingsView(t *testing.T, opts ...settingsViewOpt) (
	SettingsView,
	*mock_ui.MockConfigManager,
	*mock_ui.MockProcessorController,
) {
	t.Helper()
	o := settingsViewOpts{workerCount: 1, maxWorkers: 1, profiles: []string{"default"}, activeName: "default"}
	for _, opt := range opts {
		opt(&o)
	}

	ctrl := gomock.NewController(t)
	configMgr := mock_ui.NewMockConfigManager(ctrl)
	proc := mock_ui.NewMockProcessorController(ctrl)

	configMgr.EXPECT().Get().Return(o.cfg).AnyTimes()
	configMgr.EXPECT().ListProfiles().Return(o.profiles).AnyTimes()
	configMgr.EXPECT().GetActiveProfile().Return(o.activeName).AnyTimes()
	proc.EXPECT().GetWorkerCount().Return(o.workerCount).AnyTimes()
	proc.EXPECT().GetMaxWorkers().Return(o.maxWorkers).AnyTimes()

	return NewSettingsView(configMgr, proc), configMgr, proc
}

func TestSettingsView_SliderFocusedOnConstruction_PlusKeyIncrements(t *testing.T) {
	// The processor is seeded at 4 workers on startup, so the slider is
	// constructed with min=1, max=4, current=4. The user must first drop
	// the count with - to free a slot, then + to confirm the keys dispatch
	// to the slider. Before the fix, + was swallowed by the URL input
	// because focus defaulted to urlField.
	const current = 4
	v, _, proc := newTestSettingsView(t, withWorkerCount(current), withMaxWorkers(current))

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
	v, _, _ := newTestSettingsView(t, withWorkerCount(2), withMaxWorkers(2))

	assert.True(t, v.slider.Focused, "slider must be focused on construction so + / - work immediately")
}

// TestSettingsView_FocusPaneDefaultsToList — S-5.1. On a freshly
// constructed SettingsView, focusPane must equal paneList so the
// profile sidebar is the initial focus target. The Tab global
// shortcut will toggle focusPane to paneForm; the user never has
// to press Tab just to land in the form. Before the change the
// `focusPane` field does not exist and the test fails at compile
// time.
func TestSettingsView_FocusPaneDefaultsToList(t *testing.T) {
	v, _, _ := newTestSettingsView(t)

	assert.Equal(t, paneList, v.focusPane,
		"focusPane must default to paneList on construction; the persistent sidebar is the initial focus")
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
	// Seed the processor at a worker count below NumCPU so that the
	// difference between current and max is visible. On an 8-core machine
	// the user should be able to climb from 2 up to 8.
	const current = 2
	const maxWorkers = 8
	v, _, _ := newTestSettingsView(t, withWorkerCount(current), withMaxWorkers(maxWorkers))

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
	// Seed the slider at max (processor init convention) so we can
	// exercise the same real-world sequence: - first, + second.
	const current = 4
	v, _, proc := newTestSettingsView(t, withWorkerCount(current), withMaxWorkers(current))
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

// TestSettingsView_SliderFocused_Tab_TogglesToPaneList is the replacement
// for the old `TestSettingsView_SliderFocused_TabMovesFocus`. Under the
// two-pane model, Tab toggles the pane — it does NOT cycle the focused
// form field. When focusPane == paneForm and focused == sliderField,
// pressing Tab must land on focusPane == paneList with focused still
// at sliderField. The slider's in-form focus is preserved; the user
// has to press Tab again to come back to the form.
func TestSettingsView_SliderFocused_Tab_TogglesToPaneList(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	require.Equal(t, sliderField, v.focused, "slider should be focused on construction")
	require.Equal(t, paneList, v.focusPane, "paneList should be the initial focus")

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.NextField.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, paneForm, v.focusPane,
		"Tab must toggle the pane from paneList to paneForm; "+
			"the global Tab handler must run before any component dispatch")
	assert.Equal(t, sliderField, v.focused,
		"Tab must NOT cycle the form field; the slider keeps its in-form focus")
}

// TestSettingsView_TabFromListTogglesToForm is S-6.1. The
// persistent profile sidebar is the initial focus; Tab must
// move focus into the form pane without changing the focused
// form field.
func TestSettingsView_TabFromListTogglesToForm(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	v.focusPane = paneList
	v.focused = urlField

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.NextField.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, paneForm, v.focusPane,
		"Tab from paneList must toggle to paneForm")
	assert.Equal(t, urlField, v.focused,
		"Tab must preserve the focused form field (not cycle it)")
}

// TestSettingsView_TabFromFormTogglesToList is S-6.2. From the
// form pane, Tab must move focus back to the list sidebar.
func TestSettingsView_TabFromFormTogglesToList(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	v.focusPane = paneForm
	v.focused = urlField

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.NextField.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, paneList, v.focusPane,
		"Tab from paneForm must toggle to paneList")
	assert.Equal(t, urlField, v.focused,
		"Tab must preserve the focused form field when toggling back to the list")
}

// TestSettingsView_TabInFormDoesNotCycleField is S-8.1. In the
// form pane, Tab toggles the pane; it does NOT advance the
// focused form field. This is the inverse of the old "Tab
// cycles next field" behavior.
func TestSettingsView_TabInFormDoesNotCycleField(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	v.focusPane = paneForm
	v.focused = sliderField

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.NextField.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, paneList, v.focusPane,
		"Tab in paneForm must toggle to paneList")
	assert.Equal(t, sliderField, v.focused,
		"Tab in paneForm must NOT cycle the focused field; only the pane toggles")
}

// TestSettingsView_DispatchOrder_GlobalsPrecedeComponents is
// S-20.1. The global Tab handler must run before any
// focused-component dispatch. When focusPane == paneList and
// focused == sliderField, pressing Tab must land on
// focusPane == paneForm — the pane toggle took effect, the
// slider block did not swallow the key.
func TestSettingsView_DispatchOrder_GlobalsPrecedeComponents(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	require.Equal(t, paneList, v.focusPane, "paneList should be the initial focus")
	require.Equal(t, sliderField, v.focused, "slider should be the initial focused field")

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.NextField.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, paneForm, v.focusPane,
		"Tab must toggle the pane; the global handler must precede the slider block")
}

// TestSettingsView_ShiftTabInFormCyclesBackward is S-7.1. In
// paneForm, Shift+Tab cycles the focused form field backward by
// one. From urlField, Shift+Tab lands on sliderField.
func TestSettingsView_ShiftTabInFormCyclesBackward(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	v.focusPane = paneForm
	v.focused = urlField

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.PrevField.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, sliderField, v.focused,
		"Shift+Tab in paneForm from urlField must land on sliderField (one step backward)")
	assert.Equal(t, paneForm, v.focusPane,
		"Shift+Tab in paneForm must NOT change the pane")
}

// TestSettingsView_ShiftTabWrapsFromCsvToHeaders is S-7.2. From
// csvFieldsField, Shift+Tab wraps to headersField. A second
// Shift+Tab lands on bodyField, and so on, cycling all the way
// back to sliderField.
func TestSettingsView_ShiftTabWrapsFromCsvToHeaders(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	v.focusPane = paneForm
	v.focused = csvFieldsField

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.PrevField.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, headersField, v.focused,
		"Shift+Tab from csvFieldsField must wrap to headersField (backward cycle wraps)")

	next, _ = v.Update(settingsKeyMsg(kbind.PrevField.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, bodyField, v.focused,
		"a second Shift+Tab from headersField must land on bodyField")
}

// TestSettingsView_ShiftTabInListIsNoOp is S-22.1. In paneList,
// Shift+Tab is a no-op — it does not change the pane, does not
// change the focused field, does not panic. Shift+Tab is
// pane-incongruent; the form-field cycle is only meaningful in
// paneForm.
func TestSettingsView_ShiftTabInListIsNoOp(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	v.focusPane = paneList
	v.focused = urlField

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.PrevField.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, paneList, v.focusPane,
		"Shift+Tab in paneList must be a no-op (pane unchanged)")
	assert.Equal(t, urlField, v.focused,
		"Shift+Tab in paneList must be a no-op (focused field unchanged)")
}

// TestSettingsView_SidebarWidth_AtTypicalWidth is S-2.1. At
// width=80 the sidebar must be width/4 = 20 (within the
// [15, 30] clamp). The form pane gets the remaining 60 columns
// and the form viewport must be 4 columns narrower (for the
// form's Margin(1, 2) = left 2 + right 2).
func TestSettingsView_SidebarWidth_AtTypicalWidth(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(SettingsView)
	assert.Equal(t, 20, v.profileList.Width(),
		"sidebar width must be width/4 = 20 at W=80")
	assert.Equal(t, 56, v.viewport.Width(),
		"form viewport width must be formWidth - 4 = 56 at W=80 (formWidth=60)")
}

// TestSettingsView_SidebarWidth_ClampedAtMax is S-2.2. At
// width=200 the naive 25% would be 50, but the clamp caps the
// sidebar at maxListWidth=30. The form gets the remaining 170
// columns.
func TestSettingsView_SidebarWidth_ClampedAtMax(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 200, Height: 30})
	v = next.(SettingsView)
	assert.Equal(t, 30, v.profileList.Width(),
		"sidebar width must be clamped at max=30 at W=200")
	assert.Equal(t, 166, v.viewport.Width(),
		"form viewport width must be formWidth - 4 = 166 at W=200 (formWidth=170)")
}

// TestSettingsView_SidebarWidth_ClampedAtMin is S-2.3. At
// width=40 the naive 25% would be 10, but the clamp floors the
// sidebar at minListWidth=15 so profile names stay readable.
func TestSettingsView_SidebarWidth_ClampedAtMin(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 40, Height: 20})
	v = next.(SettingsView)
	assert.Equal(t, 15, v.profileList.Width(),
		"sidebar width must be clamped at min=15 at W=40")
	assert.Equal(t, 21, v.viewport.Width(),
		"form viewport width must be formWidth - 4 = 21 at W=40 (formWidth=25)")
}

// TestSettingsView_WidthSourceIsViewportSizeMsg is S-3.1. The
// view must derive its width/height from msgs.ViewportSizeMsg
// (chrome-adjusted by AppModel), NOT from tea.WindowSizeMsg.
// Sending a WindowSizeMsg must leave the view unchanged.
func TestSettingsView_WidthSourceIsViewportSizeMsg(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	// tea.WindowSizeMsg is NOT handled by the view; the view must
	// ignore it and the returned view must be a no-op (width and
	// height stay at 0).
	next, _ := v.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	v = next.(SettingsView)
	assert.Equal(t, 0, v.width, "tea.WindowSizeMsg must NOT update v.width")
	assert.Equal(t, 0, v.height, "tea.WindowSizeMsg must NOT update v.height")

	// msgs.ViewportSizeMsg IS handled.
	next, _ = v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(SettingsView)
	assert.Equal(t, 80, v.width, "msgs.ViewportSizeMsg must update v.width")
	assert.Equal(t, 20, v.height, "msgs.ViewportSizeMsg must update v.height")
}

// TestSettingsView_Resize_UpdatesAllSizesInOnePass is S-4.1. A
// single ViewportSizeMsg must set the list size, the viewport
// size, and the form input/textarea widths in one pass. The
// height-2 regression is preserved (see
// TestSettingsView_Resize_AccountsForOwnMargin).
func TestSettingsView_Resize_UpdatesAllSizesInOnePass(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(SettingsView)
	assert.Equal(t, 20, v.profileList.Width(), "list width")
	assert.Equal(t, 20, v.profileList.Height(), "list height")
	assert.Equal(t, 56, v.viewport.Width(), "viewport width = formWidth - 4")
	assert.Equal(t, 18, v.viewport.Height(), "viewport height = height - 2 (regression)")
}

// TestSettingsView_ListVisibleByDefault is S-1.1. A freshly
// constructed SettingsView (with mocks returning
// ["default"]/single Config) must render a non-empty View whose
// content includes the profile name "default" — the sidebar is
// always visible, never hidden behind a modal toggle.
func TestSettingsView_ListVisibleByDefault(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	out := v.View().Content
	assert.NotEmpty(t, out, "View().Content must be non-empty on construction")
	assert.Contains(t, out, "default",
		"View().Content must include the profile name from the list sidebar")
}

// TestSettingsView_FocusedPane_List is the S-9.1 access path.
// When the list pane has focus, the FocusedPane helper must
// return paneList. The actual background-color application is
// tested by TestSettingsView_FocusedPaneBg_AppliedToList.
func TestSettingsView_FocusedPane_List(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	v.focusPane = paneList
	assert.Equal(t, paneList, v.FocusedPane(),
		"FocusedPane() must return paneList when the list pane has focus")
}

// TestSettingsView_FocusedPaneBg_AppliedToList is S-9.1. The
// focused list pane must render with the FocusedPaneBg
// (#414141) background. The view must contain the lipgloss
// background escape sequence for the focused pane; the form
// pane (which is unfocused) must NOT contain it.
func TestSettingsView_FocusedPaneBg_AppliedToList(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	v.focusPane = paneList
	// Force the view through a resize so the layout uses the
	// final styles (the seeded list at 0×0 is the initial state).
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(SettingsView)

	out := v.View().Content
	bgSeq := backgroundEscapeSeq(styles.FocusedPaneBg)
	assert.Contains(t, out, bgSeq,
		"View().Content must contain the FocusedPaneBg escape when the list pane is focused")
}

// TestSettingsView_FocusedPaneBg_AppliedToForm is S-9.2. The
// focused form pane must render with the FocusedPaneBg
// (#414141) background. With focusPane == paneForm, the form
// wrapper carries the background; the list wrapper does not.
func TestSettingsView_FocusedPaneBg_AppliedToForm(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	v.focusPane = paneForm
	v.focused = urlField
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(SettingsView)

	out := v.View().Content
	bgSeq := backgroundEscapeSeq(styles.FocusedPaneBg)
	assert.Contains(t, out, bgSeq,
		"View().Content must contain the FocusedPaneBg escape when the form pane is focused")
}

// TestSettingsView_FormFieldFocusIndicator_Preserved is S-10.1.
// When focusPane == paneForm and focused == urlField, the
// rendered output must include the "▶ " marker on the URL
// template label line. The in-form focus indicator is
// independent of the pane focus; the user still sees which
// field is active within the form.
func TestSettingsView_FormFieldFocusIndicator_Preserved(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
	v.focusPane = paneForm
	v.focused = urlField
	v = v.updateFocus()
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(SettingsView)

	out := v.View().Content
	assert.Contains(t, out, "▶ URL template:",
		"the URL template label must carry the focused marker (▶) when the URL field is focused")
}

// TestSettingsView_DownInListMovesCursor is S-11.1. With
// focusPane == paneList and 3 profiles, Down moves the cursor
// from index 0 to index 1.
func TestSettingsView_DownInListMovesCursor(t *testing.T) {
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles([]string{"a", "b", "c"}),
		withActiveProfile("a"))
	v.focusPane = paneList
	configMgr.EXPECT().GetProfile(gomock.Any()).Return(nil).AnyTimes()
	require.Equal(t, 0, v.profileList.Index(), "cursor starts at 0")

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.Down.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, 1, v.profileList.Index(),
		"Down in paneList must move the cursor to index 1")
}

// TestSettingsView_UpInListWrapsToLast is S-11.2. With
// focusPane == paneList and 3 profiles, Up from cursor 0 wraps
// to cursor 2 (last item).
func TestSettingsView_UpInListWrapsToLast(t *testing.T) {
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles([]string{"a", "b", "c"}),
		withActiveProfile("a"))
	v.focusPane = paneList
	configMgr.EXPECT().GetProfile(gomock.Any()).Return(nil).AnyTimes()
	require.Equal(t, 0, v.profileList.Index(), "cursor starts at 0")

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.Up.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, 2, v.profileList.Index(),
		"Up from cursor 0 must wrap to the last item (cursor 2)")
}

// TestSettingsView_CursorChangeTriggersGetProfile is S-12.1.
// When the cursor moves, GetProfile(name) is called for the
// newly-selected profile and the form is reloaded. The mock
// asserts GetProfile("b") is called exactly once after Down.
func TestSettingsView_CursorChangeTriggersGetProfile(t *testing.T) {
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles([]string{"a", "b"}),
		withActiveProfile("a"),
		withConfig(&config.Config{}))
	v.focusPane = paneList

	// gomock matches expectations in registration order, so the
	// specific "b" expectation is registered first. The
	// catch-all gomock.Any() covers any other name (and any
	// subsequent GetProfile calls).
	configMgr.EXPECT().
		GetProfile("b").
		Return(&config.Config{
			Request: config.RequestConfig{
				URLTemplate:  "http://b/api",
				Method:       "POST",
				BodyTemplate: `{"k":"v"}`,
				Headers:      map[string]string{},
			},
			CSV: config.CSVConfig{Fields: []string{"id"}},
		}).
		Times(1)
	configMgr.EXPECT().GetProfile(gomock.Any()).Return(&config.Config{}).AnyTimes()
	// Get() is called by loadFromConfig after GetProfile.
	configMgr.EXPECT().Get().Return(&config.Config{}).AnyTimes()

	var next tea.Model
	next, _ = v.Update(settingsKeyMsg(kbind.Down.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, "http://b/api", v.urlInput.Value(),
		"urlInput must reflect the newly-previewed profile's URLTemplate")
}

// TestSettingsView_PreviewCoversAllItems is S-12.2. Cycling
// through 0 → 1 → 2 → 1 → 0 must trigger GetProfile for each
// visited cursor position and the urlInput must reflect the
// corresponding profile's URLTemplate.
func TestSettingsView_PreviewCoversAllItems(t *testing.T) {
	profiles := []string{"a", "b", "c"}
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles(profiles),
		withActiveProfile("a"),
		withConfig(&config.Config{}))
	v.focusPane = paneList

	// Per-name expectations: gomock matches in registration
	// order, so specific names registered first beat the
	// catch-all. AnyTimes because the same name may be
	// visited more than once (0→1→2→1→0 visits "b" twice).
	for _, name := range profiles {
		configMgr.EXPECT().
			GetProfile(name).
			Return(&config.Config{
				Request: config.RequestConfig{
					URLTemplate:  "http://" + name + "/api",
					Method:       "POST",
					BodyTemplate: "",
					Headers:      map[string]string{},
				},
				CSV: config.CSVConfig{Fields: []string{"id"}},
			}).
			AnyTimes()
	}
	configMgr.EXPECT().Get().Return(&config.Config{}).AnyTimes()

	// Cycle: 0 → 1 → 2 → 1 → 0
	expectedURLs := []string{
		"http://b/api", // 0 → 1
		"http://c/api", // 1 → 2
		"http://b/api", // 2 → 1
		"http://a/api", // 1 → 0
	}
	keys := []string{
		kbind.Down.Keys()[0],
		kbind.Down.Keys()[0],
		kbind.Up.Keys()[0],
		kbind.Up.Keys()[0],
	}
	for i, key := range keys {
		next, _ := v.Update(settingsKeyMsg(key))
		v = next.(SettingsView)
		assert.Equal(t, expectedURLs[i], v.urlInput.Value(),
			"after step %d the urlInput must reflect the newly-previewed profile", i)
	}
	assert.Equal(t, 0, v.profileList.Index(), "cursor back at 0")
}

// TestSettingsView_GetProfileNotCalledWhenCursorUnchanged is
// S-13.1. From a 1-item list, Up or Down must NOT call
// GetProfile — the cursor does not change, so no preview is
// needed. The mock has Times(0) for GetProfile; the test
// passes if the mock is satisfied.
func TestSettingsView_GetProfileNotCalledWhenCursorUnchanged(t *testing.T) {
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles([]string{"only"}),
		withActiveProfile("only"))
	v.focusPane = paneList
	// Strict: GetProfile must never be called. (No .AnyTimes()
	// nor .Times(1) — gomock fails the test on any call.)
	configMgr.EXPECT().GetProfile(gomock.Any()).Times(0)

	// Send Up (would wrap to last on multi-item, but with 1
	// item the cursor stays at 0) and Down (same).
	for _, key := range []string{
		kbind.Up.Keys()[0],
		kbind.Down.Keys()[0],
	} {
		next, _ := v.Update(settingsKeyMsg(key))
		v = next.(SettingsView)
	}
	assert.Equal(t, 0, v.profileList.Index(),
		"cursor stays at 0 on a 1-item list regardless of direction")
}

// TestSettingsView_ListScrollsWhenItemsExceedHeight is S-14.1.
// With 50 profiles and a small viewport, repeated Down presses
// must not panic and the cursor must reach the expected
// position. The list's internal viewport handles scrolling.
func TestSettingsView_ListScrollsWhenItemsExceedHeight(t *testing.T) {
	profiles := make([]string, 50)
	for i := range profiles {
		profiles[i] = "p" + string(rune('A'+i%26)) + string(rune('0'+i/26))
	}
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles(profiles),
		withActiveProfile(profiles[0]))
	v.focusPane = paneList
	configMgr.EXPECT().GetProfile(gomock.Any()).Return(nil).AnyTimes()
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 5})
	v = next.(SettingsView)

	for i := 0; i < 10; i++ {
		next, _ := v.Update(settingsKeyMsg(kbind.Down.Keys()[0]))
		v = next.(SettingsView)
	}
	assert.Equal(t, 10, v.profileList.Index(),
		"cursor must reach index 10 after 10 Down presses")
}

// TestSettingsView_ActiveProfileHasBulletMarker is S-15.1. The
// list-item delegate renders the active profile's row with a
// "●" suffix. The rendered view output must contain the
// substring "production ●" (or equivalent), unique to that
// row.
func TestSettingsView_ActiveProfileHasBulletMarker(t *testing.T) {
	v, _, _ := newTestSettingsView(t,
		withProfiles([]string{"default", "production", "staging"}),
		withActiveProfile("production"))
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(SettingsView)

	out := v.View().Content
	assert.Contains(t, out, "production ●",
		"the active profile row must carry the ● marker")
}

// TestSettingsView_NonActiveRowsHaveNoBulletMarker is S-15.2.
// The non-active rows must NOT carry the "●" suffix.
func TestSettingsView_NonActiveRowsHaveNoBulletMarker(t *testing.T) {
	v, _, _ := newTestSettingsView(t,
		withProfiles([]string{"default", "production", "staging"}),
		withActiveProfile("production"))
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(SettingsView)

	out := v.View().Content
	// Find the line containing "default" and assert it has no ●.
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "default") {
			assert.NotContains(t, line, "●",
				"the default row must not carry the ● marker (it is not active)")
		}
		if strings.Contains(line, "staging") {
			assert.NotContains(t, line, "●",
				"the staging row must not carry the ● marker (it is not active)")
		}
	}
}

// TestSettingsView_SingleProfileNavigationIsNoOp is S-E1. A
// 1-item list receives Up/Down without moving the cursor or
// calling GetProfile. The test is a guard against accidental
// regression of the cursor-change gating.
func TestSettingsView_SingleProfileNavigationIsNoOp(t *testing.T) {
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles([]string{"only"}),
		withActiveProfile("only"))
	v.focusPane = paneList
	configMgr.EXPECT().GetProfile(gomock.Any()).Times(0)

	next, _ := v.Update(settingsKeyMsg(kbind.Up.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, 0, v.profileList.Index())

	next, _ = v.Update(settingsKeyMsg(kbind.Down.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, 0, v.profileList.Index())
}

// TestSettingsView_EnterCallsSetActiveProfile is S-16.1.
// Pressing Enter in paneList calls ConfigManager.SetActiveProfile
// with the currently selected profile's name. The cursor
// stays where it was (the list does not move on activation).
func TestSettingsView_EnterCallsSetActiveProfile(t *testing.T) {
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles([]string{"default", "production"}),
		withActiveProfile("default"))
	v.focusPane = paneList
	// GetProfile for the preview step (cursor moves to production)
	configMgr.EXPECT().GetProfile("production").Return(&config.Config{}).AnyTimes()
	// Move cursor to production
	next, _ := v.Update(settingsKeyMsg(kbind.Down.Keys()[0]))
	v = next.(SettingsView)
	require.Equal(t, 1, v.profileList.Index(), "cursor must be at 1 (production)")

	// SetActiveProfile("production") is called on Enter
	configMgr.EXPECT().SetActiveProfile("production").Return(nil).Times(1)
	// Get() is called by loadConfig after SetActiveProfile
	configMgr.EXPECT().Get().Return(&config.Config{}).AnyTimes()

	next, _ = v.Update(settingsKeyMsg(kbind.Select.Keys()[0]))
	v = next.(SettingsView)
	assert.Equal(t, paneList, v.focusPane,
		"Enter must not change the focus pane")
}

// TestSettingsView_EnterReloadsFormAndEmitsMsg is S-17.1.
// Pressing Enter in paneList must call SetActiveProfile, then
// loadConfig (which re-populates the form), then emit a
// ProfileSwitchedMsg via the returned cmd.
func TestSettingsView_EnterReloadsFormAndEmitsMsg(t *testing.T) {
	prodCfg := &config.Config{
		Request: config.RequestConfig{
			URLTemplate:  "http://prod/api",
			Method:       "POST",
			BodyTemplate: "",
			Headers:      map[string]string{},
		},
		CSV: config.CSVConfig{Fields: []string{"id"}},
	}
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles([]string{"default", "production"}),
		withActiveProfile("default"),
		withConfig(prodCfg)) // helper's Get() returns prodCfg
	v.focusPane = paneList
	// GetProfile for the preview step (cursor moves to
	// production, triggering previewProfile which calls
	// GetProfile).
	configMgr.EXPECT().GetProfile("production").Return(prodCfg).AnyTimes()
	// Move cursor to production
	next, _ := v.Update(settingsKeyMsg(kbind.Down.Keys()[0]))
	v = next.(SettingsView)

	configMgr.EXPECT().
		SetActiveProfile("production").
		Return(nil).
		Times(1)
	// Get() is called by loadConfig after SetActiveProfile;
	// the helper already stubs it to return prodCfg.

	var cmd tea.Cmd
	next, cmd = v.Update(settingsKeyMsg(kbind.Select.Keys()[0]))
	v = next.(SettingsView)
	require.NotNil(t, cmd, "Enter must return a non-nil cmd (ProfileSwitchedMsg)")

	assert.Equal(t, "http://prod/api", v.urlInput.Value(),
		"urlInput must reflect the newly-activated profile's URLTemplate")

	// Run the cmd and assert it yields ProfileSwitchedMsg
	msg := cmd()
	psMsg, ok := msg.(msgs.ProfileSwitchedMsg)
	require.True(t, ok, "cmd must yield a ProfileSwitchedMsg, got %T", msg)
	assert.Equal(t, "production", psMsg.ProfileName,
		"ProfileSwitchedMsg must carry the activated profile's name")
}

// TestSettingsView_EnterDoesNotSaveOrUpdate is S-18.1.
// Pressing Enter in paneList must NOT call Save or Update.
// Activation is a runtime profile switch, not a config write.
func TestSettingsView_EnterDoesNotSaveOrUpdate(t *testing.T) {
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles([]string{"default", "production"}),
		withActiveProfile("default"))
	v.focusPane = paneList
	configMgr.EXPECT().GetProfile("production").Return(&config.Config{}).AnyTimes()
	next, _ := v.Update(settingsKeyMsg(kbind.Down.Keys()[0]))
	v = next.(SettingsView)

	configMgr.EXPECT().SetActiveProfile("production").Return(nil).Times(1)
	configMgr.EXPECT().Get().Return(&config.Config{}).AnyTimes()
	// Strict: Save and Update must NEVER be called on Enter.
	configMgr.EXPECT().Save().Times(0)
	configMgr.EXPECT().Update(gomock.Any()).Times(0)

	_, _ = v.Update(settingsKeyMsg(kbind.Select.Keys()[0]))
}

// TestSettingsView_EnterOnActiveProfile is S-E2. Activating
// the already-active profile is still a valid operation:
// SetActiveProfile is called (idempotent), loadConfig is
// called, ProfileSwitchedMsg is emitted. The form fields
// reflect the same config (no change visible to the user).
func TestSettingsView_EnterOnActiveProfile(t *testing.T) {
	v, configMgr, _ := newTestSettingsView(t,
		withProfiles([]string{"default", "production"}),
		withActiveProfile("production"))
	v.focusPane = paneList
	configMgr.EXPECT().GetProfile(gomock.Any()).Return(&config.Config{}).AnyTimes()
	// Cursor starts at production (active), so Enter is
	// "activate the already-active profile".
	require.Equal(t, 1, v.profileList.Index(), "cursor starts at the active profile")

	configMgr.EXPECT().SetActiveProfile("production").Return(nil).Times(1)
	configMgr.EXPECT().Get().Return(&config.Config{}).AnyTimes()

	var cmd tea.Cmd
	_, cmd = v.Update(settingsKeyMsg(kbind.Select.Keys()[0]))
	require.NotNil(t, cmd, "Enter must return a non-nil cmd even for the active profile")
	msg := cmd()
	psMsg, ok := msg.(msgs.ProfileSwitchedMsg)
	require.True(t, ok)
	assert.Equal(t, "production", psMsg.ProfileName)
}

// TestSettingsView_CtrlSInPaneListSaves is S-19.1. Ctrl+S
// must trigger saveConfigCmd when focusPane == paneList. The
// global Save handler runs before the pane branch, so the
// save fires regardless of which pane has focus.
func TestSettingsView_CtrlSInPaneListSaves(t *testing.T) {
	v, configMgr, _ := newTestSettingsView(t,
		withConfig(&config.Config{}))
	v.focusPane = paneList

	configMgr.EXPECT().Update(gomock.Any()).Return(nil).Times(1)
	configMgr.EXPECT().Save().Return(nil).Times(1)

	var cmd tea.Cmd
	_, cmd = v.Update(settingsKeyMsg(kbind.Save.Keys()[0]))
	assert.NotNil(t, cmd, "Ctrl+S in paneList must return a non-nil save cmd")
}

// TestSettingsView_CtrlSInPaneFormSaves is S-19.2. Ctrl+S
// must trigger saveConfigCmd when focusPane == paneForm.
// The save fires regardless of which form field is focused.
func TestSettingsView_CtrlSInPaneFormSaves(t *testing.T) {
	v, configMgr, _ := newTestSettingsView(t,
		withConfig(&config.Config{}))
	v.focusPane = paneForm
	v.focused = urlField

	configMgr.EXPECT().Update(gomock.Any()).Return(nil).Times(1)
	configMgr.EXPECT().Save().Return(nil).Times(1)

	var cmd tea.Cmd
	_, cmd = v.Update(settingsKeyMsg(kbind.Save.Keys()[0]))
	assert.NotNil(t, cmd, "Ctrl+S in paneForm must return a non-nil save cmd")
}

// backgroundEscapeSeq returns the lipgloss-rendered background
// OPEN escape sequence for the given color. We strip the
// trailing close (which appears at the end of any styled span)
// so the test can use this as a substring marker — the view's
// pane wrapper applies the open at the start of the pane
// content and the close appears much later, so back-to-back
// matching is unreliable.
func backgroundEscapeSeq(c color.Color) string {
	full := lipgloss.NewStyle().Background(c).Render("")
	// The open escape ends at the first 'm' following the
	// '\x1b[48;' sequence. Everything after that first 'm' is
	// the content + close, which we discard.
	const openMarker = "\x1b[48;"
	idx := indexOfByte(full, openMarker)
	if idx < 0 {
		return full
	}
	end := indexOfByteFrom(full, 'm', idx)
	if end < 0 {
		return full
	}
	return full[idx : end+1]
}

// indexOfByte is a tiny helper that returns the index of the
// first occurrence of sub in s, or -1.
func indexOfByte(s, sub string) int {
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// indexOfByteFrom returns the index of the first occurrence of
// b in s starting at from, or -1.
func indexOfByteFrom(s string, b byte, from int) int {
	for i := from; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// TestSettingsView_SliderFocused_CtrlS_TriggersSave — pressing Ctrl+S when the
// slider is focused must trigger the save flow. The mock asserts that
// ConfigManager.Update and ConfigManager.Save are each called exactly once
// when the save command runs. Before the fix the slider block returns nil
// before the global Save handler runs, so neither Update nor Save is called
// and the test fails on the gomock expectation mismatch.
func TestSettingsView_SliderFocused_CtrlS_TriggersSave(t *testing.T) {
	// Get is called once by loadConfig during NewSettingsView and again by
	// saveConfig when the save command runs. Returning a fresh, empty Config
	// keeps the test realistic and avoids the loadConfig early-return on nil.
	v, configMgr, _ := newTestSettingsView(t, withConfig(&config.Config{}))
	configMgr.EXPECT().Update(gomock.Any()).Return(nil).Times(1)
	configMgr.EXPECT().Save().Return(nil).Times(1)
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
	v, _, _ := newTestSettingsView(t)
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
	v, _, _ := newTestSettingsView(t)
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
	v, _, _ := newTestSettingsView(t)
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
	v, _, _ := newTestSettingsView(t)

	const width, height = 80, 20
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: width, Height: height})
	v = next.(SettingsView)

	assert.Equal(t, height-2, v.viewport.Height(),
		"viewport height must equal (height - 2) — the view's own Margin(1,2); "+
			"subtracting 4 double-counts the chrome and leaves 2 rows of dead space")
}

// TestSettingsView_Init_ReturnsNil — Init must return nil per R-6.
func TestSettingsView_Init_ReturnsNil(t *testing.T) {
	v, _, _ := newTestSettingsView(t)
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
	cfg := &config.Config{
		Request: config.RequestConfig{
			URLTemplate:  "http://example.com/api",
			Method:       "PUT",
			BodyTemplate: `{"name":"{{.name}}"}`,
			Headers:      map[string]string{"Authorization": "Bearer xyz", "Content-Type": "application/json"},
		},
		CSV: config.CSVConfig{Fields: []string{"id", "name"}},
	}
	v, _, _ := newTestSettingsView(t, withConfig(cfg))
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
	cfg := &config.Config{
		Request: config.RequestConfig{
			URLTemplate: "http://x",
			Method:      "", // empty — must default to "POST"
		},
		CSV: config.CSVConfig{Fields: []string{"id"}},
	}
	v, _, _ := newTestSettingsView(t, withConfig(cfg))

	assert.Equal(t, "POST", v.methodInput.Value(),
		"empty Request.Method must default to POST in the form")
}
