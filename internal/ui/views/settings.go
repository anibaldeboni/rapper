package views

import (
	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// SettingsView displays and edits configuration settings
type SettingsView struct {
	configMgr config.Manager
	width     int
	height    int
}

// NewSettingsView creates a new SettingsView
func NewSettingsView(configMgr config.Manager) *SettingsView {
	return &SettingsView{
		configMgr: configMgr,
	}
}

// Update handles messages for the settings view
func (v *SettingsView) Update(msg tea.Msg) tea.Cmd {
	// TODO: Implement in Phase 3
	return nil
}

// Resize updates the view dimensions
func (v *SettingsView) Resize(width, height int) {
	v.width = width
	v.height = height
}

// View renders the settings view
func (v *SettingsView) View() string {
	// TODO: Implement in Phase 3
	// For now, show a placeholder
	title := ui.TitleStyle.Render("⚙️  Settings")
	content := "\n\nSettings view will be implemented in Phase 3.\n\n"
	content += "Features:\n"
	content += "  • Edit URL template, body template, headers\n"
	content += "  • Switch between profiles (Ctrl+P)\n"
	content += "  • Save configuration (Ctrl+S)\n\n"
	content += "Press Esc to go back."

	return ui.AppStyle(title + content)
}
