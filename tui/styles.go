package tui

import "github.com/charmbracelet/lipgloss"

// Styles for the TUI
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	PlayingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("46"))

	StoppedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	// Timeline styles
	TimelineStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	// Event log style
	EventLogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			Foreground(lipgloss.Color("252"))

	// Keybinding styles
	KeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			PaddingLeft(1).
			PaddingRight(2)

	DescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			PaddingLeft(1).
			PaddingRight(1)

	// Highlight style for selected synth
	HighlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))
)
