package ui

import (
	tea "charm.land/bubbletea/v2"
)

// DirtyConfirmModal shows a warning when working tree has uncommitted changes
type DirtyConfirmModal struct {
	branch    string
	confirmed bool
	cancelled bool
}

// NewDirtyConfirmModal creates a new dirty confirm modal
func NewDirtyConfirmModal(branch string) DirtyConfirmModal {
	return DirtyConfirmModal{
		branch: branch,
	}
}

// Init returns the initial command
func (m DirtyConfirmModal) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m DirtyConfirmModal) Update(msg tea.Msg) (DirtyConfirmModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			m.confirmed = true
		case "n", "N", "esc":
			m.cancelled = true
		}
	}
	return m, nil
}

// View renders the modal
func (m DirtyConfirmModal) View() string {
	content := ErrorStyle.Render("Warning: Uncommitted Changes") + "\n\n"
	content += "You have uncommitted changes in your working tree.\n"
	content += "Checking out '" + m.branch + "' may cause conflicts.\n\n"
	content += "Proceed anyway?\n\n"
	content += "[y] Yes, checkout  |  [n] No, cancel"

	return ModalStyle.Render(content)
}

// IsConfirmed returns true if user confirmed
func (m DirtyConfirmModal) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if user cancelled
func (m DirtyConfirmModal) IsCancelled() bool {
	return m.cancelled
}
