package ui

import (
	"fmt"

	"github.com/Constantine-Kostikas/GitQuick/internal/git"
	"github.com/Constantine-Kostikas/GitQuick/internal/platform"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// CheckoutState represents the checkout status
type CheckoutState int

const (
	CheckoutInProgress CheckoutState = iota
	CheckoutDone
	CheckoutError
)

// CheckoutModal handles the checkout flow UI
type CheckoutModal struct {
	mr       *platform.MR // nil for direct branch checkout
	branch   string       // branch to checkout
	repoPath string
	state    CheckoutState
	spinner  spinner.Model
	err      error
	errStep  string
}

// CheckoutCompleteMsg is sent when checkout finishes
type CheckoutCompleteMsg struct {
	Err error
}

// NewCheckoutModal creates a new checkout modal for an MR
func NewCheckoutModal(mr platform.MR, repoPath string) CheckoutModal {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return CheckoutModal{
		mr:       &mr,
		branch:   mr.Branch,
		repoPath: repoPath,
		state:    CheckoutInProgress,
		spinner:  s,
	}
}

// NewBranchCheckoutModal creates a checkout modal for a direct branch checkout
func NewBranchCheckoutModal(branch, repoPath string) CheckoutModal {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return CheckoutModal{
		mr:       nil,
		branch:   branch,
		repoPath: repoPath,
		state:    CheckoutInProgress,
		spinner:  s,
	}
}

// Init starts the checkout process
func (m CheckoutModal) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.doCheckout())
}

func (m CheckoutModal) doCheckout() tea.Cmd {
	return func() tea.Msg {
		err := git.Checkout(m.repoPath, m.branch)
		return CheckoutCompleteMsg{Err: err}
	}
}

// Update handles messages
func (m CheckoutModal) Update(msg tea.Msg) (CheckoutModal, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case CheckoutCompleteMsg:
		if msg.Err != nil {
			m.state = CheckoutError
			m.err = msg.Err
			if ce, ok := msg.Err.(*git.CheckoutError); ok {
				m.errStep = ce.Step
			}
		} else {
			m.state = CheckoutDone
		}
		return m, nil
	}

	return m, nil
}

// View renders the modal
func (m CheckoutModal) View() string {
	var content string
	if m.mr != nil {
		content = fmt.Sprintf("#%d %s\n", m.mr.Number, m.mr.Title)
		content += fmt.Sprintf("Branch: %s\n\n", m.branch)
	} else {
		content = fmt.Sprintf("Checkout to default branch\n")
		content += fmt.Sprintf("Branch: %s\n\n", m.branch)
	}

	switch m.state {
	case CheckoutInProgress:
		content += m.spinner.View() + " Checking out...\n"
		content += "\n[esc] cancel"
	case CheckoutDone:
		content += SuccessStyle.Render("✓ Checkout complete") + "\n"
		content += "\nPress any key to continue"
	case CheckoutError:
		if m.errStep != "" {
			content += ErrorStyle.Render(fmt.Sprintf("✗ Failed at %s", m.errStep)) + "\n"
		}
		content += "\n" + ErrorStyle.Render("Error: "+m.err.Error())
		content += "\n\nPress any key to continue"
	}

	return ModalStyle.Render(content)
}

// IsDone returns true if checkout is complete (success or error)
func (m CheckoutModal) IsDone() bool {
	return m.state == CheckoutDone || m.state == CheckoutError
}

// HasError returns true if checkout failed
func (m CheckoutModal) HasError() bool {
	return m.state == CheckoutError
}
