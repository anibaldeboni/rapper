package views

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui/components"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	"github.com/anibaldeboni/rapper/internal/utils"
)

var (
	settingsTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1).Bold(true)
	labelStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	focusedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Background(styles.FocusedPaneBg)
	helpStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	headerStyle        = lipgloss.NewStyle().MarginTop(1)
	inputStyle         = lipgloss.NewStyle().MarginBottom(1)
	profileBadgeStyle  = lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("230")).Padding(0, 1).MarginLeft(2)
)

// focus panes — the two-pane layout introduced by the persistent
// profile sidebar. focusPane == paneList means keystrokes go to the
// profile list; focusPane == paneForm means they go to the form
// fields. Tab toggles between them.
const (
	paneList = 0
	paneForm = 1
)

// sidebar width clamps. The sidebar is sized to 25 % of the viewport
// width, but the result is clamped to [minListWidth, maxListWidth] so
// profile names stay readable on narrow terminals and the sidebar
// doesn't run away on wide ones.
const (
	minListWidth = 15
	maxListWidth = 30
)

// focusable fields
const (
	sliderField = iota
	urlField
	methodField
	bodyField
	headersField
	csvFieldsField
	maxFields
)

// SettingsView displays and edits configuration settings. It is a
// value-type tea.Model so AppModel can store it behind the unified
// viewModel type. Callers must capture the value returned by Update
// to preserve state.
type SettingsView struct {
	configMgr ports.ConfigManager
	proc      ports.ProcessorController
	width     int
	height    int
	viewport  viewport.Model

	// Worker count slider (above the form fields). Value-typed — the
	// historical *components.Slider pointer was a relic of the
	// pre-tea.Model design.
	slider components.Slider

	// Form fields
	urlInput       textinput.Model
	methodInput    textinput.Model
	bodyInput      textarea.Model
	headersInput   textarea.Model
	csvFieldsInput textarea.Model

	// Persistent profile sidebar. Always visible; cursor drives
	// preview/activation. The bubbles list is a value type — its
	// pointer-receiver Update/Select return the modified copy that
	// the SettingsView captures locally before reassigning.
	profileList list.Model

	// Focus management
	focused   int
	focusPane int
	focusable []int

	// State
	modified bool
}

// Compile-time guard: SettingsView must satisfy tea.Model with a value
// receiver. Phase 3 converts the historical pointer-receiver surface
// to value receivers; this assertion fails at build time if the
// conversion regresses.
var _ tea.Model = SettingsView{}

// NewSettingsView creates a new SettingsView. The proc controller is required
// because the worker-count slider at the top of the view mutates the runtime
// processor immediately on change.
func NewSettingsView(configMgr ports.ConfigManager, proc ports.ProcessorController) SettingsView {
	// Create URL input
	urlInput := textinput.New()
	urlInput.Placeholder = "http://localhost:8080/api/v1/users"
	urlInput.CharLimit = 500
	urlInput.Prompt = ""

	// Create method input
	methodInput := textinput.New()
	methodInput.Placeholder = "POST"
	methodInput.CharLimit = 10
	methodInput.Prompt = ""

	// Create body textarea
	bodyInput := textarea.New()
	bodyInput.Placeholder = `{"name": "{{.name}}", "email": "{{.email}}"}`
	bodyInput.CharLimit = 5000
	bodyInput.SetHeight(5)
	bodyInput.ShowLineNumbers = false

	// Create headers textarea
	headersInput := textarea.New()
	headersInput.Placeholder = `Content-Type: application/json
Authorization: Bearer {{.token}}`
	headersInput.CharLimit = 2000
	headersInput.SetHeight(5)
	headersInput.ShowLineNumbers = false

	// Create CSV fields textarea
	csvFieldsInput := textarea.New()
	csvFieldsInput.Placeholder = `id
name
email`
	csvFieldsInput.CharLimit = 1000
	csvFieldsInput.SetHeight(4)
	csvFieldsInput.ShowLineNumbers = false

	initial := proc.GetWorkerCount()
	slider := components.NewSlider("Worker Count", 1, proc.GetMaxWorkers(), initial)

	profileNames := configMgr.ListProfiles()
	items := make([]list.Item, len(profileNames))
	for i, name := range profileNames {
		items[i] = Option[string]{Value: name, Title: name}
	}
	profileList := list.New(items, profileItemDelegate{active: configMgr.GetActiveProfile()}, 0, 0)
	profileList.InfiniteScrolling = true
	profileList.SetShowStatusBar(false)
	profileList.SetShowPagination(false)
	profileList.SetShowTitle(false)
	profileList.SetFilteringEnabled(false)
	profileList.SetShowHelp(false)
	profileList.DisableQuitKeybindings()
	profileList.KeyMap.CursorUp = kbind.Up
	profileList.KeyMap.CursorDown = kbind.Down
	if activeIdx := indexOf(profileNames, configMgr.GetActiveProfile()); activeIdx >= 0 {
		profileList.Select(activeIdx)
	}

	v := SettingsView{
		configMgr:      configMgr,
		proc:           proc,
		slider:         *slider,
		profileList:    profileList,
		urlInput:       urlInput,
		methodInput:    methodInput,
		bodyInput:      bodyInput,
		headersInput:   headersInput,
		csvFieldsInput: csvFieldsInput,
		focused:        sliderField,
		focusPane:      paneList,
		focusable:      []int{sliderField, urlField, methodField, bodyField, headersField, csvFieldsField},
		viewport:       viewport.New(viewport.WithWidth(0), viewport.WithHeight(0)),
	}

	v = v.loadConfig()
	v = v.updateFocus()

	return v
}

// indexOf returns the index of needle in haystack, or -1 if absent.
// Used to seed the profile list's cursor at the active profile.
func indexOf(haystack []string, needle string) int {
	for i, s := range haystack {
		if s == needle {
			return i
		}
	}
	return -1
}

// profileItemDelegate renders a single row of the profile sidebar.
// The cursor row gets a "▶ " prefix; the active profile row gets
// a " ●" suffix. The two decorations are independent so the
// active profile on the cursor row shows both ("▶ name ●").
type profileItemDelegate struct {
	active string // injected at construction; the name of the active profile
}

func (d profileItemDelegate) Height() int                             { return 1 }
func (d profileItemDelegate) Spacing() int                            { return 0 }
func (d profileItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d profileItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	opt, ok := item.(Option[string])
	if !ok {
		return
	}
	prefix := "  "
	if index == m.Index() {
		prefix = "▶ "
	}
	tag := ""
	if opt.Value == d.active {
		tag = " ●"
	}
	fmt.Fprintf(w, "%s%s%s", prefix, opt.Value, tag)
}

// FocusedPane returns the currently focused pane constant. Exposed
// for test assertions (S-9.1, S-9.2) and for AppModel state
// inspection; the value is one of {paneList, paneForm}.
func (v SettingsView) FocusedPane() int { return v.focusPane }

// Init returns nil per R-6.
func (v SettingsView) Init() tea.Cmd { return nil }

// loadConfig loads the current configuration into the form fields.
// Operates on a value-receiver copy and returns the modified
// SettingsView so callers can preserve the populated inputs. Callers
// MUST capture the returned value — the value receiver means the
// original struct is never mutated.
func (v SettingsView) loadConfig() SettingsView {
	return v.loadFromConfig(v.configMgr.Get())
}

// loadFromConfig is the single form-population code path shared by
// activation (loadConfig → Get) and preview (previewProfile →
// GetProfile). Returns the modified SettingsView. Callers MUST
// capture the returned value (value-receiver contract). A nil cfg
// is a no-op (the form keeps its current values).
func (v SettingsView) loadFromConfig(cfg *config.Config) SettingsView {
	if cfg == nil {
		return v
	}

	v.urlInput.SetValue(cfg.Request.URLTemplate)

	// Set method (default to POST if empty)
	method := cfg.Request.Method
	if method == "" {
		method = "POST"
	}
	v.methodInput.SetValue(method)

	v.bodyInput.SetValue(cfg.Request.BodyTemplate)

	// Convert headers map to string
	var headerLines []string
	for key, value := range cfg.Request.Headers {
		headerLines = append(headerLines, fmt.Sprintf("%s: %s", key, value))
	}
	v.headersInput.SetValue(strings.Join(headerLines, "\n"))

	// Convert CSV fields slice to string
	v.csvFieldsInput.SetValue(strings.Join(cfg.CSV.Fields, "\n"))

	return v
}

// parseHeaders converts headers string to map
func parseHeaders(headersText string) map[string]string {
	headers := make(map[string]string)
	lines := strings.SplitSeq(headersText, "\n")

	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}

	return headers
}

// parseCSVFields converts CSV fields string to slice
func parseCSVFields(fieldsText string) []string {
	var fields []string
	lines := strings.SplitSeq(fieldsText, "\n")

	for line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			fields = append(fields, line)
		}
	}

	return fields
}

// saveConfig saves the current form values to configuration
func (v SettingsView) saveConfig() error {
	cfg := v.configMgr.Get()
	if cfg == nil {
		cfg = &config.Config{}
	}

	// Update config from form fields
	cfg.Request.URLTemplate = v.urlInput.Value()
	cfg.Request.Method = v.methodInput.Value()
	cfg.Request.BodyTemplate = v.bodyInput.Value()
	cfg.Request.Headers = parseHeaders(v.headersInput.Value())
	cfg.CSV.Fields = parseCSVFields(v.csvFieldsInput.Value())

	// Update and save
	if err := v.configMgr.Update(cfg); err != nil {
		return err
	}

	if err := v.configMgr.Save(); err != nil {
		return err
	}

	return nil
}

// updateFocus updates the focus state of all inputs and returns the
// modified copy.
func (v SettingsView) updateFocus() SettingsView {
	v.urlInput.Blur()
	v.methodInput.Blur()
	v.bodyInput.Blur()
	v.headersInput.Blur()
	v.csvFieldsInput.Blur()
	v.slider.Focused = false

	switch v.focused {
	case sliderField:
		v.slider.Focused = true
	case urlField:
		v.urlInput.Focus()
	case methodField:
		v.methodInput.Focus()
	case bodyField:
		v.bodyInput.Focus()
	case headersField:
		v.headersInput.Focus()
	case csvFieldsField:
		v.csvFieldsInput.Focus()
	}
	return v
}

// nextFormField moves focus to the next field and returns the modified copy.
func (v SettingsView) nextField() SettingsView {
	v.focused++
	if v.focused >= maxFields {
		v.focused = 0
	}
	return v.updateFocus()
}

func (v SettingsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case msgs.ViewportSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

		listW := utils.Clamp(msg.Width/4, minListWidth, maxListWidth)
		formW := msg.Width - listW
		v.profileList.SetSize(listW, msg.Height-6)
		v.viewport.SetWidth(formW - 4)
		v.viewport.SetHeight(msg.Height - 6)
		v.urlInput.SetWidth(formW - 6)
		v.methodInput.SetWidth(formW - 6)
		v.bodyInput.SetWidth(formW - 6)
		v.headersInput.SetWidth(formW - 6)
		v.csvFieldsInput.SetWidth(formW - 6)
		return v, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, kbind.Save):
			return v, v.saveConfigCmd()

		case key.Matches(msg, kbind.PaneToggle):
			// Tab toggles the focus pane: paneList ↔ paneForm. It
			// does NOT cycle the focused form field
			v.focusPane = paneForm - v.focusPane
			return v, nil

		case key.Matches(msg, kbind.PrevField):
			if v.focusPane == paneForm {
				return v.nextField(), nil
			}
			return v, nil

		// Handle viewport scrolling
		case key.Matches(msg, kbind.PageUp):
			v.viewport.PageUp()
			return v, nil
		case key.Matches(msg, kbind.PageDown):
			v.viewport.PageDown()
			return v, nil
		case key.Matches(msg, kbind.GotoTop):
			v.viewport.GotoTop()
			return v, nil
		case key.Matches(msg, kbind.GotoBottom):
			v.viewport.GotoBottom()
			return v, nil
		}

		if v.focusPane == paneList {
			oldIdx := v.profileList.Index()
			v.modified = false
			var listCmd tea.Cmd
			v.profileList, listCmd = v.profileList.Update(msg)
			if v.profileList.Index() != oldIdx {
				if opt, ok := v.profileList.SelectedItem().(Option[string]); ok {
					v = v.previewProfile(opt.Value)
				}
			}
			// Enter activates the selected profile.
			if key.Matches(msg, kbind.Select) {
				if opt, ok := v.profileList.SelectedItem().(Option[string]); ok {
					return v.activateProfile(opt.Value)
				}
				return v, listCmd
			}
			return v, listCmd
		}
	}

	// Update focused field
	var cmd tea.Cmd
	oldValue := ""

	switch v.focused {
	case sliderField:
		prev := v.slider.Value
		updated, _ := v.slider.Update(msg)
		v.slider = updated
		if v.slider.Value != prev {
			v.proc.SetWorkers(v.slider.Value)
			v.modified = true
		}
		return v, nil
	case urlField:
		oldValue = v.urlInput.Value()
		v.urlInput, cmd = v.urlInput.Update(msg)
		if v.urlInput.Value() != oldValue {
			v.modified = true
		}

	case methodField:
		oldValue = v.methodInput.Value()
		v.methodInput, cmd = v.methodInput.Update(msg)
		if v.methodInput.Value() != oldValue {
			v.modified = true
		}

	case bodyField:
		oldValue = v.bodyInput.Value()
		v.bodyInput, cmd = v.bodyInput.Update(msg)
		if v.bodyInput.Value() != oldValue {
			v.modified = true
		}

	case headersField:
		oldValue = v.headersInput.Value()
		v.headersInput, cmd = v.headersInput.Update(msg)
		if v.headersInput.Value() != oldValue {
			v.modified = true
		}

	case csvFieldsField:
		oldValue = v.csvFieldsInput.Value()
		v.csvFieldsInput, cmd = v.csvFieldsInput.Update(msg)
		if v.csvFieldsInput.Value() != oldValue {
			v.modified = true
		}
	}

	return v, cmd
}

func (v SettingsView) View() tea.View {
	listW := min(max(v.width/4, minListWidth), maxListWidth)
	formW := v.width - listW

	listBg := v.paneBg(v.focusPane == paneList)
	listTitle := headerStyle.MarginLeft(1).Render(settingsTitleStyle.Render("📋 Profiles"))
	listPane := lipgloss.JoinVertical(
		lipgloss.Top,
		listTitle,
		listBg.
			Width(listW).
			Height(v.height-3).
			Render(v.profileList.View()),
	)

	var help string
	if v.modified {
		help = "⚠️  Unsaved changes"
	}
	formContent := lipgloss.JoinVertical(
		lipgloss.Top,
		inputStyle.Render(v.slider.View()),
		v.renderInput(urlField, "URL template:", v.urlInput),
		v.renderInput(methodField, "Method:", v.methodInput),
		v.renderTextArea(bodyField, "Body template:", v.bodyInput),
		v.renderTextArea(headersField, "Headers:", v.headersInput),
		v.renderTextArea(csvFieldsField, "CSV Fields (one per line):", v.csvFieldsInput),
		helpStyle.Render(help),
	)
	v.viewport.SetContent(formContent)

	formBg := v.paneBg(v.focusPane == paneForm)

	header := headerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			settingsTitleStyle.Render("⚙️ Settings"),
			profileBadgeStyle.Render("📋 "+v.getActiveProfileName()),
		),
	)
	formPane := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		formBg.
			Width(formW).
			Height(v.height-3).
			Render(v.viewport.View()),
	)

	content := lipgloss.JoinHorizontal(lipgloss.Left, listPane, formPane)
	return tea.NewView(content)
}

func (v SettingsView) paneBg(focused bool) lipgloss.Style {
	if focused {
		return lipgloss.NewStyle().BorderForeground(styles.FocusedPaneBg).BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1).MarginRight(1)
	}
	return lipgloss.NewStyle().MarginTop(1).MarginLeft(1).MarginRight(0).Padding(0, 1)
}

func (v SettingsView) renderTextArea(fieldIdx int, text string, input textarea.Model) string {
	label := v.renderLabel(text, fieldIdx)
	return inputStyle.Render(lipgloss.JoinVertical(lipgloss.Left, label, input.View()))
}

func (v SettingsView) renderInput(fieldIdx int, text string, input textinput.Model) string {
	label := v.renderLabel(text, fieldIdx)
	return inputStyle.Render(lipgloss.JoinVertical(lipgloss.Left, label, input.View()))
}

// renderLabel renders a label with focus indication
func (v SettingsView) renderLabel(text string, fieldIdx int) string {
	if v.focused == fieldIdx {
		return focusedStyle.Render("▶ " + text)
	}
	return labelStyle.Render(text)
}

// getActiveProfileName returns the name of the active profile
func (v SettingsView) getActiveProfileName() string {
	profileName := v.configMgr.GetActiveProfile()
	if profileName == "" {
		return "default"
	}
	return profileName
}

// previewProfile loads the named profile's config into the form
// fields without making it active. The list cursor already moved
// before this is called; this is just the form-repopulation step.
// Returns the modified SettingsView. A nil config (unknown name)
// is a no-op so the form keeps its current values.
func (v SettingsView) previewProfile(name string) SettingsView {
	return v.loadFromConfig(v.configMgr.GetProfile(name))
}

// activateProfile switches to the named profile, reloads the
// form, and emits a ProfileSwitchedMsg / ProfileSwitchErrorMsg.
// It does NOT call Save or Update — activation is a runtime
// profile switch, not a config write (per R-18).
func (v SettingsView) activateProfile(name string) (SettingsView, tea.Cmd) {
	if err := v.configMgr.SetActiveProfile(name); err != nil {
		return v, func() tea.Msg {
			return msgs.ProfileSwitchErrorMsg{Err: err}
		}
	}

	// Reload configuration from the new active profile.
	v = v.loadConfig()
	// Note: v.modified is not reset here because the value
	// receiver loses the assignment on return. See saveConfig
	// for context.

	return v, func() tea.Msg {
		return msgs.ProfileSwitchedMsg{ProfileName: name}
	}
}

// saveConfigCmd saves the configuration and returns a command
func (v SettingsView) saveConfigCmd() tea.Cmd {
	if err := v.saveConfig(); err != nil {
		return func() tea.Msg {
			return msgs.ConfigSaveErrorMsg{Err: err}
		}
	}

	return func() tea.Msg {
		return msgs.ConfigSavedMsg{}
	}
}
