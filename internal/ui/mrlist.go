package ui

import (
	"fmt"

	"gitHelper/internal/platform"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = "Merge Requests"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = ActiveTabStyle

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
