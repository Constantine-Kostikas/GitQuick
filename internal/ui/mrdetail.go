package ui

import (
	"fmt"
	"strings"

	"gitHelper/internal/platform"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MRDetailLoadedMsg is sent when MR detail data is loaded
type MRDetailLoadedMsg struct {
	Detail platform.MRDetail
	Err    error
}

// MRDetailModal displays detailed information about an MR/PR
type MRDetailModal struct {
	mr            platform.MR
	detail        platform.MRDetail
	platformName  string // "github" or "gitlab"
	loading       bool
	err           error
	spinner       spinner.Model
	cursor        int  // cursor position in file list
	wantsCheckout bool // signals dashboard to start checkout
	width         int
	height        int
}

// NewMRDetailModal creates a new MR detail modal
func NewMRDetailModal(mr platform.MR, platformName string, width, height int) MRDetailModal {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return MRDetailModal{
		mr:           mr,
		platformName: platformName,
		loading:      true,
		spinner:      s,
		cursor:       0,
		width:        width,
		height:       height,
	}
}

// Init returns the initial command (spinner tick)
func (m MRDetailModal) Init() tea.Cmd {
	return m.spinner.Tick
}

// SetDetail sets the detail data after loading
func (m *MRDetailModal) SetDetail(detail platform.MRDetail, err error) {
	m.loading = false
	if err != nil {
		m.err = err
	} else {
		m.detail = detail
	}
}

// Update handles messages
func (m MRDetailModal) Update(msg tea.Msg) (MRDetailModal, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case MRDetailLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.detail = msg.Detail
		}
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			maxCursor := len(m.detail.Files) - 1
			if maxCursor < 0 {
				maxCursor = 0
			}
			if m.cursor < maxCursor {
				m.cursor++
			}
		case "enter":
			m.wantsCheckout = true
		}
	}

	return m, nil
}

// View renders the modal
func (m MRDetailModal) View() string {
	// Calculate modal width - use available width with some margin
	modalWidth := m.width - 10
	if modalWidth < 50 {
		modalWidth = 50
	}
	if modalWidth > 80 {
		modalWidth = 80
	}

	contentWidth := modalWidth - 6 // Account for padding and borders

	var sections []string

	// Header section: title and branch
	titleLine := fmt.Sprintf("#%d %s", m.mr.Number, truncateString(m.mr.Title, contentWidth-8))
	branchLine := fmt.Sprintf("Branch: %s", m.mr.Branch)
	headerSection := titleLine + "\n" + branchLine
	sections = append(sections, headerSection)

	// Loading state
	if m.loading {
		loadingSection := m.spinner.View() + " Loading details..."
		sections = append(sections, loadingSection)

		footerSection := "[esc] close"
		sections = append(sections, footerSection)

		content := strings.Join(sections, "\n"+strings.Repeat("-", contentWidth)+"\n")
		return ModalStyle.Width(modalWidth).Render(content)
	}

	// Error state
	if m.err != nil {
		errorSection := ErrorStyle.Render("Error: " + m.err.Error())
		sections = append(sections, errorSection)

		footerSection := "[esc] close"
		sections = append(sections, footerSection)

		content := strings.Join(sections, "\n"+strings.Repeat("-", contentWidth)+"\n")
		return ModalStyle.Width(modalWidth).Render(content)
	}

	// Body section (truncated to ~5 lines)
	if m.detail.Body != "" {
		bodyLines := truncateBody(m.detail.Body, 5, contentWidth)
		sections = append(sections, bodyLines)
	}

	// File list section with scrolling
	if len(m.detail.Files) > 0 {
		fileSection := m.renderFileList(contentWidth)
		sections = append(sections, fileSection)
	}

	// Summary section
	var summarySection string
	if m.platformName == "gitlab" {
		summarySection = fmt.Sprintf("%d files changed", len(m.detail.Files))
	} else {
		summarySection = fmt.Sprintf("%d files changed  %s %s",
			len(m.detail.Files),
			SuccessStyle.Render(fmt.Sprintf("+%d", m.detail.Additions)),
			ErrorStyle.Render(fmt.Sprintf("-%d", m.detail.Deletions)))
	}
	sections = append(sections, summarySection)

	// Footer section with keybinds
	footerSection := DimStyle.Render("[j/k] scroll | [enter] checkout | [esc] close")
	sections = append(sections, footerSection)

	// Join sections with dividers
	content := strings.Join(sections, "\n"+strings.Repeat("-", contentWidth)+"\n")

	return ModalStyle.Width(modalWidth).Render(content)
}

// renderFileList renders the scrollable file list
func (m MRDetailModal) renderFileList(contentWidth int) string {
	if len(m.detail.Files) == 0 {
		return "No files changed"
	}

	// Calculate visible window (show ~5 files at a time)
	visibleCount := 5
	if len(m.detail.Files) < visibleCount {
		visibleCount = len(m.detail.Files)
	}

	// Calculate scroll offset to keep cursor visible
	startIdx := 0
	if m.cursor >= visibleCount {
		startIdx = m.cursor - visibleCount + 1
	}
	endIdx := startIdx + visibleCount
	if endIdx > len(m.detail.Files) {
		endIdx = len(m.detail.Files)
		startIdx = endIdx - visibleCount
		if startIdx < 0 {
			startIdx = 0
		}
	}

	var lines []string
	for i := startIdx; i < endIdx; i++ {
		file := m.detail.Files[i]

		var line string
		if m.platformName == "gitlab" {
			// GitLab: just show the file path (no line counts available)
			maxPathLen := contentWidth - 8
			if maxPathLen < 20 {
				maxPathLen = 20
			}
			path := truncateString(file.Path, maxPathLen)
			line = path
		} else {
			// GitHub: Format: +12  -3   src/auth/login.go
			addStr := SuccessStyle.Render(fmt.Sprintf("+%-4d", file.Additions))
			delStr := ErrorStyle.Render(fmt.Sprintf("-%-4d", file.Deletions))

			// Truncate path if needed
			maxPathLen := contentWidth - 20
			if maxPathLen < 20 {
				maxPathLen = 20
			}
			path := truncateString(file.Path, maxPathLen)

			line = fmt.Sprintf("%s %s %s", addStr, delStr, path)
		}

		if i == m.cursor {
			// Highlight the selected line
			indicator := " <- "
			line = SelectedItemStyle.Render(line) + DimStyle.Render(indicator)
		} else {
			line = NormalItemStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// Add scroll indicators if needed
	var result strings.Builder
	if startIdx > 0 {
		result.WriteString(DimStyle.Render("  ... more above ...") + "\n")
	}
	result.WriteString(strings.Join(lines, "\n"))
	if endIdx < len(m.detail.Files) {
		result.WriteString("\n" + DimStyle.Render("  ... more below ..."))
	}

	return result.String()
}

// IsLoading returns true if the modal is still loading data
func (m MRDetailModal) IsLoading() bool {
	return m.loading
}

// WantsCheckout returns true if user pressed enter to checkout
func (m MRDetailModal) WantsCheckout() bool {
	return m.wantsCheckout
}

// GetMR returns the MR associated with this modal
func (m MRDetailModal) GetMR() platform.MR {
	return m.mr
}

// truncateString truncates a string to maxLen, adding ellipsis if needed
func truncateString(s string, maxLen int) string {
	if maxLen <= 3 {
		maxLen = 4
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// truncateBody truncates body text to a maximum number of lines
func truncateBody(body string, maxLines, maxWidth int) string {
	// Replace carriage returns and normalize line endings
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")

	lines := strings.Split(body, "\n")

	var result []string
	lineCount := 0

	for _, line := range lines {
		if lineCount >= maxLines {
			break
		}

		// Wrap long lines
		wrapped := wrapLine(line, maxWidth)
		for _, wl := range wrapped {
			if lineCount >= maxLines {
				break
			}
			result = append(result, wl)
			lineCount++
		}
	}

	// Add ellipsis if we truncated
	totalLines := 0
	for _, line := range lines {
		totalLines += len(wrapLine(line, maxWidth))
	}

	if totalLines > maxLines && len(result) > 0 {
		// Indicate truncation with ellipsis
		lastLine := result[len(result)-1]
		if len(lastLine) > maxWidth-3 {
			result[len(result)-1] = lastLine[:maxWidth-3] + "..."
		} else {
			result[len(result)-1] = lastLine + "..."
		}
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Render(strings.Join(result, "\n"))
}

// wrapLine wraps a single line to fit within maxWidth
func wrapLine(line string, maxWidth int) []string {
	if maxWidth <= 0 {
		maxWidth = 80
	}
	if len(line) == 0 {
		return []string{""}
	}
	if len(line) <= maxWidth {
		return []string{line}
	}

	var result []string
	for len(line) > maxWidth {
		// Try to break at a space
		breakPoint := maxWidth
		for i := maxWidth; i > maxWidth/2; i-- {
			if line[i] == ' ' {
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
	return result
}
