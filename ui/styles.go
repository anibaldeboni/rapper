package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	SpinnerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("47"))
	DotStyle         = HelpStyle.Copy().UnsetMargins()
	AppStyle         = lipgloss.NewStyle().Margin(1, 2, 0, 2)
	SpinnerHelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(0).PaddingBottom(1).Foreground(lipgloss.Color("241"))

	TitleStyle        = lipgloss.NewStyle().MarginLeft(2)
	ItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("47"))
	PaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	HelpStyle         = list.DefaultStyles().HelpStyle.PaddingBottom(1).Foreground(lipgloss.Color("241"))
	QuitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 1, 2)
	IconFireCracker   = "🧨"
	IconTrophy        = "🏆"
	IconInformation   = "ℹ️"
	IconWarning       = "⚠️"
)
