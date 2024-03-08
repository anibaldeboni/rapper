package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	TitleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Bold(true)
	TitleBarStyle     = lipgloss.NewStyle().PaddingBottom(1)
	ItemStyle         = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("255"))
	SelectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d6acff"))
	PaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(0)
	QuitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 1, 2).Render
	AppStyle          = lipgloss.NewStyle().Margin(1, 1, 1, 2).Render
	ProgressStyle     = lipgloss.NewStyle().Padding(0, 2, 1, 3).Render
	HelpStyle         = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("245"))
	ViewPortStyle     = lipgloss.NewStyle().Padding(1, 0, 2, 0).Render

	Bold   = lipgloss.NewStyle().Bold(true).Render
	Italic = lipgloss.NewStyle().Italic(true).Render
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render
	Pink   = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render
	Purple = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render

	IconError        = "❌"
	IconWomanDancing = "💃"
	IconTrophy       = "🏆"
	IconInformation  = "ℹ️"
	IconWarning      = "⚠️"
	IconSkull        = "💀"
)
