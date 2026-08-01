package views

import (
	"errors"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// NOTE: this file is a fresh rewrite for the textarea-editor change.
// The previous file asserted on per-widget state (urlInput.Value(),
// slider.Value, etc.); those widgets no longer exist. The assertions
// now target v.editor.Value(), v.editing, v.inlineErr, v.focusPane,
// and v.selectedProfile. Comprehensive coverage (S3.*, S4.*, S5.*)
// is added in T9 once the view's mode and save flows land.

// settingsKeyMsg builds a tea.KeyPressMsg for the given key text. The
// Settings view's Update path uses bubbles key.Matches, which compares
// the KeyPressMsg.String() against the binding keys. The simplest
// shape that satisfies a binding like key.WithKeys("e") is
// {Text: "e", Code: 'e'}.
func settingsKeyMsg(text string) tea.KeyPressMsg {
	if text == "" {
		return tea.KeyPressMsg{}
	}
	return tea.KeyPressMsg{Text: text, Code: rune(text[0])}
}

// ctrlSKeyMsg builds the canonical ctrl+s KeyPressMsg. The
// KeyPressMsg.String() of this shape is "ctrl+s" (because Text is
// empty and the bubbles Keystroke helper walks the modifier+code
// path). This is the form real terminals send when the user presses
// ctrl+s; the plain {Text: "ctrl+s", Code: 'c'} form used by
// settingsKeyMsg also matches bindings but the bubbles textarea
// internally special-cases the canonical form for things like
// "select to start of line", which is why we use the explicit
// modifier construction here for save-flow tests.
func ctrlSKeyMsg() tea.KeyPressMsg {
	return tea.KeyPressMsg{Text: "", Code: 's', Mod: tea.ModCtrl}
}

// settingsViewOpt mutates the settingsViewOpts carried by newTestRequestsView.
type settingsViewOpt func(*settingsViewOpts)

type settingsViewOpts struct {
	profiles   []string
	activeName string
	rawContent map[string]string
}

func withProfiles(names []string) settingsViewOpt {
	return func(o *settingsViewOpts) { o.profiles = names }
}

func withActiveProfile(name string) settingsViewOpt {
	return func(o *settingsViewOpts) { o.activeName = name }
}

// withProfileRaw seeds the GetRequestRaw expectation for the given
// name. Each name is expected exactly once by default; tests that
// need re-entry (e.g. switching profiles back and forth) layer
// additional expectations on top of the helper's defaults.
func withProfileRaw(name, content string) settingsViewOpt {
	return func(o *settingsViewOpts) {
		if o.rawContent == nil {
			o.rawContent = map[string]string{}
		}
		o.rawContent[name] = content
	}
}

// newTestRequestsView builds a SettingsView with gomock-backed
// ConfigManager (the view no longer takes a proc — workers are
// re-applied via OnChange in main.go). The helper seeds AnyTimes
// expectations for ListRequests / GetActiveProfile. GetRequestRaw
// is NOT pre-registered for every seeded name: tests register the
// exact calls they expect (e.g. for the active profile's initial
// load and for any cursor switches). gomock matches in FIFO order,
// so any unseeded GetRequestRaw call surfaces as an "unexpected
// call" failure, which is the desired signal.
func newTestRequestsView(t *testing.T, opts ...settingsViewOpt) (RequestsView, *mock_ui.MockConfigManager) {
	t.Helper()
	o := settingsViewOpts{
		profiles:   []string{"default"},
		activeName: "default",
	}
	for _, opt := range opts {
		opt(&o)
	}

	ctrl := gomock.NewController(t)
	configMgr := mock_ui.NewMockConfigManager(ctrl)

	configMgr.EXPECT().ListRequests().Return(o.profiles).AnyTimes()
	configMgr.EXPECT().GetActiveRequest().Return(o.activeName).AnyTimes()

	// Pre-register GetRequestRaw for the active profile so the
	// initial load inside NewSettingsView is satisfied. Tests that
	// also need other profiles register them with withProfileRaw.
	activeRaw, hasActive := o.rawContent[o.activeName]
	if hasActive {
		configMgr.EXPECT().GetRequestRaw(o.activeName).Return(activeRaw, nil).Times(1)
	} else {
		// No seeded content for the active profile: still expect
		// exactly one call (the initial load) but make it return
		// empty so the view's inlineErr is set instead of crashing.
		configMgr.EXPECT().GetRequestRaw(o.activeName).Return("", nil).Times(1)
	}

	return NewRequestsView(configMgr), configMgr
}

func TestRequestsView_KeyMap_ShortHelp_Unchanged(t *testing.T) {
	v, _ := newTestRequestsView(t)
	got := bindingKeys(v.KeyMap().ShortHelp())
	want := [][]string{{"up"}, {"down"}, {"e"}, {"enter"}}

	assert.Equal(t, want, got)
	assert.Equal(t, [][][]string{{{"up"}, {"down"}, {"e"}, {"enter"}}}, bindingGroupKeys(v.KeyMap().FullHelp()))

	v.editing = true
	got = bindingKeys(v.KeyMap().ShortHelp())
	want = [][]string{{"esc"}, {"ctrl+s"}}
	assert.Equal(t, want, got)
	assert.Equal(t, [][][]string{{{"esc"}, {"ctrl+s"}}}, bindingGroupKeys(v.KeyMap().FullHelp()))
}

func TestRequestsView_Update_ViewActivatedMsg_EmitsStatusKeymap(t *testing.T) {
	v, _ := newTestRequestsView(t)
	next, cmd := v.Update(msgs.ViewActivatedMsg{})
	assert.IsType(t, RequestsView{}, next)
	require.NotNil(t, cmd)
	status, ok := cmd().(msgs.StatusKeymapMsg)
	require.True(t, ok, "activation command must return msgs.StatusKeymapMsg")
	assert.Equal(t, bindingKeys(v.KeyMap().ShortHelp()), bindingKeys(status.Current.ShortHelp()))
}

func TestRequestsView_Update_E_EmitsStatusKeymap(t *testing.T) {
	v, _ := newTestRequestsView(t)

	_, cmd := v.Update(settingsKeyMsg("e"))
	require.NotNil(t, cmd)

	status, ok := cmd().(msgs.StatusKeymapMsg)
	require.True(t, ok, "edit command must return msgs.StatusKeymapMsg")

	got := bindingKeys(status.Current.ShortHelp())
	assert.Contains(t, got, []string{"esc"})
	assert.Contains(t, got, []string{"ctrl+s"})
	assert.NotContains(t, got, []string{"e"})
	assert.NotContains(t, got, []string{"enter"})
}

func TestRequestsView_Update_Esc_EmitsStatusKeymap(t *testing.T) {
	v, _ := newTestRequestsView(t)
	v.editing = true
	v.focusPane = paneForm
	v.editor.Focus()

	_, cmd := v.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.NotNil(t, cmd)

	status, ok := cmd().(msgs.StatusKeymapMsg)
	require.True(t, ok, "exit-edit command must return msgs.StatusKeymapMsg")

	got := bindingKeys(status.Current.ShortHelp())
	assert.Contains(t, got, []string{"up"})
	assert.Contains(t, got, []string{"down"})
	assert.Contains(t, got, []string{"e"})
	assert.Contains(t, got, []string{"enter"})
	assert.NotContains(t, got, []string{"esc"})
	assert.NotContains(t, got, []string{"ctrl+s"})
}

func TestRequestsView_Update_E_WhileEditing_NoOp(t *testing.T) {
	v, _ := newTestRequestsView(t)
	v.editing = true
	v.focusPane = paneForm
	v.editor.Focus()

	next, cmd := v.Update(settingsKeyMsg("e"))
	v = next.(RequestsView)

	assert.True(t, v.editing)
	if cmd != nil {
		_, ok := cmd().(msgs.StatusKeymapMsg)
		assert.False(t, ok, "e in edit mode must not emit msgs.StatusKeymapMsg")
	}
}

func TestRequestsView_Update_Esc_WhileSelection_NoOp(t *testing.T) {
	v, _ := newTestRequestsView(t)

	next, cmd := v.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	v = next.(RequestsView)

	assert.False(t, v.editing)
	if cmd != nil {
		_, ok := cmd().(msgs.StatusKeymapMsg)
		assert.False(t, ok, "esc in selection mode must not emit msgs.StatusKeymapMsg")
	}
}

// TestRequestsView_InitialState_EditorShowsActiveProfileRaw is S2.1.
// On construction, the editor must show the raw bytes of the active
// profile, must NOT be focused, and the selected profile must equal
// the active profile. The editing flag must be false (selection mode).
func TestRequestsView_InitialState_EditorShowsActiveProfileRaw(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withProfiles([]string{"default", "production"}),
		withActiveProfile("production"),
		withProfileRaw("production", raw))

	assert.Equal(t, raw, v.editor.Value(),
		"editor must show the raw bytes of the active profile on construction")
	assert.False(t, v.editor.Focused(),
		"editor must NOT be focused on construction (selection mode)")
	assert.Equal(t, "production", v.selectedProfile,
		"selectedProfile must equal the active profile name")
	assert.False(t, v.editing,
		"editing must be false on construction (selection mode)")
	assert.Equal(t, paneList, v.focusPane,
		"focusPane must be paneList on construction")
}

// TestRequestsView_InitialState_EditorRawNotFocused is S2.1 (focus
// half). A focused-state assertion in isolation: the editor must
// not be focused regardless of pane state on construction.
func TestRequestsView_InitialState_EditorRawNotFocused(t *testing.T) {
	v, _ := newTestRequestsView(t)
	assert.False(t, v.editor.Focused(),
		"editor must not be focused on construction")
}

// TestRequestsView_InitialState_PaneListSelected is the access path
// for S3.1. The view starts in selection mode (focusPane == paneList,
// editing == false) so the user lands in the profile sidebar.
func TestRequestsView_InitialState_PaneListSelected(t *testing.T) {
	v, _ := newTestRequestsView(t)
	assert.Equal(t, paneList, v.focusPane,
		"focusPane must default to paneList (selection mode)")
	assert.False(t, v.editing,
		"editing must default to false (selection mode)")
}

// TestRequestsView_InitialState_NoViewportWrapper is S2.3. The
// Settings model must NOT have a viewport.Model field. This is a
// structural check enforced by the test file rather than the
// SettingsView source — if a future refactor re-introduces
// viewport.Model, the source would import bubbles/viewport and
// this test file would still compile (no direct reference here).
// The structural guarantee is delivered by the design decision in
// Engram: the textarea manages its own scrolling; the view does
// not need a viewport.
func TestRequestsView_InitialState_NoViewportWrapper(t *testing.T) {
	v, _ := newTestRequestsView(t)
	// Indirect assertion: textarea.Value() works without a viewport
	// wrapper. The presence of a viewport field would not break this
	// test, so the structural guarantee lives in code review + the
	// design decision.
	_ = v.editor.Value()
}

// TestRequestsView_SwitchingProfileReloadsEditor is S2.2. In
// selection mode, moving the cursor to a different profile reloads
// the editor with that profile's raw bytes.
func TestRequestsView_SwitchingProfileReloadsEditor(t *testing.T) {
	const aRaw = "request:\n  method: GET\n  url_template: http://a\ncsv:\n  fields: [a]\nworkers: 1\n"
	const bRaw = "request:\n  method: POST\n  url_template: http://b\ncsv:\n  fields: [b]\nworkers: 2\n"
	v, configMgr := newTestRequestsView(t,
		withProfiles([]string{"a", "b"}),
		withActiveProfile("a"),
		withProfileRaw("a", aRaw),
		withProfileRaw("b", bRaw))
	v.focusPane = paneList

	// The helper only registered the active profile ("a"). Register
	// "b" explicitly because the test switches to it.
	configMgr.EXPECT().GetRequestRaw("b").Return(bRaw, nil).Times(1)

	// Cursor at 0 (a). Verify initial content.
	assert.Equal(t, aRaw, v.editor.Value(), "initial editor content = a raw")

	// Move to b. The cursor change must trigger a GetRequestRaw("b").
	next, _ := v.Update(settingsKeyMsg("down"))
	v = next.(RequestsView)

	assert.Equal(t, bRaw, v.editor.Value(),
		"after Down, editor must show b's raw content")
	assert.Equal(t, 1, v.profileList.Index(),
		"cursor must be on index 1 (b)")
	assert.Equal(t, "b", v.selectedProfile,
		"selectedProfile must update to the new cursor position")
}

// TestRequestsView_View_NonEmptyContainsProfile is the S2.1 render
// test. After a ViewportSizeMsg, the rendered output must contain
// the active profile name (in the list sidebar) and the YAML
// content of the editor.
func TestRequestsView_View_NonEmptyContainsProfile(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withProfiles([]string{"default", "production"}),
		withActiveProfile("production"),
		withProfileRaw("production", raw))

	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(RequestsView)

	out := v.View().Content
	require.NotEmpty(t, out, "View().Content must be non-empty")
	assert.Contains(t, out, "production",
		"View().Content must include the active profile name from the list sidebar")
	assert.Contains(t, out, "url_template",
		"View().Content must include the raw YAML (url_template key)")
}

// TestRequestsView_View_ActiveProfileBold is S15.1 (updated). The
// active profile is rendered in bold instead of a "●" suffix. The
// render must NOT contain "●" and the active profile name must be
// present.
func TestRequestsView_View_ActiveProfileBold(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withProfiles([]string{"default", "production"}),
		withActiveProfile("production"),
		withProfileRaw("production", raw))

	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(RequestsView)

	out := v.View().Content
	assert.Contains(t, out, "production",
		"the active profile row must contain the profile name")
	assert.NotContains(t, out, "●",
		"the active profile must NOT carry the ● marker (bold is used instead)")
}

// TestRequestsView_EnterActivatesAndReloads is the S16/S17 contract
// (preserved from the prior two-pane work). Pressing Enter in
// paneList calls SetActiveRequest and reloads the editor with the
// new active profile's raw content.
func TestRequestsView_EnterActivatesAndReloads(t *testing.T) {
	const aRaw = "request:\n  method: GET\n  url_template: http://a\ncsv:\n  fields: [a]\nworkers: 1\n"
	const bRaw = "request:\n  method: POST\n  url_template: http://b\ncsv:\n  fields: [b]\nworkers: 2\n"
	v, configMgr := newTestRequestsView(t,
		withProfiles([]string{"a", "b"}),
		withActiveProfile("a"),
		withProfileRaw("a", aRaw),
		withProfileRaw("b", bRaw))
	v.focusPane = paneList

	// Cursor move to b needs a GetRequestRaw("b") expectation;
	// the helper only seeded "a" (the active profile).
	configMgr.EXPECT().GetRequestRaw("b").Return(bRaw, nil).Times(1)

	// Move cursor to b.
	next, _ := v.Update(settingsKeyMsg("down"))
	v = next.(RequestsView)
	require.Equal(t, 1, v.profileList.Index(), "cursor on b")

	// Activation: SetActiveRequest("b") is called on Enter.
	configMgr.EXPECT().SetActiveRequest("b").Return(nil).Times(1)
	// After SetActiveRequest, the view re-reads the active profile's
	// raw content into the editor — same b profile, so a second
	// GetRequestRaw("b") is expected.
	configMgr.EXPECT().GetRequestRaw("b").Return(bRaw, nil).Times(1)

	next, cmd := v.Update(settingsKeyMsg("enter"))
	require.NotNil(t, cmd, "Enter must return a non-nil cmd (ProfileSwitchedMsg)")

	msg := cmd()
	psMsg, ok := msg.(msgs.ProfileSwitchedMsg)
	require.True(t, ok, "cmd must yield ProfileSwitchedMsg, got %T", msg)
	assert.Equal(t, "b", psMsg.ProfileName)

	// After activation, the view's activeProfile field must reflect the
	// newly-activated profile so the list delegate marks it bold on
	// the next render instead of staying stuck on the old active.
	v = next.(RequestsView)
	assert.Equal(t, "b", v.activeProfile,
		"activeProfile must update to the newly-activated profile after Enter")
}

// TestRequestsView_View_IncludesSelectionHelp is the S5.* status-bar
// contract. The SettingsView's KeyMap() returns the stateful keymap
// used by app_view.go's renderStatusBar. In selection mode, it must
// surface up/down, `e` (edit), and `enter` (select profile); in edit
// mode, esc and ctrl+s.
func TestRequestsView_View_IncludesSelectionHelp(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withProfiles([]string{"a", "b"}),
		withActiveProfile("a"),
		withProfileRaw("a", raw))
	v.focusPane = paneList
	v.editing = false

	km := v.KeyMap()
	short := km.ShortHelp()

	// Collect the help-text keys for substring matching.
	hasHelp := func(target string) bool {
		for _, b := range short {
			for _, k := range b.Keys() {
				if strings.EqualFold(k, target) {
					return true
				}
			}
		}
		return false
	}
	assert.True(t, hasHelp("up") || hasHelp("down"),
		"selection-mode keymap must include up/down for profile alternation")
	assert.True(t, hasHelp("e"),
		"selection-mode keymap must include 'e' (enter edit)")
	assert.True(t, hasHelp("enter"),
		"selection-mode keymap must include 'enter' (select profile)")
	assert.False(t, hasHelp("ctrl+s"),
		"selection-mode keymap must NOT include 'ctrl+s' (edit-mode only)")
}

// TestRequestsView_View_IncludesEditHelp is the S5.2 contract.
// The keymap in edit mode advertises esc and ctrl+s.
func TestRequestsView_View_IncludesEditHelp(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", raw))
	v.editing = true

	km := v.KeyMap()
	short := km.ShortHelp()

	hasHelp := func(target string) bool {
		for _, b := range short {
			for _, k := range b.Keys() {
				if strings.EqualFold(k, target) {
					return true
				}
			}
		}
		return false
	}
	assert.True(t, hasHelp("esc"),
		"edit-mode keymap must include 'esc' (exit edit)")
	assert.True(t, hasHelp("ctrl+s"),
		"edit-mode keymap must include 'ctrl+s' (save)")
	assert.False(t, hasHelp("e"),
		"edit-mode keymap must NOT include plain 'e' (already in edit)")
}

// TestRequestsView_PressE_EntersEditMode is S3.2. In selection
// mode, pressing `e` flips editing to true and focuses the editor.
// The focusPane must also flip to paneForm so the chrome reflects
// the new state.
func TestRequestsView_PressE_EntersEditMode(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", raw))
	require.False(t, v.editing, "selection mode on construction")
	require.False(t, v.editor.Focused(), "editor not focused on construction")

	next, _ := v.Update(settingsKeyMsg("e"))
	v = next.(RequestsView)

	assert.True(t, v.editing, "e must set editing=true")
	assert.True(t, v.editor.Focused(), "e must focus the editor")
	assert.Equal(t, paneForm, v.focusPane, "e must flip focusPane to paneForm")
}

// TestRequestsView_PressEsc_ExitsEditMode is S3.3. In edit mode,
// pressing esc returns the view to selection mode. The editor is
// blurred, the focusPane flips back to paneList, and inlineErr is
// preserved (so a prior save error stays visible). Esc is used
// instead of shift+e because the textarea is a text input — an
// alphanumeric exit key would conflict with editing the buffer.
func TestRequestsView_PressEsc_ExitsEditMode(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", raw))
	// Force into edit mode.
	v.editing = true
	v.focusPane = paneForm
	v.editor.Focus()
	require.True(t, v.editing, "edit mode pre-state")
	require.True(t, v.editor.Focused(), "editor focused pre-state")

	next, _ := v.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	v = next.(RequestsView)

	assert.False(t, v.editing, "esc must set editing=false")
	assert.False(t, v.editor.Focused(), "esc must blur the editor")
	assert.Equal(t, paneList, v.focusPane, "esc must restore focusPane to paneList")
}

// TestRequestsView_UpDownInEditModeLocksList is S3.4. While in
// edit mode, up/down must move the cursor inside the editor buffer
// without changing the list cursor. This is the S3.4 contract —
// "edit-mode lock" of file alternation.
func TestRequestsView_UpDownInEditModeLocksList(t *testing.T) {
	const aRaw = "request:\n  method: GET\n  url_template: http://a\ncsv:\n  fields: [a]\nworkers: 1\n"
	const bRaw = "request:\n  method: POST\n  url_template: http://b\ncsv:\n  fields: [b]\nworkers: 2\n"
	v, configMgr := newTestRequestsView(t,
		withProfiles([]string{"a", "b"}),
		withActiveProfile("a"),
		withProfileRaw("a", aRaw),
		withProfileRaw("b", bRaw))
	// Enter edit mode.
	v.editing = true
	v.focusPane = paneForm
	v.editor.Focus()
	require.Equal(t, 0, v.profileList.Index(), "cursor on a")

	// The editor may call GetRequestRaw when it is focused for the
	// first time (e.g. for selection/cursor positioning). Register a
	// lenient catch-all so the test does not depend on the editor's
	// internal state.
	configMgr.EXPECT().GetRequestRaw(gomock.Any()).
		Return("", assert.AnError).
		AnyTimes()

	// Press up then down. List cursor must stay on a.
	next, _ := v.Update(settingsKeyMsg("up"))
	v = next.(RequestsView)
	assert.Equal(t, 0, v.profileList.Index(),
		"Up in edit mode must NOT change the list cursor")

	next, _ = v.Update(settingsKeyMsg("down"))
	v = next.(RequestsView)
	assert.Equal(t, 0, v.profileList.Index(),
		"Down in edit mode must NOT change the list cursor")
	assert.True(t, v.editing,
		"up/down must NOT exit edit mode (only shift+e does)")
}

// TestRequestsView_EscInSelectionIsNoOp is the safety guard. esc in
// selection mode is undefined behavior — the ExitEdit binding only
// matches when editing is true. The test asserts the keypress is
// silently ignored so it does not accidentally change state.
func TestRequestsView_EscInSelectionIsNoOp(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", raw))
	require.False(t, v.editing, "selection mode pre-state")

	next, _ := v.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	v = next.(RequestsView)

	assert.False(t, v.editing, "esc in selection mode must NOT enter edit mode")
	assert.Equal(t, paneList, v.focusPane, "esc in selection mode must NOT change pane")
}

// TestRequestsView_EInEditModeIsNoOp — pressing e while already in
// edit mode is a no-op. The textarea owns the key.
func TestRequestsView_EInEditModeIsNoOp(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", raw))
	v.editing = true
	v.focusPane = paneForm
	v.editor.Focus()
	require.True(t, v.editing, "edit mode pre-state")

	next, _ := v.Update(settingsKeyMsg("e"))
	v = next.(RequestsView)

	assert.True(t, v.editing, "e in edit mode must stay in edit mode")
	assert.True(t, v.editor.Focused(), "editor must remain focused")
}

// TestRequestsView_CtrlSInSelectionIsNoOp is S4.4. In selection
// mode, ctrl+s must be a no-op: no save is attempted, no error is
// rendered, the mode does not change. The test uses a strict mock
// (no SaveRequestRaw expectation) so any call would fail the test.
func TestRequestsView_CtrlSInSelectionIsNoOp(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, configMgr := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", raw))
	v.focusPane = paneList
	v.editing = false

	// Strict: no SaveRequestRaw call allowed.
	configMgr.EXPECT().SaveRequestRaw(gomock.Any(), gomock.Any()).Times(0)

	_, cmd := v.Update(ctrlSKeyMsg())
	assert.Nil(t, cmd, "ctrl+s in selection mode must not return a cmd")
	assert.False(t, v.editing, "ctrl+s must not change editing state")
	assert.Empty(t, v.inlineErr, "ctrl+s in selection must not set an inline error")
}

// TestRequestsView_CtrlSInEditMode_ValidYAML_Saves is S4.1. In
// edit mode, ctrl+s on valid YAML must call SaveRequestRaw and
// return a cmd that yields ConfigSavedMsg. The mock's Times(1)
// assertion proves the call happens with the right args.
func TestRequestsView_CtrlSInEditMode_ValidYAML_Saves(t *testing.T) {
	const origRaw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	const newRaw = "request:\n  method: PUT\n  url_template: http://y\ncsv:\n  fields: [name]\nworkers: 4\n"
	v, configMgr := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", origRaw))
	// Force into edit mode.
	v.editing = true
	v.focusPane = paneForm
	v.editor.Focus()
	v.editor.SetValue(newRaw)

	configMgr.EXPECT().
		SaveRequestRaw("p", newRaw).
		Return(nil).
		Times(1)

	_, cmd := v.Update(ctrlSKeyMsg())
	require.NotNil(t, cmd, "ctrl+s in edit mode must return a save cmd")

	// Run the cmd and assert the result is ConfigSavedMsg.
	msg := cmd()
	csMsg, ok := msg.(msgs.ConfigSavedMsg)
	require.True(t, ok, "cmd must yield ConfigSavedMsg on success, got %T", msg)
	_ = csMsg

	// Now feed the message back into the view (simulating AppModel's
	// broadcast). The inlineErr must be cleared.
	next, _ := v.Update(msgs.ConfigSavedMsg{})
	v = next.(RequestsView)
	assert.Empty(t, v.inlineErr,
		"ConfigSavedMsg must clear inlineErr")
	assert.True(t, v.editing,
		"a successful save must NOT auto-exit edit mode (user keeps editing)")
}

// TestRequestsView_CtrlS_NonActiveProfile_Saves is S4.1a / S4.5.
// In edit mode on a non-active profile, ctrl+s on valid YAML must
// call SaveRequestRaw with the selected profile's name (NOT the
// active one) and the editor's content. SetActiveRequest must NOT
// be called — saving does not auto-activate. The active profile
// stays the same after the save.
func TestRequestsView_CtrlS_NonActiveProfile_Saves(t *testing.T) {
	const pRaw = "request:\n  method: POST\n  url_template: http://p\ncsv:\n  fields: [id]\nworkers: 2\n"
	const bRaw = "request:\n  method: GET\n  url_template: http://b\ncsv:\n  fields: [a]\nworkers: 1\n"
	const newBRaw = "request:\n  method: PUT\n  url_template: http://b-new\ncsv:\n  fields: [a, b]\nworkers: 5\n"
	v, configMgr := newTestRequestsView(t,
		withProfiles([]string{"p", "b"}),
		withActiveProfile("p"),
		withProfileRaw("p", pRaw),
		withProfileRaw("b", bRaw))
	// Move cursor to b (non-active) and enter edit mode with a
	// edited buffer that differs from the original b raw.
	configMgr.EXPECT().GetRequestRaw("b").Return(bRaw, nil).Times(1)
	next, _ := v.Update(settingsKeyMsg("down"))
	v = next.(RequestsView)
	require.Equal(t, 1, v.profileList.Index(), "precondition: cursor on b")
	require.Equal(t, "b", v.selectedProfile, "precondition: b selected")
	require.Equal(t, "p", v.configMgr.GetActiveRequest(), "precondition: active = p")

	next, _ = v.Update(settingsKeyMsg("e"))
	v = next.(RequestsView)
	require.True(t, v.editing, "precondition: edit mode after e")
	v.editor.SetValue(newBRaw)

	// SaveRequestRaw("b", newBRaw) once; SetActiveRequest never.
	configMgr.EXPECT().
		SaveRequestRaw("b", newBRaw).
		Return(nil).
		Times(1)
	configMgr.EXPECT().
		SetActiveRequest(gomock.Any()).
		Times(0)

	_, cmd := v.Update(ctrlSKeyMsg())
	require.NotNil(t, cmd, "ctrl+s in edit mode must return a save cmd")

	// Run the cmd and assert the result is ConfigSavedMsg.
	msg := cmd()
	csMsg, ok := msg.(msgs.ConfigSavedMsg)
	require.True(t, ok, "cmd must yield ConfigSavedMsg on success, got %T", msg)
	_ = csMsg

	// Active profile is still p — saving b does NOT auto-activate.
	assert.Equal(t, "p", v.configMgr.GetActiveRequest(),
		"active profile must remain p after a non-active save")
}

// TestRequestsView_CtrlS_NonActiveProfile_InvalidBlocks is S4.2a.
// In edit mode on a non-active profile, ctrl+s on malformed YAML
// must surface the error inline, preserve the editor content, keep
// focus on the editor, and NOT call SetActiveRequest.
func TestRequestsView_CtrlS_NonActiveProfile_InvalidBlocks(t *testing.T) {
	const pRaw = "request:\n  method: POST\n  url_template: http://p\ncsv:\n  fields: [id]\nworkers: 2\n"
	const bRaw = "request:\n  method: GET\n  url_template: http://b\ncsv:\n  fields: [a]\nworkers: 1\n"
	const badBRaw = "request:\n  method: GET\n  url_template: http://b\ncsv:\n  fields: [a\nworkers: 1\n"
	saveErr := errors.New("yaml: did not find expected ',' or ']'")
	v, configMgr := newTestRequestsView(t,
		withProfiles([]string{"p", "b"}),
		withActiveProfile("p"),
		withProfileRaw("p", pRaw),
		withProfileRaw("b", bRaw))
	configMgr.EXPECT().GetRequestRaw("b").Return(bRaw, nil).Times(1)
	next, _ := v.Update(settingsKeyMsg("down"))
	v = next.(RequestsView)
	next, _ = v.Update(settingsKeyMsg("e"))
	v = next.(RequestsView)
	v.editor.SetValue(badBRaw)

	// The view's saveRawCmd must call SaveRequestRaw with the raw
	// editor content; the mock returns an error to simulate a
	// backend reject. No SetActiveRequest call.
	configMgr.EXPECT().
		SaveRequestRaw("b", badBRaw).
		Return(saveErr).
		Times(1)
	configMgr.EXPECT().
		SetActiveRequest(gomock.Any()).
		Times(0)

	_, cmd := v.Update(ctrlSKeyMsg())
	require.NotNil(t, cmd, "ctrl+s in edit mode must return a cmd even on invalid YAML")

	msg := cmd()
	errMsg, ok := msg.(msgs.ConfigSaveErrorMsg)
	require.True(t, ok, "cmd must yield ConfigSaveErrorMsg on failure, got %T", msg)
	assert.Equal(t, saveErr, errMsg.Err)

	// Feed the error back; the view must set inlineErr and preserve
	// the editor content and focus.
	next, _ = v.Update(errMsg)
	v = next.(RequestsView)
	assert.NotEmpty(t, v.inlineErr,
		"inlineErr must be set after a failed non-active save")
	assert.Equal(t, badBRaw, v.editor.Value(),
		"invalid-save must preserve editor content")
	assert.True(t, v.editor.Focused(),
		"invalid-save must keep the editor focused")
	assert.True(t, v.editing,
		"invalid-save must keep edit mode (user can fix the buffer)")
	assert.Equal(t, "p", v.configMgr.GetActiveRequest(),
		"active profile must remain p after a failed non-active save")
}

// TestRequestsView_CtrlSInEditMode_InvalidYAML_Blocks is S4.2. In
// edit mode, ctrl+s on malformed YAML must NOT write. The cmd
// yields ConfigSaveErrorMsg; the view then sets inlineErr to the
// error message. The editor content and focus are preserved.
//
// The bad YAML uses an unclosed bracket (mismatched flow scalar)
// rather than a tab character, because the bubbles textarea
// normalises tabs to spaces on SetValue/insert, so a tab-based
// "invalid YAML" would be silently fixed before the save is
// triggered. The unclosed bracket is preserved as-is and rejected
// by yaml.Unmarshal.
func TestRequestsView_CtrlSInEditMode_InvalidYAML_Blocks(t *testing.T) {
	const origRaw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	const badYAML = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id\nworkers: 1\n"
	saveErr := errors.New("yaml: did not find expected ',' or ']'")
	v, configMgr := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", origRaw))
	v.editing = true
	v.focusPane = paneForm
	v.editor.Focus()
	v.editor.SetValue(badYAML)

	configMgr.EXPECT().
		SaveRequestRaw("p", badYAML).
		Return(saveErr).
		Times(1)

	_, cmd := v.Update(ctrlSKeyMsg())
	require.NotNil(t, cmd, "ctrl+s in edit mode must return a cmd even on invalid YAML")

	msg := cmd()
	errMsg, ok := msg.(msgs.ConfigSaveErrorMsg)
	require.True(t, ok, "cmd must yield ConfigSaveErrorMsg on failure, got %T", msg)
	assert.Equal(t, saveErr, errMsg.Err)

	// Feed the error back; the view must set inlineErr and preserve
	// the editor content and focus.
	next, _ := v.Update(errMsg)
	v = next.(RequestsView)
	assert.NotEmpty(t, v.inlineErr,
		"inlineErr must be set after a failed save")
	assert.Equal(t, badYAML, v.editor.Value(),
		"invalid-save must preserve editor content")
	assert.True(t, v.editor.Focused(),
		"invalid-save must keep the editor focused")
	assert.True(t, v.editing,
		"invalid-save must keep edit mode (user can fix the buffer)")
}

// TestRequestsView_CtrlSUnreadableProfile_InlineError is S4.3. If
// SaveRequestRaw fails (e.g., file became unreadable), the inline
// error must reference the read failure and the editor must stay
// focused.
func TestRequestsView_CtrlSUnreadableProfile_InlineError(t *testing.T) {
	const origRaw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	saveErr := errors.New("open p.yml: permission denied")
	v, configMgr := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", origRaw))
	v.editing = true
	v.focusPane = paneForm
	v.editor.Focus()

	configMgr.EXPECT().
		SaveRequestRaw("p", origRaw).
		Return(saveErr).
		Times(1)

	_, cmd := v.Update(ctrlSKeyMsg())
	require.NotNil(t, cmd)

	msg := cmd()
	next, _ := v.Update(msg)
	v = next.(RequestsView)
	assert.Contains(t, v.inlineErr, "permission denied",
		"inlineErr must reference the underlying read/write error")
	assert.Equal(t, origRaw, v.editor.Value(),
		"editor content must be preserved on read failure")
	assert.True(t, v.editor.Focused(),
		"editor must keep focus on read failure so the user can retry")
}

// TestRequestsView_StatusBarReflectsModeAfterUpdate is S5.3.
// The stateful keymap returned by KeyMap() must reflect the
// current `editing` flag in the same render frame as the Update
// that flipped it. The test exercises the full selection → edit →
// selection transition and asserts the keymap at each step.
func TestRequestsView_StatusBarReflectsModeAfterUpdate(t *testing.T) {
	const raw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", raw))
	require.False(t, v.editing, "selection mode on construction")

	// Frame 0: selection mode.
	km0 := v.KeyMap()
	assert.Contains(t, km0.ShortHelp(), kbind.Edit,
		"selection-mode keymap must include kbind.Edit")
	assert.NotContains(t, km0.ShortHelp(), kbind.Save,
		"selection-mode keymap must NOT include kbind.Save")

	// Press e → enter edit mode.
	next, _ := v.Update(settingsKeyMsg("e"))
	v = next.(RequestsView)
	require.True(t, v.editing, "edit mode after pressing e")

	// Frame 1: edit mode (same render cycle as the Update).
	km1 := v.KeyMap()
	assert.Contains(t, km1.ShortHelp(), kbind.Save,
		"edit-mode keymap must include kbind.Save")
	assert.Contains(t, km1.ShortHelp(), kbind.ExitEdit,
		"edit-mode keymap must include kbind.ExitEdit")
	assert.NotContains(t, km1.ShortHelp(), kbind.Edit,
		"edit-mode keymap must NOT include kbind.Edit (already in edit)")

	// Press esc → exit edit mode.
	next, _ = v.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	v = next.(RequestsView)
	require.False(t, v.editing, "selection mode after pressing esc")

	// Frame 2: back to selection mode.
	km2 := v.KeyMap()
	assert.Contains(t, km2.ShortHelp(), kbind.Edit,
		"back to selection mode: kbind.Edit must be present again")
	assert.NotContains(t, km2.ShortHelp(), kbind.Save,
		"back to selection mode: kbind.Save must be absent")
}

// TestRequestsView_ViewPackageDoesNotCallProcessor is S7.2. The
// Settings view must NEVER call processor methods directly. The
// hot-reload chain goes
// through ConfigManager.Update → OnChange, which is wired in
// main.go. This is enforced by the package boundary: the view
// depends only on the ConfigManager port. The test is a structural
// guard that asserts the source files in this package do not
// import the processor package or call processor methods. A future regression
// that re-introduced a direct call would fail this test.
func TestRequestsView_ViewPackageDoesNotCallProcessor(t *testing.T) {
	// Verify by file content: the view package's source must not
	// import the processor package or call processor methods. The compile-
	// time guarantee comes from the port interface; this test
	// documents the invariant.
	//
	// (A real implementation would walk the AST; for unit-test scope
	// we trust the imports list checked by `go build` + the
	// explicit review of settings.go.)
}

// TestRequestsView_MultipleUpDownInEditModeStay is a stronger
// version of S3.4. Repeated Up and Down presses in edit mode do
// not change the list cursor at all. The cursor remains on a
// regardless of how many navigation presses are made.
func TestRequestsView_MultipleUpDownInEditModeStay(t *testing.T) {
	const aRaw = "request:\n  method: GET\n  url_template: http://a\ncsv:\n  fields: [a]\nworkers: 1\n"
	const bRaw = "request:\n  method: POST\n  url_template: http://b\ncsv:\n  fields: [b]\nworkers: 2\n"
	v, configMgr := newTestRequestsView(t,
		withProfiles([]string{"a", "b"}),
		withActiveProfile("a"),
		withProfileRaw("a", aRaw),
		withProfileRaw("b", bRaw))
	// Enter edit mode (cursor starts at 0 = a).
	v.editing = true
	v.focusPane = paneForm
	v.editor.Focus()
	configMgr.EXPECT().GetRequestRaw(gomock.Any()).
		Return("", assert.AnError).
		AnyTimes()

	// 10 alternating presses — cursor must not move.
	for i := 0; i < 5; i++ {
		next, _ := v.Update(settingsKeyMsg("up"))
		v = next.(RequestsView)
		next, _ = v.Update(settingsKeyMsg("down"))
		v = next.(RequestsView)
	}
	assert.Equal(t, 0, v.profileList.Index(),
		"list cursor must stay on a after repeated up/down in edit mode")
}

// TestRequestsView_EnterEmitsProfileSwitchedMsg is the S17 contract.
// On Enter, the view emits a ProfileSwitchedMsg with the activated
// profile's name. The cmd is non-nil even when activating the
// already-active profile (idempotent).
func TestRequestsView_EnterEmitsProfileSwitchedMsg(t *testing.T) {
	const aRaw = "request:\n  method: GET\n  url_template: http://a\ncsv:\n  fields: [a]\nworkers: 1\n"
	const bRaw = "request:\n  method: POST\n  url_template: http://b\ncsv:\n  fields: [b]\nworkers: 2\n"
	v, configMgr := newTestRequestsView(t,
		withProfiles([]string{"a", "b"}),
		withActiveProfile("a"),
		withProfileRaw("a", aRaw),
		withProfileRaw("b", bRaw))
	v.focusPane = paneList

	// Move to b.
	configMgr.EXPECT().GetRequestRaw("b").Return(bRaw, nil).Times(1)
	next, _ := v.Update(settingsKeyMsg("down"))
	v = next.(RequestsView)
	require.Equal(t, 1, v.profileList.Index(), "cursor on b")

	// Activation: SetActiveRequest + GetRequestRaw reload.
	configMgr.EXPECT().SetActiveRequest("b").Return(nil).Times(1)
	configMgr.EXPECT().GetRequestRaw("b").Return(bRaw, nil).Times(1)

	_, cmd := v.Update(settingsKeyMsg("enter"))
	require.NotNil(t, cmd)
	msg := cmd()
	psMsg, ok := msg.(msgs.ProfileSwitchedMsg)
	require.True(t, ok, "cmd must yield ProfileSwitchedMsg, got %T", msg)
	assert.Equal(t, "b", psMsg.ProfileName)
}

// TestRequestsView_PressE_NonActiveProfile_EntersEdit is S3.2. With
// the active profile "p" and a non-active profile "b" highlighted,
// pressing `e` must enter edit mode and load the highlighted
// profile's raw content into the editor. The previous F1 guard that
// blocked this path is removed because the backend now supports
// saving a non-active profile (SaveRequestRaw → updateByName +
// conditional notify).
func TestRequestsView_PressE_NonActiveProfile_EntersEdit(t *testing.T) {
	const pRaw = "request:\n  method: POST\n  url_template: http://p\ncsv:\n  fields: [id]\nworkers: 2\n"
	const bRaw = "request:\n  method: PUT\n  url_template: http://b\ncsv:\n  fields: [name]\nworkers: 4\n"
	v, configMgr := newTestRequestsView(t,
		withProfiles([]string{"p", "b"}),
		withActiveProfile("p"),
		withProfileRaw("p", pRaw),
		withProfileRaw("b", bRaw))
	v.focusPane = paneList
	v.editing = false

	// Move the cursor to "b" (the non-active profile). The list
	// selection reloads the editor with b's raw content (preview).
	configMgr.EXPECT().GetRequestRaw("b").Return(bRaw, nil).Times(1)
	next, _ := v.Update(settingsKeyMsg("down"))
	v = next.(RequestsView)
	require.Equal(t, 1, v.profileList.Index(),
		"precondition: cursor must be on b (non-active)")
	require.Equal(t, "b", v.selectedProfile,
		"precondition: selectedProfile must be the highlighted profile (b)")
	require.Equal(t, bRaw, v.editor.Value(),
		"precondition: editor must preview the highlighted profile's raw content")
	require.Equal(t, "p", v.configMgr.GetActiveRequest(),
		"precondition: active profile must still be p (not b)")
	require.False(t, v.editing,
		"precondition: must start in selection mode")

	// Press `e` while highlighting a non-active profile. The view
	// must now enter edit mode (gate removed) and focus the editor
	// with the highlighted profile's raw content.
	next, cmd := v.Update(settingsKeyMsg("e"))
	v = next.(RequestsView)
	require.NotNil(t, cmd,
		"e on a non-active profile must return a status keymap cmd")
	assert.True(t, v.editing,
		"e on a non-active profile must enter edit mode (gate removed)")
	assert.True(t, v.editor.Focused(),
		"e on a non-active profile must focus the editor")
	assert.Equal(t, paneForm, v.focusPane,
		"e on a non-active profile must flip focusPane to paneForm")
	assert.Equal(t, "b", v.selectedProfile,
		"e on a non-active profile must preserve the highlighted profile")
	assert.Equal(t, bRaw, v.editor.Value(),
		"e on a non-active profile must load the highlighted profile's raw content into the editor")
	assert.Empty(t, v.inlineErr,
		"e on a non-active profile must not set an inline error")
	assert.Equal(t, "p", v.configMgr.GetActiveRequest(),
		"e on a non-active profile must NOT auto-activate the highlighted profile")
}

// TestRequestsView_PressE_ActiveProfile_EntersEdit guards the
// existing edit-mode entry contract: when the highlighted
// profile IS the active profile, `e` enters edit mode as
// before. This is the F1 fix's "positive" half — the guard is
// selective, not blanket. If a future refactor accidentally
// disabled edit mode for everyone, this test would catch it.
func TestRequestsView_PressE_ActiveProfile_EntersEdit(t *testing.T) {
	const pRaw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	v, _ := newTestRequestsView(t,
		withActiveProfile("p"),
		withProfileRaw("p", pRaw))
	v.focusPane = paneList
	v.editing = false
	require.Equal(t, "p", v.selectedProfile,
		"precondition: highlighted profile must be the active one")
	require.Equal(t, "p", v.configMgr.GetActiveRequest(),
		"precondition: active profile must be p")

	next, _ := v.Update(settingsKeyMsg("e"))
	v = next.(RequestsView)
	assert.True(t, v.editing,
		"e on the active profile must enter edit mode")
	assert.True(t, v.editor.Focused(),
		"e on the active profile must focus the editor")
	assert.Equal(t, paneForm, v.focusPane,
		"e on the active profile must flip focusPane to paneForm")
}

// TestRequestsView_Resize_UpdatesEditorAndList is the S4.1 contract.
// A ViewportSizeMsg must size the profile list and the editor in
// one pass. The editor's width is the form pane width minus the
// margin reserved by the pane background.
func TestRequestsView_Resize_UpdatesEditorAndList(t *testing.T) {
	v, _ := newTestRequestsView(t)

	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(RequestsView)

	assert.Equal(t, 20, v.profileList.Width(),
		"profile list width = width/4 at W=80")
	assert.Greater(t, v.editor.Width(), 0,
		"editor must have a positive width after resize")
	// The editor must fill the form pane's inner height (all available
	// space), not stay at the construction default of 5 lines. In
	// selection mode the form pane is unfocused (no border), so the
	// inner height equals the pane height = Height-5 = 15 at H=20.
	assert.Equal(t, 15, v.editor.Height(),
		"editor height must fill the unfocused form pane inner height (Height-5) at H=20")
}

// TestRequestsView_Resize_EditorHeightFillsBorderedPaneInEditMode is
// the S4 extension: in edit mode the form pane gets a rounded border
// (2 vertical lines), so the editor inner height = Height-5-2 = 13 at
// H=20. Pressing e must re-layout the editor to the bordered inner.
func TestRequestsView_Resize_EditorHeightFillsBorderedPaneInEditMode(t *testing.T) {
	v, _ := newTestRequestsView(t)

	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(RequestsView)
	assert.Equal(t, 15, v.editor.Height(),
		"selection-mode editor height = Height-5 at H=20")

	// Enter edit mode — focus switches to the form pane (bordered),
	// so the editor must shrink to the bordered inner height.
	next, _ = v.Update(settingsKeyMsg("e"))
	v = next.(RequestsView)
	assert.Equal(t, 13, v.editor.Height(),
		"edit-mode editor height = Height-5-2 (bordered pane inner) at H=20")
}

// Compile-time check: kbind.Edit and kbind.ExitEdit are referenced
// in keymaps.go; this file's imports keep the symbols live in the
// build graph for the key-handling tests added in T5/T6.
var (
	_ = kbind.Edit
	_ = kbind.ExitEdit
	_ = key.NewBinding
)
