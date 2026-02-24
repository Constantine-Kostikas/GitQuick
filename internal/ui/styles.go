package ui

import "charm.land/lipgloss/v2"

var (
	// Colors - night-friendly terminal palette
	primaryColor   = lipgloss.Color("36")  // Dark cyan
	secondaryColor = lipgloss.Color("242") // Dim gray
	successColor   = lipgloss.Color("35")  // Muted green
	errorColor     = lipgloss.Color("167") // Soft red
	selectedColor  = lipgloss.Color("35")  // Muted green
	accentColor    = lipgloss.Color("72")  // Teal

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

	SelectedRowStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(accentColor).
				PaddingLeft(1)

	NormalRowStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))

	// Selection indicator
	SelectedIndicator = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	// Dimmed style for descriptions
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	// Branch style - grey for branch names on second line
	BranchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	// Status styles
	StatusOpenStyle = lipgloss.NewStyle().
			Foreground(successColor)

	StatusDraftStyle = lipgloss.NewStyle().
				Foreground(secondaryColor)

	StatusMergedStyle = lipgloss.NewStyle().
				Foreground(accentColor) // Teal

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
			BorderForeground(accentColor).
			Padding(0, 2)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor)
)
