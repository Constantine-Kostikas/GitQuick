package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// DescriptionViewer displays a scrollable full description
type DescriptionViewer struct {
	title    string
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// NewDescriptionViewer creates a new description viewer
func NewDescriptionViewer(title, content string, width, height int) DescriptionViewer {
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

	// Wrap the content to fit the width
	wrappedContent := wrapText(content, contentWidth)

	vp := viewport.New(viewport.WithWidth(contentWidth), viewport.WithHeight(modalHeight-4)) // Leave room for title and footer
	vp.SetContent(wrappedContent)
	vp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	return DescriptionViewer{
		title:    title,
		viewport: vp,
		width:    modalWidth,
		height:   modalHeight,
		ready:    true,
	}
}

// Init returns the initial command
func (d DescriptionViewer) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (d DescriptionViewer) Update(msg tea.Msg) (DescriptionViewer, tea.Cmd) {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

// View renders the viewer
func (d DescriptionViewer) View() string {
	titleLine := lipgloss.NewStyle().
		Bold(true).
		Foreground(accentColor).
		Render(d.title)

	// Scroll indicator
	scrollPercent := d.viewport.ScrollPercent() * 100
	scrollInfo := DimStyle.Render(
		lipgloss.NewStyle().
			Render(strings.Repeat("─", d.width-6)),
	)
	if d.viewport.TotalLineCount() > d.viewport.Height() {
		scrollInfo = DimStyle.Render(
			lipgloss.NewStyle().
				Render(formatScrollPercent(scrollPercent, d.width-6)),
		)
	}

	footer := DimStyle.Render("[j/k] scroll | [esc] close")

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleLine,
		scrollInfo,
		d.viewport.View(),
		footer,
	)

	return ModalStyle.Width(d.width).Render(content)
}

// formatScrollPercent creates a scroll indicator line
func formatScrollPercent(percent float64, width int) string {
	if width < 20 {
		width = 20
	}
	percentStr := lipgloss.NewStyle().
		Foreground(accentColor).
		Render(fmt.Sprintf(" %.0f%% ", percent))
	lineWidth := (width - 6) / 2
	return strings.Repeat("─", lineWidth) + percentStr + strings.Repeat("─", lineWidth)
}

// wrapText wraps text to fit within maxWidth
func wrapText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 80
	}

	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if len(line) == 0 {
			result = append(result, "")
			continue
		}

		// Wrap long lines
		for len(line) > maxWidth {
			// Try to break at a space
			breakPoint := maxWidth
			for i := maxWidth; i > maxWidth/2; i-- {
				if i < len(line) && line[i] == ' ' {
					breakPoint = i
					break
				}
			}
			result = append(result, line[:breakPoint])
			line = strings.TrimLeft(line[breakPoint:], " ")
		}
		if len(line) > 0 {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
