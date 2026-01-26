package ui

import (
	"fmt"

	"gitHelper/internal/git"
	"gitHelper/internal/platform"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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
	mr       platform.MR
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

// NewCheckoutModal creates a new checkout modal
func NewCheckoutModal(mr platform.MR, repoPath string) CheckoutModal {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return CheckoutModal{
		mr:       mr,
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
		err := git.Checkout(m.repoPath, m.mr.Branch)
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
	content := fmt.Sprintf("#%d %s\n", m.mr.Number, m.mr.Title)
	content += fmt.Sprintf("Branch: %s\n\n", m.mr.Branch)

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
