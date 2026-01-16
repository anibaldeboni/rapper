package views

import (
	"fmt"
	"strings"

	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	settingsTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1).Bold(true)
	settingsAppStyle   = lipgloss.NewStyle().Margin(1, 2)
	labelStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	focusedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	helpStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	headerStyle        = lipgloss.NewStyle().MarginBottom(1)
	inputStyle         = lipgloss.NewStyle().MarginBottom(1)
	profileBadgeStyle  = lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("230")).Padding(0, 1).MarginLeft(2)
)

// focusable fields
const (
	urlField = iota
	methodField
	bodyField
	headersField
	csvFieldsField
	maxFields
)

// SettingsView displays and edits configuration settings
type SettingsView struct {
	configMgr config.Manager
	width     int
	height    int

	// Form fields
	urlInput       textinput.Model
	methodInput    textinput.Model
	bodyInput      textarea.Model
	headersInput   textarea.Model
	csvFieldsInput textarea.Model

	// Focus management
	focused   int
	focusable []int

	// Profile selector
	showProfileSelector bool
	profileListIndex    int

	// State
	modified bool
}

// NewSettingsView creates a new SettingsView
func NewSettingsView(configMgr config.Manager) *SettingsView {
	// Create URL input
	urlInput := textinput.New()
	urlInput.Placeholder = "http://localhost:8080/api/v1/users"
	urlInput.CharLimit = 500
	// urlInput.Width = 80
	urlInput.Prompt = ""

	// Create method input
	methodInput := textinput.New()
	methodInput.Placeholder = "POST"
	methodInput.CharLimit = 10
	// methodInput.Width = 20
	methodInput.Prompt = ""

	// Create body textarea
	bodyInput := textarea.New()
	bodyInput.Placeholder = `{"name": "{{.name}}", "email": "{{.email}}"}`
	bodyInput.CharLimit = 5000
	bodyInput.SetHeight(5)
	// bodyInput.SetWidth(80)
	bodyInput.ShowLineNumbers = false

	// Create headers textarea
	headersInput := textarea.New()
	headersInput.Placeholder = `Content-Type: application/json
Authorization: Bearer {{.token}}`
	headersInput.CharLimit = 2000
	headersInput.SetHeight(5)
	// headersInput.SetWidth(80)
	headersInput.ShowLineNumbers = false

	// Create CSV fields textarea
	csvFieldsInput := textarea.New()
	csvFieldsInput.Placeholder = `id
name
email`
	csvFieldsInput.CharLimit = 1000
	csvFieldsInput.SetHeight(4)
	// csvFieldsInput.SetWidth(80)
	csvFieldsInput.ShowLineNumbers = false

	v := &SettingsView{
		configMgr:      configMgr,
		urlInput:       urlInput,
		methodInput:    methodInput,
		bodyInput:      bodyInput,
		headersInput:   headersInput,
		csvFieldsInput: csvFieldsInput,
		focused:        urlField,
		focusable:      []int{urlField, methodField, bodyField, headersField, csvFieldsField},
	}

	// Load current configuration
	v.loadConfig()

	// Set initial focus
	v.updateFocus()

	return v
}

// loadConfig loads the current configuration into the form fields
func (v *SettingsView) loadConfig() {
	cfg := v.configMgr.Get()
	if cfg == nil {
		return
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
}

// parseHeaders converts headers string to map
func parseHeaders(headersText string) map[string]string {
	headers := make(map[string]string)
	lines := strings.Split(headersText, "\n")

	for _, line := range lines {
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
	lines := strings.Split(fieldsText, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			fields = append(fields, line)
		}
	}

	return fields
}

// saveConfig saves the current form values to configuration
func (v *SettingsView) saveConfig() error {
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

	v.modified = false
	return nil
}

// updateFocus updates the focus state of all inputs
func (v *SettingsView) updateFocus() {
	v.urlInput.Blur()
	v.methodInput.Blur()
	v.bodyInput.Blur()
	v.headersInput.Blur()
	v.csvFieldsInput.Blur()

	switch v.focused {
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
}

// nextField moves focus to the next field
func (v *SettingsView) nextField() {
	v.focused = (v.focused + 1) % maxFields
	v.updateFocus()
}

// prevField moves focus to the previous field
func (v *SettingsView) prevField() {
	v.focused--
	if v.focused < 0 {
		v.focused = maxFields - 1
	}
	v.updateFocus()
}

// Update handles messages for the settings view
func (v *SettingsView) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if v.showProfileSelector {
			profiles := v.getProfiles()
			switch {
			case key.Matches(msg, kbind.Up):
				v.profileListIndex--
				if v.profileListIndex < 0 {
					v.profileListIndex = len(profiles) - 1
				}
				return nil

			case key.Matches(msg, kbind.Down):
				v.profileListIndex++
				if v.profileListIndex >= len(profiles) {
					v.profileListIndex = 0
				}
				return nil

			case key.Matches(msg, kbind.Select):
				// Switch to selected profile
				if v.profileListIndex >= 0 && v.profileListIndex < len(profiles) {
					v.showProfileSelector = false
					return v.switchProfile(profiles[v.profileListIndex])
				}
				v.showProfileSelector = false
				return nil

			case key.Matches(msg, kbind.Cancel):
				v.showProfileSelector = false
				return nil
			}
			return nil
		}

		// Handle keyboard shortcuts
		switch {
		case key.Matches(msg, kbind.Save):
			// Save configuration
			return v.saveConfigCmd()

		case key.Matches(msg, kbind.Profile):
			// Toggle profile selector
			v.showProfileSelector = !v.showProfileSelector
			// Set initial selection to current profile
			profiles := v.getProfiles()
			activeProfile := v.getActiveProfileName()
			for i, name := range profiles {
				if name == activeProfile {
					v.profileListIndex = i
					break
				}
			}
			return nil

		case key.Matches(msg, kbind.NextField):
			v.nextField()
			return nil

		case key.Matches(msg, kbind.PrevField):
			v.prevField()
			return nil
		}
	}

	// Update focused field
	var cmd tea.Cmd
	oldValue := ""

	switch v.focused {
	case urlField:
		oldValue = v.urlInput.Value()
		v.urlInput, cmd = v.urlInput.Update(msg)
		if v.urlInput.Value() != oldValue {
			v.modified = true
		}
		cmds = append(cmds, cmd)

	case methodField:
		oldValue = v.methodInput.Value()
		v.methodInput, cmd = v.methodInput.Update(msg)
		if v.methodInput.Value() != oldValue {
			v.modified = true
		}
		cmds = append(cmds, cmd)

	case bodyField:
		oldValue = v.bodyInput.Value()
		v.bodyInput, cmd = v.bodyInput.Update(msg)
		if v.bodyInput.Value() != oldValue {
			v.modified = true
		}
		cmds = append(cmds, cmd)

	case headersField:
		oldValue = v.headersInput.Value()
		v.headersInput, cmd = v.headersInput.Update(msg)
		if v.headersInput.Value() != oldValue {
			v.modified = true
		}
		cmds = append(cmds, cmd)

	case csvFieldsField:
		oldValue = v.csvFieldsInput.Value()
		v.csvFieldsInput, cmd = v.csvFieldsInput.Update(msg)
		if v.csvFieldsInput.Value() != oldValue {
			v.modified = true
		}
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// Resize updates the view dimensions
func (v *SettingsView) Resize(width, height int) {
	v.width = width
	v.height = height
}

// View renders the settings view
func (v *SettingsView) View() string {
	// Show profile selector modal if active
	if v.showProfileSelector {
		return v.renderWithProfileSelector()
	}

	// Help text
	var help string
	if v.modified {
		help = "⚠️  Unsaved changes"
	}

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		headerStyle.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				settingsTitleStyle.Render("⚙️ Settings"),
				profileBadgeStyle.Render(fmt.Sprintf("📋 %s", v.getActiveProfileName())),
			),
		),
		v.renderInput(urlField, "URL template:", v.urlInput),
		v.renderInput(methodField, "Method:", v.methodInput),
		v.renderTextArea(bodyField, "Body template:", v.bodyInput),
		v.renderTextArea(headersField, "Headers:", v.headersInput),
		v.renderTextArea(csvFieldsField, "CSV Fields (one per line):", v.csvFieldsInput),
		helpStyle.Render(help),
	)

	return settingsAppStyle.Render(content)
}
func (v *SettingsView) renderTextArea(fieldIdx int, text string, input textarea.Model) string {
	label := v.renderLabel(text, fieldIdx)
	return inputStyle.Render(lipgloss.JoinVertical(lipgloss.Left, label, input.View()))
}

func (v *SettingsView) renderInput(fieldIdx int, text string, input textinput.Model) string {
	label := v.renderLabel(text, fieldIdx)
	return inputStyle.Render(lipgloss.JoinVertical(lipgloss.Left, label, input.View()))
}

// renderLabel renders a label with focus indication
func (v *SettingsView) renderLabel(text string, fieldIdx int) string {
	if v.focused == fieldIdx {
		return focusedStyle.Render("▶ " + text)
	}
	return labelStyle.Render(text)
}

// renderWithProfileSelector renders the profile selector modal overlay
func (v *SettingsView) renderWithProfileSelector() string {
	profiles := v.getProfiles()
	activeProfile := v.getActiveProfileName()

	// Build profile list
	var profileList strings.Builder

	// Styles for profile items
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	activeTagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("40")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	for i, name := range profiles {
		var line string
		isActive := name == activeProfile
		isSelected := i == v.profileListIndex

		// Add selection indicator
		if isSelected {
			line = "▶ "
		} else {
			line = "  "
		}

		// Add profile name with style
		if isSelected {
			line += selectedStyle.Render(name)
		} else {
			line += normalStyle.Render(name)
		}

		// Add active badge
		if isActive {
			line += " " + activeTagStyle.Render("●")
		}

		profileList.WriteString(line)
		profileList.WriteString("\n")
	}

	// Modal styles with enhanced visual appeal
	modalTitleStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 2).
		Bold(true).
		Align(lipgloss.Center)

	modalBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(2, 3).
		Background(lipgloss.Color("235")).
		Width(50)

	modalHelpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Italic(true)

	overlayStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("0")).
		Padding(0, 0)

	// Build modal content
	modalTitle := modalTitleStyle.Render("📋 Switch Profile")
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", 44))
	modalHelp := modalHelpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Cancel")

	modalContent := fmt.Sprintf("%s\n\n%s\n\n%s\n%s",
		modalTitle,
		profileList.String(),
		separator,
		modalHelp)

	modal := modalBoxStyle.Render(modalContent)

	// Simple overlay - place modal in center
	centeredModal := lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)

	// Layer modal over base view
	return overlayStyle.Render(centeredModal)
}

// getActiveProfileName returns the name of the active profile
func (v *SettingsView) getActiveProfileName() string {
	profileMgr := v.configMgr.GetProfileManager()
	if profileMgr == nil {
		return "default"
	}

	active := profileMgr.GetActive()
	if active == nil {
		return "default"
	}

	return active.Name
}

// getProfiles returns all available profiles
func (v *SettingsView) getProfiles() []string {
	profileMgr := v.configMgr.GetProfileManager()
	if profileMgr == nil {
		return []string{"default"}
	}

	profiles := profileMgr.List()
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}

	return names
}

// switchProfile switches to a different profile and reloads the configuration
func (v *SettingsView) switchProfile(name string) tea.Cmd {
	profileMgr := v.configMgr.GetProfileManager()
	if profileMgr == nil {
		return func() tea.Msg {
			return msgs.ProfileSwitchErrorMsg{Err: fmt.Errorf("profile manager not available")}
		}
	}

	// Switch the active profile
	if err := profileMgr.SetActive(name); err != nil {
		return func() tea.Msg {
			return msgs.ProfileSwitchErrorMsg{Err: err}
		}
	}

	// Reload configuration from the new profile
	v.loadConfig()
	v.modified = false

	return func() tea.Msg {
		return msgs.ProfileSwitchedMsg{ProfileName: name}
	}
}

// saveConfigCmd saves the configuration and returns a command
func (v *SettingsView) saveConfigCmd() tea.Cmd {
	if err := v.saveConfig(); err != nil {
		return func() tea.Msg {
			return msgs.ConfigSaveErrorMsg{Err: err}
		}
	}

	return func() tea.Msg {
		return msgs.ConfigSavedMsg{}
	}
}
