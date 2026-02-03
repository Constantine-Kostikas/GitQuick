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

	_, _ = fmt.Fprint(w, line)
}

// AuthorPicker is a modal for selecting an author
type AuthorPicker struct {
	list        list.Model
	listWidth   int
	allAuthors  []platform.Author
	searching   bool
	searchInput textinput.Model
}

// NewAuthorPicker creates a new author picker
func NewAuthorPicker(authors []platform.Author, currentAuthor string, width, height int) AuthorPicker {
	// Add @me as the first option
	allAuthors := make([]platform.Author, 0, len(authors)+1)
	allAuthors = append(allAuthors, platform.Author{Username: "@me", Name: ""})
	allAuthors = append(allAuthors, authors...)

	items := make([]list.Item, len(allAuthors))
	for i, author := range allAuthors {
		items[i] = AuthorItem{Author: author}
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

	return AuthorPicker{
		list:        l,
		listWidth:   listWidth,
		allAuthors:  allAuthors,
		searchInput: ti,
	}
}

// Update handles messages
func (p AuthorPicker) Update(msg tea.Msg) (AuthorPicker, tea.Cmd) {
	// Handle search mode
	if p.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				p.searching = false
				p.searchInput.Blur()
				p.searchInput.SetValue("")
				p.filterItems()
				return p, nil
			case "enter":
				p.searching = false
				p.searchInput.Blur()
				return p, nil
			}
		}

		var cmd tea.Cmd
		p.searchInput, cmd = p.searchInput.Update(msg)
		p.filterItems()
		return p, cmd
	}

	// Not in search mode
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f", "/":
			p.searching = true
			p.searchInput.Focus()
			return p, textinput.Blink
		}
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

// filterItems filters the list based on search input
func (p *AuthorPicker) filterItems() {
	query := strings.ToLower(p.searchInput.Value())
	if query == "" {
		// Show all items
		items := make([]list.Item, len(p.allAuthors))
		for i, author := range p.allAuthors {
			items[i] = AuthorItem{Author: author}
		}
		p.list.SetItems(items)
		return
	}

	// Filter items
	var filtered []list.Item
	for _, author := range p.allAuthors {
		username := strings.ToLower(author.Username)
		name := strings.ToLower(author.Name)
		if strings.Contains(username, query) || strings.Contains(name, query) {
			filtered = append(filtered, AuthorItem{Author: author})
		}
	}
	p.list.SetItems(filtered)
}

// SearchBar returns the search bar view if searching, empty string otherwise
func (p AuthorPicker) SearchBar() string {
	if !p.searching {
		return ""
	}
	searchStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)
	return searchStyle.Render("Find: ") + p.searchInput.View()
}

// IsSearching returns true if search mode is active
func (p AuthorPicker) IsSearching() bool {
	return p.searching
}

// View renders the picker
func (p AuthorPicker) View() string {
	content := p.list.View()
	if p.searching {
		content = content + "\n" + p.SearchBar()
	}
	return ModalStyle.Width(p.listWidth).Render(content)
}

// SelectedAuthor returns the currently highlighted author
func (p AuthorPicker) SelectedAuthor() string {
	item, ok := p.list.SelectedItem().(AuthorItem)
	if !ok {
		return "@me"
	}
	return item.Author.Username
}
