package ui

import (
	"fmt"
	"io"
	"strings"

	"gitHelper/internal/platform"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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
	list        list.Model
	allItems    []platform.MR // All MRs (unfiltered)
	width       int
	height      int
	searching   bool
	searchInput textinput.Model
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
	l.SetFilteringEnabled(false) // We handle filtering ourselves
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.TitleBar = lipgloss.NewStyle()
	l.Styles.Title = lipgloss.NewStyle()

	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.CharLimit = 50
	ti.Width = 30
	ti.PromptStyle = lipgloss.NewStyle().Foreground(accentColor)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	return MRList{
		list:        l,
		allItems:    mrs,
		width:       width,
		height:      height,
		searchInput: ti,
	}
}

// SetItems updates the list items
func (m *MRList) SetItems(mrs []platform.MR) {
	m.allItems = mrs
	items := make([]list.Item, len(mrs))
	for i, mr := range mrs {
		items[i] = MRItem{MR: mr}
	}
	m.list.SetItems(items)
	// Clear search when new items are set
	m.searching = false
	m.searchInput.SetValue("")
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
	// Handle search mode
	if m.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "enter":
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			}
		}

		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)

		// Filter items based on search
		m.filterItems()

		return m, cmd
	}

	// Not in search mode
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f", "/":
			m.searching = true
			m.searchInput.Focus()
			return m, textinput.Blink
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// filterItems filters the list based on search input
func (m *MRList) filterItems() {
	query := strings.ToLower(m.searchInput.Value())
	if query == "" {
		// Show all items
		items := make([]list.Item, len(m.allItems))
		for i, mr := range m.allItems {
			items[i] = MRItem{MR: mr}
		}
		m.list.SetItems(items)
		return
	}

	// Filter items
	var filtered []list.Item
	for _, mr := range m.allItems {
		title := strings.ToLower(mr.Title)
		branch := strings.ToLower(mr.Branch)
		number := fmt.Sprintf("#%d", mr.Number)
		if strings.Contains(title, query) || strings.Contains(branch, query) || strings.Contains(number, query) {
			filtered = append(filtered, MRItem{MR: mr})
		}
	}
	m.list.SetItems(filtered)
}

// View renders the list
func (m MRList) View() string {
	return m.list.View()
}

// SearchBar returns the search bar view if searching, empty string otherwise
func (m MRList) SearchBar() string {
	if !m.searching {
		return ""
	}
	searchStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)
	return searchStyle.Render("Find: ") + m.searchInput.View()
}

// IsSearching returns true if search mode is active
func (m MRList) IsSearching() bool {
	return m.searching
}

// SetSize updates the list dimensions
func (m *MRList) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}
