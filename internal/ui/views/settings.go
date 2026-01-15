package views

import (
	"fmt"
	"strings"

	"github.com/anibaldeboni/rapper/internal/config"
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

	// Method selector
	showMethodSelector  bool
	methodSelectorIndex int

	// State
	modified bool
}

// NewSettingsView creates a new SettingsView
func NewSettingsView(configMgr config.Manager) *SettingsView {
	// Create URL input
	urlInput := textinput.New()
	urlInput.Placeholder = "http://localhost:8080/api/v1/users"
	urlInput.CharLimit = 500
	urlInput.Width = 80
	urlInput.Prompt = ""

	// Create method input
	methodInput := textinput.New()
	methodInput.Placeholder = "POST"
	methodInput.CharLimit = 10
	methodInput.Width = 20
	methodInput.Prompt = ""

	// Create body textarea
	bodyInput := textarea.New()
	bodyInput.Placeholder = `{"name": "{{.name}}", "email": "{{.email}}"}`
	bodyInput.CharLimit = 5000
	bodyInput.SetHeight(5)
	bodyInput.SetWidth(80)
	bodyInput.ShowLineNumbers = false

	// Create headers textarea
	headersInput := textarea.New()
	headersInput.Placeholder = `Content-Type: application/json
Authorization: Bearer {{.token}}`
	headersInput.CharLimit = 2000
	headersInput.SetHeight(3)
	headersInput.SetWidth(80)
	headersInput.ShowLineNumbers = false

	// Create CSV fields textarea
	csvFieldsInput := textarea.New()
	csvFieldsInput.Placeholder = `id
name
email`
	csvFieldsInput.CharLimit = 1000
	csvFieldsInput.SetHeight(4)
	csvFieldsInput.SetWidth(80)
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
		// If profile selector is open, handle its navigation
		if v.showProfileSelector {
			profiles := v.getProfiles()
			switch msg.String() {
			case "up", "k":
				v.profileListIndex--
				if v.profileListIndex < 0 {
					v.profileListIndex = len(profiles) - 1
				}
				return nil

			case "down", "j":
				v.profileListIndex++
				if v.profileListIndex >= len(profiles) {
					v.profileListIndex = 0
				}
				return nil

			case "enter":
				// Switch to selected profile
				if v.profileListIndex >= 0 && v.profileListIndex < len(profiles) {
					v.showProfileSelector = false
					return v.switchProfile(profiles[v.profileListIndex])
				}
				v.showProfileSelector = false
				return nil

			case "esc", tea.KeyEscape.String():
				v.showProfileSelector = false
				return nil
			}
			return nil
		}

		// If method selector is open, handle its navigation
		if v.showMethodSelector {
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
			switch msg.String() {
			case "up", "k":
				v.methodSelectorIndex--
				if v.methodSelectorIndex < 0 {
					v.methodSelectorIndex = len(methods) - 1
				}
				return nil

			case "down", "j":
				v.methodSelectorIndex++
				if v.methodSelectorIndex >= len(methods) {
					v.methodSelectorIndex = 0
				}
				return nil

			case "enter":
				// Select method
				if v.methodSelectorIndex >= 0 && v.methodSelectorIndex < len(methods) {
					v.methodInput.SetValue(methods[v.methodSelectorIndex])
					v.modified = true
				}
				v.showMethodSelector = false
				return nil

			case "esc", tea.KeyEscape.String():
				v.showMethodSelector = false
				return nil
			}
			return nil
		}

		// Handle keyboard shortcuts
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
			// Save configuration
			return v.saveConfigCmd()

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+p"))):
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

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+m"))):
			// Toggle method selector
			v.showMethodSelector = !v.showMethodSelector
			// Set initial selection to current method
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
			currentMethod := v.methodInput.Value()
			for i, method := range methods {
				if method == currentMethod {
					v.methodSelectorIndex = i
					break
				}
			}
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			v.nextField()
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
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

	// Adjust input widths
	inputWidth := width - 8 // Account for margins and labels
	if inputWidth > 120 {
		inputWidth = 120
	}
	if inputWidth < 40 {
		inputWidth = 40
	}

	v.urlInput.Width = inputWidth
	v.methodInput.Width = min(inputWidth, 30)
	v.bodyInput.SetWidth(inputWidth)
	v.headersInput.SetWidth(inputWidth)
	v.csvFieldsInput.SetWidth(inputWidth)

	// Adjust textarea heights based on available height
	// Reserve space for: header(3) + labels(10) + help(3) + margins(4) = ~20 lines
	availableForTextareas := max(height-20, 10)

	// Distribute height: body gets 40%, headers gets 30%, csvFields gets 30%
	bodyHeight := max((availableForTextareas*4)/10, 3)
	if bodyHeight > 8 {
		bodyHeight = 8
	}

	headersHeight := max((availableForTextareas*3)/10, 2)
	if headersHeight > 6 {
		headersHeight = 6
	}

	csvFieldsHeight := max(availableForTextareas-bodyHeight-headersHeight, 3)
	if csvFieldsHeight > 6 {
		csvFieldsHeight = 6
	}

	v.bodyInput.SetHeight(bodyHeight)
	v.headersInput.SetHeight(headersHeight)
	v.csvFieldsInput.SetHeight(csvFieldsHeight)
}

// View renders the settings view
func (v *SettingsView) View() string {
	var b strings.Builder

	// Header with profile badge
	profile := v.getActiveProfileName()
	title := settingsTitleStyle.Render("⚙️  Settings")
	profileBadge := profileBadgeStyle.Render(fmt.Sprintf("📋 %s", profile))
	header := lipgloss.JoinHorizontal(lipgloss.Left, title, profileBadge)
	b.WriteString(header)
	b.WriteString("\n\n")

	// URL field
	urlLabel := labelStyle.Render("URL Template:")
	if v.focused == urlField {
		urlLabel = focusedStyle.Render("▶ URL Template:")
	}
	b.WriteString(urlLabel)
	b.WriteString("\n")
	b.WriteString(v.urlInput.View())
	b.WriteString("\n\n")

	// Method field
	methodLabel := labelStyle.Render("Method:")
	if v.focused == methodField {
		methodLabel = focusedStyle.Render("▶ Method:")
	}
	b.WriteString(methodLabel)
	b.WriteString("\n")
	b.WriteString(v.methodInput.View())
	b.WriteString("\n\n")

	// Body field
	bodyLabel := labelStyle.Render("Body Template:")
	if v.focused == bodyField {
		bodyLabel = focusedStyle.Render("▶ Body Template:")
	}
	b.WriteString(bodyLabel)
	b.WriteString("\n")
	b.WriteString(v.bodyInput.View())
	b.WriteString("\n\n")

	// Headers field
	headersLabel := labelStyle.Render("Headers:")
	if v.focused == headersField {
		headersLabel = focusedStyle.Render("▶ Headers:")
	}
	b.WriteString(headersLabel)
	b.WriteString("\n")
	b.WriteString(v.headersInput.View())
	b.WriteString("\n\n")

	// CSV Fields field
	csvFieldsLabel := labelStyle.Render("CSV Fields (one per line):")
	if v.focused == csvFieldsField {
		csvFieldsLabel = focusedStyle.Render("▶ CSV Fields (one per line):")
	}
	b.WriteString(csvFieldsLabel)
	b.WriteString("\n")
	b.WriteString(v.csvFieldsInput.View())
	b.WriteString("\n\n")

	// Help text
	var help string
	if v.modified {
		help = "⚠️  Unsaved changes"
	}
	b.WriteString(helpStyle.Render(help))

	baseView := settingsAppStyle.Render(b.String())

	// Show method selector modal if active
	if v.showMethodSelector {
		return v.renderWithMethodSelector(baseView)
	}

	// Show profile selector modal if active
	if v.showProfileSelector {
		return v.renderWithProfileSelector(baseView)
	}

	return baseView
}

// renderWithProfileSelector renders the profile selector modal overlay
func (v *SettingsView) renderWithProfileSelector(baseView string) string {
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

// renderWithMethodSelector renders the method selector modal overlay
func (v *SettingsView) renderWithMethodSelector(baseView string) string {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	currentMethod := v.methodInput.Value()

	// Build method list
	var methodList strings.Builder

	// Styles for method items
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	activeTagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("40")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	for i, method := range methods {
		var line string
		isActive := method == currentMethod
		isSelected := i == v.methodSelectorIndex

		// Add selection indicator
		if isSelected {
			line = "▶ "
		} else {
			line = "  "
		}

		// Add method name with style
		if isSelected {
			line += selectedStyle.Render(method)
		} else {
			line += normalStyle.Render(method)
		}

		// Add active badge
		if isActive {
			line += " " + activeTagStyle.Render("●")
		}

		methodList.WriteString(line)
		methodList.WriteString("\n")
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
		Width(40)

	modalHelpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Italic(true)

	overlayStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("0")).
		Padding(0, 0)

	// Build modal content
	modalTitle := modalTitleStyle.Render("🌐 Select HTTP Method")
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", 34))
	modalHelp := modalHelpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Cancel")

	modalContent := fmt.Sprintf("%s\n\n%s\n%s\n%s",
		modalTitle,
		methodList.String(),
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

// ConfigSavedMsg is sent when configuration is successfully saved
type ConfigSavedMsg struct{}

// ConfigSaveErrorMsg is sent when configuration save fails
type ConfigSaveErrorMsg struct {
	Err error
}

// ProfileSwitchedMsg is sent when profile is successfully switched
type ProfileSwitchedMsg struct {
	ProfileName string
}

// ProfileSwitchErrorMsg is sent when profile switch fails
type ProfileSwitchErrorMsg struct {
	Err error
}

// switchProfile switches to a different profile and reloads the configuration
func (v *SettingsView) switchProfile(name string) tea.Cmd {
	profileMgr := v.configMgr.GetProfileManager()
	if profileMgr == nil {
		return func() tea.Msg {
			return ProfileSwitchErrorMsg{Err: fmt.Errorf("profile manager not available")}
		}
	}

	// Switch the active profile
	if err := profileMgr.SetActive(name); err != nil {
		return func() tea.Msg {
			return ProfileSwitchErrorMsg{Err: err}
		}
	}

	// Reload configuration from the new profile
	v.loadConfig()
	v.modified = false

	return func() tea.Msg {
		return ProfileSwitchedMsg{ProfileName: name}
	}
}

// saveConfigCmd saves the configuration and returns a command
func (v *SettingsView) saveConfigCmd() tea.Cmd {
	if err := v.saveConfig(); err != nil {
		return func() tea.Msg {
			return ConfigSaveErrorMsg{Err: err}
		}
	}

	return func() tea.Msg {
		return ConfigSavedMsg{}
	}
}
