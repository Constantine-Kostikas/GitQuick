package ui

import (
	"fmt"

	"gitHelper/internal/git"
	"gitHelper/internal/platform"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// CheckoutState represents the current step in checkout
type CheckoutState int

const (
	CheckoutFetching CheckoutState = iota
	CheckoutCheckingOut
	CheckoutPulling
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
	steps    []stepStatus
}

type stepStatus struct {
	name   string
	done   bool
	failed bool
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
		state:    CheckoutFetching,
		spinner:  s,
		steps: []stepStatus{
			{name: "Fetching origin", done: false},
			{name: "Checking out " + mr.Branch, done: false},
			{name: "Pulling latest changes", done: false},
		},
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
			// Mark failed step based on error
			if ce, ok := msg.Err.(*git.CheckoutError); ok {
				for i := range m.steps {
					if i == 0 && ce.Step == "fetch" {
						m.steps[i].failed = true
						break
					} else if i == 1 && ce.Step == "checkout" {
						m.steps[0].done = true
						m.steps[i].failed = true
						break
					} else if i == 2 && ce.Step == "pull" {
						m.steps[0].done = true
						m.steps[1].done = true
						m.steps[i].failed = true
						break
					}
				}
			}
		} else {
			m.state = CheckoutDone
			for i := range m.steps {
				m.steps[i].done = true
			}
		}
		return m, nil
	}

	return m, nil
}

// View renders the modal
func (m CheckoutModal) View() string {
	content := fmt.Sprintf("#%d %s\n", m.mr.Number, m.mr.Title)
	content += fmt.Sprintf("Branch: %s\n\n", m.mr.Branch)

	for _, step := range m.steps {
		if step.failed {
			content += ErrorStyle.Render("✗ "+step.name) + "\n"
		} else if step.done {
			content += SuccessStyle.Render("✓ "+step.name) + "\n"
		} else if m.state != CheckoutDone && m.state != CheckoutError {
			content += m.spinner.View() + " " + step.name + "...\n"
			break // Only show spinner for current step
		} else {
			content += "  " + step.name + "\n"
		}
	}

	if m.err != nil {
		content += "\n" + ErrorStyle.Render("Error: "+m.err.Error())
	}

	if m.state == CheckoutDone || m.state == CheckoutError {
		content += "\n\nPress any key to continue"
	} else {
		content += "\n\n[esc] cancel"
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
