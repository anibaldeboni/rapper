package ui

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"golang.org/x/term"
)

var (
	bullet      = "⦿"
	inactiveDot = "⦁"
	TitleStyle  = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Bold(true)
	TitleBarStyle       = lipgloss.NewStyle().PaddingBottom(1)
	ItemStyle           = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("255"))
	SelectedItemStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#d6acff"))
	PaginationStyle     = lipgloss.NewStyle().PaddingLeft(2)
	ActivePaginationDot = lipgloss.NewStyle().
				Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#d3d3d3"), Dark: lipgloss.Color("#d3d3d3")}).
				SetString(bullet).
				Bold(true)
	InactivePaginationDot = lipgloss.NewStyle().
				Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#8d8d8d"), Dark: lipgloss.Color("#8d8d8d")}).
				SetString(inactiveDot).
				Bold(true)
	ProgressStyle = lipgloss.NewStyle().Padding(0, 2, 1, 3).Render
	HelpStyle     = lipgloss.NewStyle().PaddingLeft(1).Render //.Foreground(lipgloss.Color("245"))
	ViewPortStyle = lipgloss.NewStyle().PaddingTop(1).Render
	LogoStyle     = lipgloss.NewStyle().
			Background(lipgloss.Color("#F25D94")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Padding(0, 1).
			Render
	HelpKeyStyle = lipgloss.NewStyle().Foreground(compat.AdaptiveColor{
		Light: lipgloss.Color("#d3d3d3"),
		Dark:  lipgloss.Color("#d3d3d3"),
	}).
		Bold(true)

	HelpDescStyle = lipgloss.NewStyle().Foreground(compat.AdaptiveColor{
		Light: lipgloss.Color("#8d8d8d"),
		Dark:  lipgloss.Color("#8d8d8d"),
	})

	HelpSepStyle = lipgloss.NewStyle().Foreground(compat.AdaptiveColor{
		Light: lipgloss.Color("#DDDADA"),
		Dark:  lipgloss.Color("#535353"),
	})

	ScreenCenteredStyle = lipgloss.NewStyle().
				Width(TerminalWidth()).
				Align(lipgloss.Center).Render

	Bold   = lipgloss.NewStyle().Bold(true).Render
	Italic = lipgloss.NewStyle().Italic(true).Render
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render
	Pink   = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render
	Purple = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render

	IconError        = "❌"
	IconWomanDancing = "💃"
	IconTrophy       = "🏆"
	IconInformation  = "ℹ️ "
	IconWarning      = "⚠️"
	IconSkull        = "💀"
)

func TerminalWidth() int {
	width, _, err := term.GetSize(0)
	if err != nil {
		width = 80
	}
	return width
}
