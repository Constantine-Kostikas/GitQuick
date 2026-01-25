package ui

import (
	"fmt"

	"gitHelper/internal/platform"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tab represents a dashboard tab
type Tab int

const (
	TabMRs Tab = iota
	TabIssues
	TabBranches
)

// Dashboard is the main UI component
type Dashboard struct {
	platform      platform.Platform
	repoInfo      platform.RepoInfo
	repoPath      string
	author        string
	authorInput   textinput.Model
	editingAuthor bool
	activeTab     Tab
	mrList        MRList
	checkout      *CheckoutModal
	width         int
	height        int
	err           error
	loading       bool
}

// MRsLoadedMsg is sent when MRs are loaded
type MRsLoadedMsg struct {
	MRs []platform.MR
	Err error
}

// RepoInfoLoadedMsg is sent when repo info is loaded
type RepoInfoLoadedMsg struct {
	Info platform.RepoInfo
	Err  error
}

// NewDashboard creates a new dashboard
func NewDashboard(p platform.Platform, repoPath string) Dashboard {
	ti := textinput.New()
	ti.Placeholder = "username or @me"
	ti.CharLimit = 50

	return Dashboard{
		platform:    p,
		repoPath:    repoPath,
		author:      "@me",
		authorInput: ti,
		activeTab:   TabMRs,
		mrList:      NewMRList(nil, 80, 20),
		loading:     true,
	}
}

// Init loads initial data
func (d Dashboard) Init() tea.Cmd {
	return tea.Batch(
		d.loadRepoInfo(),
		d.loadMRs(),
	)
}

func (d Dashboard) loadRepoInfo() tea.Cmd {
	return func() tea.Msg {
		info, err := d.platform.GetRepoInfo()
		return RepoInfoLoadedMsg{Info: info, Err: err}
	}
}

func (d Dashboard) loadMRs() tea.Cmd {
	return func() tea.Msg {
		mrs, err := d.platform.ListMRs(d.author)
		return MRsLoadedMsg{MRs: mrs, Err: err}
	}
}

// Update handles messages
func (d Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If checkout modal is active, delegate to it
	if d.checkout != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if d.checkout.IsDone() {
				d.checkout = nil
				return d, nil
			}
			if msg.String() == "esc" {
				d.checkout = nil
				return d, nil
			}
		}
		newCheckout, cmd := d.checkout.Update(msg)
		d.checkout = &newCheckout
		return d, cmd
	}

	// Handle author input mode
	if d.editingAuthor {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				d.author = d.authorInput.Value()
				if d.author == "" {
					d.author = "@me"
				}
				d.editingAuthor = false
				d.loading = true
				return d, d.loadMRs()
			case "esc":
				d.editingAuthor = false
				d.authorInput.SetValue(d.author)
				return d, nil
			}
		}
		var cmd tea.Cmd
		d.authorInput, cmd = d.authorInput.Update(msg)
		return d, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		d.mrList.SetSize(msg.Width-4, msg.Height-10)
		return d, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return d, tea.Quit
		case "a":
			d.editingAuthor = true
			d.authorInput.SetValue(d.author)
			d.authorInput.Focus()
			return d, textinput.Blink
		case "tab":
			d.activeTab = (d.activeTab + 1) % 3
			return d, nil
		case "enter":
			if d.activeTab == TabMRs {
				if mr := d.mrList.SelectedMR(); mr != nil {
					checkout := NewCheckoutModal(*mr, d.repoPath)
					d.checkout = &checkout
					return d, d.checkout.Init()
				}
			}
			return d, nil
		}

	case RepoInfoLoadedMsg:
		if msg.Err != nil {
			d.err = msg.Err
		} else {
			d.repoInfo = msg.Info
		}
		return d, nil

	case MRsLoadedMsg:
		d.loading = false
		if msg.Err != nil {
			d.err = msg.Err
		} else {
			d.mrList.SetItems(msg.MRs)
		}
		return d, nil
	}

	// Pass to MR list
	if d.activeTab == TabMRs && !d.loading {
		var cmd tea.Cmd
		d.mrList, cmd = d.mrList.Update(msg)
		return d, cmd
	}

	return d, nil
}

// View renders the dashboard
func (d Dashboard) View() string {
	if d.width == 0 {
		return "Loading..."
	}

	// Header
	header := d.renderHeader()

	// Author row
	authorRow := d.renderAuthorRow()

	// Tabs
	tabs := d.renderTabs()

	// Content
	var content string
	if d.loading {
		content = "\n  Loading...\n"
	} else if d.err != nil {
		content = "\n  " + ErrorStyle.Render("Error: "+d.err.Error()) + "\n"
	} else {
		switch d.activeTab {
		case TabMRs:
			content = d.mrList.View()
		case TabIssues:
			content = "\n  Issues tab (coming soon)\n"
		case TabBranches:
			content = "\n  Branches tab (coming soon)\n"
		}
	}

	// Footer
	footer := d.renderFooter()

	// Combine
	view := lipgloss.JoinVertical(lipgloss.Left,
		header,
		authorRow,
		tabs,
		content,
		footer,
	)

	// Overlay checkout modal if active
	if d.checkout != nil {
		modalView := d.checkout.View()
		// Center the modal (simplified)
		view = lipgloss.Place(d.width, d.height,
			lipgloss.Center, lipgloss.Center,
			modalView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
		)
	}

	return view
}

func (d Dashboard) renderHeader() string {
	repoName := d.repoInfo.Name
	if repoName == "" {
		repoName = "..."
	}
	platformName := d.repoInfo.Platform
	if platformName == "" {
		platformName = "..."
	}

	left := fmt.Sprintf("repo: %s", repoName)
	right := fmt.Sprintf("platform: %s", platformName)

	gap := d.width - len(left) - len(right) - 4
	if gap < 1 {
		gap = 1
	}

	return HeaderStyle.Width(d.width).Render(
		left + fmt.Sprintf("%*s", gap, "") + right,
	)
}

func (d Dashboard) renderAuthorRow() string {
	if d.editingAuthor {
		return fmt.Sprintf("  Author: %s", d.authorInput.View())
	}
	return fmt.Sprintf("  Author: [%s]", d.author)
}

func (d Dashboard) renderTabs() string {
	tabs := []string{"MRs", "Issues", "Branches"}
	var rendered []string
	for i, tab := range tabs {
		if Tab(i) == d.activeTab {
			rendered = append(rendered, ActiveTabStyle.Render(tab))
		} else {
			rendered = append(rendered, InactiveTabStyle.Render(tab))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (d Dashboard) renderFooter() string {
	help := "↑↓ navigate │ enter checkout │ a author │ tab switch │ q quit"
	return FooterStyle.Width(d.width).Render(help)
}
