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
	bodyField
	headersField
	maxFields
)

// SettingsView displays and edits configuration settings
type SettingsView struct {
	configMgr config.Manager
	width     int
	height    int

	// Form fields
	urlInput     textinput.Model
	bodyInput    textarea.Model
	headersInput textarea.Model

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
	urlInput.Width = 80
	urlInput.Prompt = ""

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

	v := &SettingsView{
		configMgr:    configMgr,
		urlInput:     urlInput,
		bodyInput:    bodyInput,
		headersInput: headersInput,
		focused:      urlField,
		focusable:    []int{urlField, bodyField, headersField},
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
	v.bodyInput.SetValue(cfg.Request.BodyTemplate)

	// Convert headers map to string
	var headerLines []string
	for key, value := range cfg.Request.Headers {
		headerLines = append(headerLines, fmt.Sprintf("%s: %s", key, value))
	}
	v.headersInput.SetValue(strings.Join(headerLines, "\n"))
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

// saveConfig saves the current form values to configuration
func (v *SettingsView) saveConfig() error {
	cfg := v.configMgr.Get()
	if cfg == nil {
		cfg = &config.Config{}
	}

	// Update config from form fields
	cfg.Request.URLTemplate = v.urlInput.Value()
	cfg.Request.BodyTemplate = v.bodyInput.Value()
	cfg.Request.Headers = parseHeaders(v.headersInput.Value())

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
	v.bodyInput.Blur()
	v.headersInput.Blur()

	switch v.focused {
	case urlField:
		v.urlInput.Focus()
	case bodyField:
		v.bodyInput.Focus()
	case headersField:
		v.headersInput.Focus()
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
					if err := v.switchProfile(profiles[v.profileListIndex]); err != nil {
						// TODO: Show error toast
					}
				}
				v.showProfileSelector = false
				return nil

			case "esc":
				v.showProfileSelector = false
				return nil
			}
			return nil
		}

		// Handle keyboard shortcuts
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
			// Save configuration
			if err := v.saveConfig(); err != nil {
				// TODO: Show error toast
				return nil
			}
			// TODO: Show success toast
			return nil

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
	v.bodyInput.SetWidth(inputWidth)
	v.headersInput.SetWidth(inputWidth)
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

	// Help text
	help := "Tab/Shift+Tab: Switch fields • Ctrl+S: Save • Ctrl+P: Switch profile • Esc: Back"
	if v.modified {
		help += " • ⚠️  Unsaved changes"
	}
	b.WriteString(helpStyle.Render(help))

	baseView := settingsAppStyle.Render(b.String())

	// Show profile selector modal if active
	if v.showProfileSelector {
		return v.renderWithProfileSelector(baseView)
	}

	return baseView
}

// renderWithProfileSelector renders the profile selector modal overlay
func (v *SettingsView) renderWithProfileSelector(baseView string) string {
	profiles := v.getProfiles()

	// Build profile list
	var profileList strings.Builder
	for i, name := range profiles {
		if i == v.profileListIndex {
			profileList.WriteString(selectedItemStyle.Render(fmt.Sprintf("▶ %s", name)))
		} else {
			profileList.WriteString(itemStyle.Render(fmt.Sprintf("  %s", name)))
		}
		profileList.WriteString("\n")
	}

	// Modal styles
	modalTitleStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1).
		Bold(true)

	modalBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Background(lipgloss.Color("235"))

	overlayStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("0")).
		Padding(0, 0)

	// Build modal content
	modalTitle := modalTitleStyle.Render("📋 Select Profile")
	modalHelp := helpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Cancel")
	modalContent := fmt.Sprintf("%s\n\n%s\n%s", modalTitle, profileList.String(), modalHelp)
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
func (v *SettingsView) switchProfile(name string) error {
	profileMgr := v.configMgr.GetProfileManager()
	if profileMgr == nil {
		return fmt.Errorf("profile manager not available")
	}

	// Switch the active profile
	if err := profileMgr.SetActive(name); err != nil {
		return err
	}

	// Reload configuration from the new profile
	v.loadConfig()
	v.modified = false

	return nil
}
