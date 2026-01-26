package ui

import (
	"fmt"
	"io"

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

// AuthorDelegate is a custom delegate for author list rendering
type AuthorDelegate struct{}

func (d AuthorDelegate) Height() int                             { return 3 }
func (d AuthorDelegate) Spacing() int                            { return 0 }
func (d AuthorDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d AuthorDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	authorItem, ok := item.(AuthorItem)
	if !ok {
		return
	}

	author := authorItem.Author
	isSelected := index == m.Index()

	username := author.Username
	name := author.Name

	var content string
	if name != "" && name != username {
		content = fmt.Sprintf("%s (%s)", username, name)
	} else {
		content = username
	}

	var line string
	if isSelected {
		line = SelectedRowStyle.Render(SelectedItemStyle.Render(content))
	} else {
		line = NormalRowStyle.Render(NormalItemStyle.Render(content))
	}

	fmt.Fprint(w, line)
}

// AuthorPicker is a modal for selecting an author
type AuthorPicker struct {
	list      list.Model
	selected  string
	listWidth int
	width     int
	height    int
}

// NewAuthorPicker creates a new author picker
func NewAuthorPicker(authors []platform.Author, currentAuthor string, width, height int) AuthorPicker {
	// Add @me as the first option
	items := make([]list.Item, 0, len(authors)+1)
	items = append(items, AuthorItem{Author: platform.Author{Username: "@me", Name: ""}})

	for _, author := range authors {
		items = append(items, AuthorItem{Author: author})
	}

	listWidth := width - 4
	listHeight := height - 4
	if listWidth < 40 {
		listWidth = 40
	}
	if listHeight < 15 {
		listHeight = 15
	}

	l := list.New(items, AuthorDelegate{}, listWidth, listHeight)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.TitleBar = lipgloss.NewStyle()
	l.Styles.Title = lipgloss.NewStyle()
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(accentColor)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(accentColor)

	return AuthorPicker{
		list:      l,
		selected:  currentAuthor,
		listWidth: listWidth,
		width:     width,
		height:    height,
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
	return ModalStyle.Width(p.listWidth).Render(p.list.View())
}

// SelectedAuthor returns the currently highlighted author
func (p AuthorPicker) SelectedAuthor() string {
	item, ok := p.list.SelectedItem().(AuthorItem)
	if !ok {
		return "@me"
	}
	return item.Author.Username
}
