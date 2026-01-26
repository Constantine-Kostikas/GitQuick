package ui

import (
	"fmt"
	"io"

	"gitHelper/internal/platform"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MRItem wraps an MR for the list component
type MRItem struct {
	MR platform.MR
}

func (i MRItem) Title() string {
	return fmt.Sprintf("#%d %s", i.MR.Number, i.MR.Title)
}

func (i MRItem) Description() string {
	return i.MR.Branch
}

func (i MRItem) FilterValue() string {
	return i.MR.Title
}

// CompactDelegate is a custom delegate for compact MR list rendering
type CompactDelegate struct{}

func (d CompactDelegate) Height() int                             { return 3 }
func (d CompactDelegate) Spacing() int                            { return 0 }
func (d CompactDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d CompactDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	mrItem, ok := item.(MRItem)
	if !ok {
		return
	}

	mr := mrItem.MR
	isSelected := index == m.Index()

	// Status indicator
	var status string
	switch mr.Status {
	case "open":
		status = StatusOpenStyle.Render("●")
	case "draft":
		status = StatusDraftStyle.Render("○")
	case "merged":
		status = StatusMergedStyle.Render("●")
	case "closed":
		status = StatusClosedStyle.Render("●")
	default:
		status = StatusDraftStyle.Render("○")
	}

	// Format: [indicator] [status] #number title (branch)
	number := fmt.Sprintf("#%-4d", mr.Number)
	title := mr.Title
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	var line string
	if isSelected {
		titleStyled := SelectedItemStyle.Render(fmt.Sprintf("%s %s", number, title))
		branchStyled := SelectedItemStyle.Render(fmt.Sprintf("(%s)", mr.Branch))
		content := fmt.Sprintf("%s %s %s", status, titleStyled, branchStyled)
		line = SelectedRowStyle.Render(content)
	} else {
		titleStyled := NormalItemStyle.Render(fmt.Sprintf("%s %s", number, title))
		branchStyled := DimStyle.Render(fmt.Sprintf("(%s)", mr.Branch))
		content := fmt.Sprintf("%s %s %s", status, titleStyled, branchStyled)
		line = NormalRowStyle.Render(content)
	}

	_, _ = fmt.Fprint(w, line)
}

// MRList is a bubbletea component for displaying MRs
type MRList struct {
	list   list.Model
	width  int
	height int
}

// NewMRList creates a new MR list component
func NewMRList(mrs []platform.MR, width, height int) MRList {
	items := make([]list.Item, len(mrs))
	for i, mr := range mrs {
		items[i] = MRItem{MR: mr}
	}

	l := list.New(items, CompactDelegate{}, width, height)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.TitleBar = lipgloss.NewStyle()
	l.Styles.Title = lipgloss.NewStyle()
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(accentColor)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(accentColor)

	return MRList{
		list:   l,
		width:  width,
		height: height,
	}
}

// SetItems updates the list items
func (m *MRList) SetItems(mrs []platform.MR) {
	items := make([]list.Item, len(mrs))
	for i, mr := range mrs {
		items[i] = MRItem{MR: mr}
	}
	m.list.SetItems(items)
}

// SelectedMR returns the currently selected MR, or nil if none
func (m MRList) SelectedMR() *platform.MR {
	item, ok := m.list.SelectedItem().(MRItem)
	if !ok {
		return nil
	}
	return &item.MR
}

// Update handles messages for the list
func (m MRList) Update(msg tea.Msg) (MRList, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the list
func (m MRList) View() string {
	return m.list.View()
}

// SetSize updates the list dimensions
func (m *MRList) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}
