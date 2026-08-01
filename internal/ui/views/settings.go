package views

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui/kbind"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	"github.com/anibaldeboni/rapper/internal/utils"
)

var (
	settingsTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1).Bold(true)
	headerStyle        = lipgloss.NewStyle().MarginTop(1)
	profileBadgeStyle  = lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("230")).Padding(0, 1).MarginLeft(2)
	inlineErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).MarginTop(1)
)

const (
	paneList     = 0
	paneForm     = 1
	minListWidth = 15
	maxListWidth = 30
)

type SettingsView struct {
	configMgr ports.ConfigManager
	width     int
	height    int

	profileList list.Model

	editor textarea.Model

	editing         bool
	focusPane       int
	selectedProfile string
	activeProfile   string
	inlineErr       string
}

var _ tea.Model = SettingsView{}

func NewSettingsView(configMgr ports.ConfigManager) SettingsView {
	editor := textarea.New()
	editor.Placeholder = "# raw YAML — edit and press ctrl+s to save"
	editor.CharLimit = 0
	editor.ShowLineNumbers = true
	editor.SetHeight(5)

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

	activeIdx := indexOf(profileNames, configMgr.GetActiveProfile())
	if activeIdx >= 0 {
		profileList.Select(activeIdx)
	}

	v := SettingsView{
		configMgr:     configMgr,
		profileList:   profileList,
		editor:        editor,
		focusPane:     paneList,
		editing:       false,
		activeProfile: configMgr.GetActiveProfile(),
	}

	v.selectedProfile = v.configMgr.GetActiveProfile()
	if raw, err := v.configMgr.GetProfileRaw(v.selectedProfile); err == nil {
		v.editor.SetValue(raw)
	} else {

		v.inlineErr = fmt.Sprintf("could not read %s: %v", v.selectedProfile, err)
	}

	return v
}

func indexOf(haystack []string, needle string) int {
	for i, s := range haystack {
		if s == needle {
			return i
		}
	}
	return -1
}

type profileItemDelegate struct {
	active string
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
	name := opt.Value
	if opt.Value == d.active {
		name = lipgloss.NewStyle().Foreground(lipgloss.Color("#d6acff")).Bold(true).Render(name)
	}
	fmt.Fprintf(w, "%s%s", prefix, name)
}

func (v SettingsView) FocusedPane() int { return v.focusPane }

func (v SettingsView) KeyMap() help.KeyMap {
	return settingsViewKeyMap{editing: v.editing}
}

type settingsViewKeyMap struct {
	editing bool
}

var selectProfileBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "select profile"),
)

func (k settingsViewKeyMap) selectionShortHelp() []key.Binding {
	return []key.Binding{kbind.Up, kbind.Down, kbind.Edit, selectProfileBinding}
}

func (k settingsViewKeyMap) selectionFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{kbind.Up, kbind.Down, kbind.Edit, selectProfileBinding},
	}
}

func (k settingsViewKeyMap) editShortHelp() []key.Binding {
	return []key.Binding{kbind.ExitEdit, kbind.Save}
}

func (k settingsViewKeyMap) editFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{kbind.ExitEdit, kbind.Save},
	}
}

func (k settingsViewKeyMap) ShortHelp() []key.Binding {
	if k.editing {
		return k.editShortHelp()
	}
	return k.selectionShortHelp()
}

func (k settingsViewKeyMap) FullHelp() [][]key.Binding {
	if k.editing {
		return k.editFullHelp()
	}
	return k.selectionFullHelp()
}

func (v SettingsView) Init() tea.Cmd { return nil }

func (v SettingsView) loadSelectedProfileRaw(name string) SettingsView {
	v.selectedProfile = name
	raw, err := v.configMgr.GetProfileRaw(name)
	if err != nil {
		v.editor.SetValue("")
		v.inlineErr = fmt.Sprintf("could not read %s: %v", name, err)
		return v
	}
	v.editor.SetValue(raw)
	v.inlineErr = ""
	return v
}

func (v SettingsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case msgs.ViewActivatedMsg:
		return v, v.emitStatusKeymap()

	case msgs.ViewportSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

		return v.layoutEditor(), nil

	case msgs.ConfigSavedMsg:

		v.inlineErr = ""
		return v.layoutEditor(), nil

	case msgs.ConfigSaveErrorMsg:

		v.inlineErr = msg.Err.Error()
		return v.layoutEditor(), nil

	case tea.KeyPressMsg:

		switch {
		case key.Matches(msg, kbind.Edit) && !v.editing:

			v.editing = true
			v.focusPane = paneForm
			v.editor.Focus()

			return v.layoutEditor(), v.emitStatusKeymap()
		case key.Matches(msg, kbind.ExitEdit) && v.editing:
			v.editing = false
			v.focusPane = paneList
			v.editor.Blur()

			return v.layoutEditor(), v.emitStatusKeymap()
		}

		if v.editing {
			return v.handleEditKey(msg)
		}
		return v.handleSelectionKey(msg)
	}

	return v, nil
}

func (v SettingsView) emitStatusKeymap() tea.Cmd {
	return func() tea.Msg {
		return msgs.StatusKeymapMsg{Current: v.KeyMap()}
	}
}

func (v SettingsView) layoutEditor() SettingsView {
	listW := utils.Clamp(v.width/4, minListWidth, maxListWidth)
	formW := v.width - listW
	v.editor.SetWidth(formW - 4)
	v.editor.SetHeight(v.editorInnerHeight())
	v.profileList.SetSize(listW, v.height-5)
	return v
}

func (v SettingsView) editorInnerHeight() int {
	borderReserve := 0
	if v.focusPane == paneForm {
		borderReserve = 2
	}
	errorReserve := 0
	if v.inlineErr != "" {
		errorReserve = 1
	}
	h := v.height - 5 - borderReserve - errorReserve
	if h < 1 {
		return 1
	}
	return h
}

func (v SettingsView) handleSelectionKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	oldIdx := v.profileList.Index()
	var listCmd tea.Cmd
	v.profileList, listCmd = v.profileList.Update(msg)

	if v.profileList.Index() != oldIdx {
		if opt, ok := v.profileList.SelectedItem().(Option[string]); ok {
			v = v.loadSelectedProfileRaw(opt.Value)
		}
	}

	if key.Matches(msg, kbind.Select) {
		if opt, ok := v.profileList.SelectedItem().(Option[string]); ok {
			return v.activateProfile(opt.Value)
		}
		return v, listCmd
	}
	return v, listCmd
}

func (v SettingsView) handleEditKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {

	if key.Matches(msg, kbind.Save) {
		return v, v.saveRawCmd()
	}
	var cmd tea.Cmd
	v.editor, cmd = v.editor.Update(msg)
	return v, cmd
}

func (v SettingsView) saveRawCmd() tea.Cmd {
	name := v.selectedProfile
	content := v.editor.Value()
	err := v.configMgr.SaveProfileRaw(name, content)
	if err != nil {

		_ = err

		return func() tea.Msg {
			return msgs.ConfigSaveErrorMsg{Err: err}
		}
	}
	return func() tea.Msg {
		return msgs.ConfigSavedMsg{}
	}
}

func (v SettingsView) activateProfile(name string) (SettingsView, tea.Cmd) {
	if err := v.configMgr.SetActiveProfile(name); err != nil {
		return v, func() tea.Msg {
			return msgs.ProfileSwitchErrorMsg{Err: err}
		}
	}

	v.activeProfile = name

	v = v.loadSelectedProfileRaw(name)

	return v, func() tea.Msg {
		return msgs.ProfileSwitchedMsg{ProfileName: name}
	}
}

func (v SettingsView) View() tea.View {
	listW := utils.Clamp(v.width/4, minListWidth, maxListWidth)
	formW := v.width - listW

	listBg := v.paneBg(v.focusPane == paneList)
	listTitle := headerStyle.MarginLeft(1).Render(settingsTitleStyle.Render("📋 Profiles"))

	v.profileList.SetDelegate(profileItemDelegate{active: v.activeProfile})
	listPane := lipgloss.JoinVertical(
		lipgloss.Top,
		listTitle,
		listBg.
			Width(listW).
			Height(v.height-3).
			Render(v.profileList.View()),
	)

	formBg := v.paneBg(v.focusPane == paneForm)
	header := headerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			settingsTitleStyle.Render("⚙️ Settings"),
			profileBadgeStyle.Render("📋 "+v.getActiveProfileName()),
		),
	)
	editorBlock := lipgloss.JoinVertical(
		lipgloss.Top,
		v.editor.View(),
		v.renderInlineError(),
	)
	formPane := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		formBg.
			Width(formW).
			Height(v.height-3).
			Render(editorBlock),
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

func (v SettingsView) renderInlineError() string {
	if v.inlineErr == "" {
		return ""
	}
	return inlineErrorStyle.Render("⚠ " + v.inlineErr)
}

func (v SettingsView) getActiveProfileName() string {
	profileName := v.configMgr.GetActiveProfile()
	if profileName == "" {
		return "default"
	}
	return profileName
}
