package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("39")  // Blue
	secondaryColor = lipgloss.Color("245") // Gray
	successColor   = lipgloss.Color("42")  // Green
	errorColor     = lipgloss.Color("196") // Red
	selectedColor  = lipgloss.Color("170") // Purple

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Padding(0, 2)

	// List item styles
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(selectedColor).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Status styles
	StatusOpenStyle = lipgloss.NewStyle().
			Foreground(successColor)

	StatusDraftStyle = lipgloss.NewStyle().
				Foreground(secondaryColor)

	StatusMergedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("141")) // Purple

	StatusClosedStyle = lipgloss.NewStyle().
				Foreground(errorColor)

	// Footer style
	FooterStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	// Modal style
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor)
)
