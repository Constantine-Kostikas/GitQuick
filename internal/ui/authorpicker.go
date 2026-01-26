package ui

import (
	"gitHelper/internal/platform"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AuthorItem wraps an Author for the list component
type AuthorItem struct {
	Author platform.Author
}

func (i AuthorItem) Title() string {
	return i.Author.Username
}

func (i AuthorItem) Description() string {
	if i.Author.Name != i.Author.Username {
		return i.Author.Name
	}
	return ""
}

func (i AuthorItem) FilterValue() string {
	return i.Author.Username
}

// AuthorPicker is a modal for selecting an author
type AuthorPicker struct {
	list     list.Model
	selected string
	width    int
	height   int
}

// NewAuthorPicker creates a new author picker
func NewAuthorPicker(authors []platform.Author, currentAuthor string, width, height int) AuthorPicker {
	// Add @me as the first option
	items := make([]list.Item, 0, len(authors)+1)
	items = append(items, AuthorItem{Author: platform.Author{Username: "@me", Name: ""}})

	for _, author := range authors {
		items = append(items, AuthorItem{Author: author})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("170"))

	listWidth := width - 4
	listHeight := height - 4
	if listWidth < 30 {
		listWidth = 30
	}
	if listHeight < 10 {
		listHeight = 10
	}

	l := list.New(items, delegate, listWidth, listHeight)
	l.Title = "Select Author"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = ActiveTabStyle

	return AuthorPicker{
		list:     l,
		selected: currentAuthor,
		width:    width,
		height:   height,
	}
}

// Update handles messages
func (p AuthorPicker) Update(msg tea.Msg) (AuthorPicker, tea.Cmd) {
	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

// View renders the picker
func (p AuthorPicker) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	return style.Render(p.list.View())
}

// SelectedAuthor returns the currently highlighted author
func (p AuthorPicker) SelectedAuthor() string {
	item, ok := p.list.SelectedItem().(AuthorItem)
	if !ok {
		return "@me"
	}
	return item.Author.Username
}
