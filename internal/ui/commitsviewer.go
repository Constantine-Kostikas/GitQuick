package ui

import (
	"fmt"
	"strings"

	"gitHelper/internal/platform"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommitsViewer displays a scrollable list of commits
type CommitsViewer struct {
	title    string
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// NewCommitsViewer creates a new commits viewer
func NewCommitsViewer(title string, commits []platform.Commit, width, height int) CommitsViewer {
	// Calculate modal dimensions
	modalWidth := width - 10
	if modalWidth < 50 {
		modalWidth = 50
	}
	if modalWidth > 100 {
		modalWidth = 100
	}

	modalHeight := height - 10
	if modalHeight < 10 {
		modalHeight = 10
	}

	contentWidth := modalWidth - 6 // Account for padding and borders

	// Format commits for display
	content := formatCommits(commits, contentWidth)

	vp := viewport.New(contentWidth, modalHeight-4) // Leave room for title and footer
	vp.SetContent(content)
	vp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	return CommitsViewer{
		title:    title,
		viewport: vp,
		width:    modalWidth,
		height:   modalHeight,
		ready:    true,
	}
}

// formatCommits formats the commits for display
func formatCommits(commits []platform.Commit, width int) string {
	if len(commits) == 0 {
		return "No commits found"
	}

	var lines []string
	for _, c := range commits {
		// SHA in accent color
		shaStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
		sha := shaStyle.Render(c.SHA)

		// Author in dim
		authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
		author := authorStyle.Render(c.Author)

		// Calculate max message length: width - sha(7) - spaces(4) - author
		maxMsgLen := width - 12 - len(c.Author)
		if maxMsgLen < 20 {
			maxMsgLen = 20
		}
		msg := c.Message
		if len(msg) > maxMsgLen {
			msg = msg[:maxMsgLen-3] + "..."
		}
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		message := msgStyle.Render(msg)

		// Single line: SHA message (author)
		line := fmt.Sprintf("%s %s %s", sha, message, author)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// Init returns the initial command
func (c CommitsViewer) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (c CommitsViewer) Update(msg tea.Msg) (CommitsViewer, tea.Cmd) {
	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)
	return c, cmd
}

// View renders the viewer
func (c CommitsViewer) View() string {
	titleLine := lipgloss.NewStyle().
		Bold(true).
		Foreground(accentColor).
		Render(c.title)

	// Scroll indicator
	scrollInfo := DimStyle.Render(strings.Repeat("─", c.width-6))
	if c.viewport.TotalLineCount() > c.viewport.Height {
		scrollPercent := c.viewport.ScrollPercent() * 100
		scrollInfo = DimStyle.Render(formatCommitsScrollPercent(scrollPercent, c.width-6))
	}

	footer := DimStyle.Render("[j/k] scroll | [esc] close")

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleLine,
		scrollInfo,
		c.viewport.View(),
		footer,
	)

	return ModalStyle.Width(c.width).Render(content)
}

// formatCommitsScrollPercent creates a scroll indicator line
func formatCommitsScrollPercent(percent float64, width int) string {
	if width < 20 {
		width = 20
	}
	percentStr := lipgloss.NewStyle().
		Foreground(accentColor).
		Render(fmt.Sprintf(" %.0f%% ", percent))
	lineWidth := (width - 6) / 2
	return strings.Repeat("─", lineWidth) + percentStr + strings.Repeat("─", lineWidth)
}
