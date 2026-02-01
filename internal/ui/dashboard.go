package ui

import (
	"fmt"
	"os/exec"
	"runtime"

	"gitHelper/internal/git"
	"gitHelper/internal/platform"

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

// PendingCheckout holds info about a checkout waiting for dirty confirmation
type PendingCheckout struct {
	MR     *platform.MR // nil for direct branch checkout
	Branch string
}

// Dashboard is the main UI component
type Dashboard struct {
	platform        platform.Platform
	repoInfo        platform.RepoInfo
	repoPath        string
	currentBranch   string
	author          string
	authors         []platform.Author
	authorPicker    *AuthorPicker
	activeTab       Tab
	mrList          MRList
	mrDetail        *MRDetailModal
	checkout        *CheckoutModal
	dirtyConfirm    *DirtyConfirmModal
	pendingCheckout *PendingCheckout
	width           int
	height          int
	err             error
	loading         bool
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

// AuthorsLoadedMsg is sent when authors are loaded
type AuthorsLoadedMsg struct {
	Authors []platform.Author
	Err     error
}

// BranchLoadedMsg is sent when current branch is loaded
type BranchLoadedMsg struct {
	Branch string
	Err    error
}

// DirtyCheckMsg is sent when dirty check completes
type DirtyCheckMsg struct {
	IsDirty bool
	Err     error
}

// NewDashboard creates a new dashboard
func NewDashboard(p platform.Platform, repoPath string) Dashboard {
	return Dashboard{
		platform:  p,
		repoPath:  repoPath,
		author:    "@me",
		activeTab: TabMRs,
		mrList:    NewMRList(nil, 80, 20),
		loading:   true,
	}
}

// Init loads initial data
func (d Dashboard) Init() tea.Cmd {
	return tea.Batch(
		d.loadRepoInfo(),
		d.loadMRs(),
		d.loadAuthors(),
		d.loadBranch(),
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

func (d Dashboard) loadAuthors() tea.Cmd {
	return func() tea.Msg {
		authors, err := d.platform.ListAuthors()
		return AuthorsLoadedMsg{Authors: authors, Err: err}
	}
}

func (d Dashboard) loadBranch() tea.Cmd {
	return func() tea.Msg {
		branch, err := git.GetCurrentBranch(d.repoPath)
		return BranchLoadedMsg{Branch: branch, Err: err}
	}
}

func (d Dashboard) loadMRDetail(number int) tea.Cmd {
	return func() tea.Msg {
		detail, err := d.platform.GetMRDetail(number)
		return MRDetailLoadedMsg{Detail: detail, Err: err}
	}
}

func (d Dashboard) checkDirty() tea.Cmd {
	return func() tea.Msg {
		isDirty, err := git.IsDirty(d.repoPath)
		return DirtyCheckMsg{IsDirty: isDirty, Err: err}
	}
}

// Update handles messages
func (d Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If dirty confirm modal is active, delegate to it
	if d.dirtyConfirm != nil {
		newConfirm, cmd := d.dirtyConfirm.Update(msg)
		d.dirtyConfirm = &newConfirm

		if d.dirtyConfirm.IsConfirmed() {
			// User confirmed, proceed with checkout
			d.dirtyConfirm = nil
			if d.pendingCheckout != nil {
				if d.pendingCheckout.MR != nil {
					checkout := NewCheckoutModal(*d.pendingCheckout.MR, d.repoPath)
					d.checkout = &checkout
				} else {
					checkout := NewBranchCheckoutModal(d.pendingCheckout.Branch, d.repoPath)
					d.checkout = &checkout
				}
				d.pendingCheckout = nil
				return d, d.checkout.Init()
			}
		} else if d.dirtyConfirm.IsCancelled() {
			// User cancelled
			d.dirtyConfirm = nil
			d.pendingCheckout = nil
		}
		return d, cmd
	}

	// If checkout modal is active, delegate to it
	if d.checkout != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if d.checkout.IsDone() {
				d.checkout = nil
				return d, d.loadBranch()
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

	// If MR detail modal is active, delegate to it
	if d.mrDetail != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				d.mrDetail = nil
				return d, nil
			}
		case MRDetailLoadedMsg:
			d.mrDetail.SetDetail(msg.Detail, msg.Err)
			return d, nil
		}

		newDetail, cmd := d.mrDetail.Update(msg)
		d.mrDetail = &newDetail

		// Check if user wants to proceed to checkout
		if d.mrDetail.WantsCheckout() {
			mr := d.mrDetail.GetMR()
			d.mrDetail = nil
			// Store pending checkout and check dirty state
			d.pendingCheckout = &PendingCheckout{MR: &mr, Branch: mr.Branch}
			return d, d.checkDirty()
		}

		return d, cmd
	}

	// Handle author picker modal
	if d.authorPicker != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				d.author = d.authorPicker.SelectedAuthor()
				d.authorPicker = nil
				d.loading = true
				return d, d.loadMRs()
			case "esc":
				d.authorPicker = nil
				return d, nil
			}
		}
		newPicker, cmd := d.authorPicker.Update(msg)
		d.authorPicker = &newPicker
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
			picker := NewAuthorPicker(d.authors, d.author, d.width-10, d.height-6)
			d.authorPicker = &picker
			return d, nil
		case "tab":
			d.activeTab = (d.activeTab + 1) % 3
			return d, nil
		case "r", "R":
			if d.activeTab == TabMRs && !d.loading {
				d.loading = true
				return d, d.loadMRs()
			}
			return d, nil
		case "m":
			// Checkout to default branch
			if d.repoInfo.DefaultBranch != "" && d.currentBranch != d.repoInfo.DefaultBranch {
				// Store pending checkout and check dirty state
				d.pendingCheckout = &PendingCheckout{MR: nil, Branch: d.repoInfo.DefaultBranch}
				return d, d.checkDirty()
			}
			return d, nil
		case "enter":
			if d.activeTab == TabMRs {
				if mr := d.mrList.SelectedMR(); mr != nil {
					// Open MR detail modal first (checkout happens from there)
					detail := NewMRDetailModal(*mr, d.repoInfo.Platform, d.width, d.height)
					d.mrDetail = &detail
					return d, tea.Batch(d.mrDetail.Init(), d.loadMRDetail(mr.Number))
				}
			}
			return d, nil
		case "w":
			if d.activeTab == TabMRs {
				if mr := d.mrList.SelectedMR(); mr != nil && mr.URL != "" {
					openBrowser(mr.URL)
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

	case AuthorsLoadedMsg:
		if msg.Err == nil {
			d.authors = msg.Authors
		}
		return d, nil

	case BranchLoadedMsg:
		if msg.Err == nil {
			d.currentBranch = msg.Branch
		}
		return d, nil

	case DirtyCheckMsg:
		if d.pendingCheckout == nil {
			return d, nil
		}
		// If error checking dirty, proceed anyway
		if msg.Err != nil || !msg.IsDirty {
			// Not dirty or error, proceed to checkout
			if d.pendingCheckout.MR != nil {
				checkout := NewCheckoutModal(*d.pendingCheckout.MR, d.repoPath)
				d.checkout = &checkout
			} else {
				checkout := NewBranchCheckoutModal(d.pendingCheckout.Branch, d.repoPath)
				d.checkout = &checkout
			}
			d.pendingCheckout = nil
			return d, d.checkout.Init()
		}
		// Dirty, show confirmation
		confirm := NewDirtyConfirmModal(d.pendingCheckout.Branch)
		d.dirtyConfirm = &confirm
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

	// Footer
	footer := d.renderFooter()

	// Calculate content height
	chromeHeight := lipgloss.Height(header) + lipgloss.Height(authorRow) + lipgloss.Height(tabs) + lipgloss.Height(footer)
	contentHeight := d.height - chromeHeight

	// Content
	var rawContent string
	if d.loading {
		rawContent = "Loading..."
	} else if d.err != nil {
		rawContent = ErrorStyle.Render("Error: " + d.err.Error())
	} else {
		switch d.activeTab {
		case TabMRs:
			rawContent = d.mrList.View()
		case TabIssues:
			rawContent = "Issues tab (coming soon)"
		case TabBranches:
			rawContent = "Branches tab (coming soon)"
		}
	}

	content := lipgloss.NewStyle().
		Width(d.width).
		Height(contentHeight).
		Render(rawContent)

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

	// Overlay MR detail modal if active
	if d.mrDetail != nil {
		modalView := d.mrDetail.View()
		view = lipgloss.Place(d.width, d.height,
			lipgloss.Center, lipgloss.Center,
			modalView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
		)
	}

	// Overlay author picker if active
	if d.authorPicker != nil {
		modalView := d.authorPicker.View()
		view = lipgloss.Place(d.width, d.height,
			lipgloss.Center, lipgloss.Center,
			modalView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
		)
	}

	// Overlay dirty confirm if active
	if d.dirtyConfirm != nil {
		modalView := d.dirtyConfirm.View()
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
	branch := d.currentBranch
	if branch == "" {
		branch = "..."
	}
	platformName := d.repoInfo.Platform
	if platformName == "" {
		platformName = "..."
	}

	left := fmt.Sprintf("repo: %s", repoName)
	middle := fmt.Sprintf("branch: %s", branch)
	right := fmt.Sprintf("platform: %s", platformName)

	// Calculate gaps to spread items evenly
	totalContent := len(left) + len(middle) + len(right)
	totalGap := d.width - totalContent - 4
	gap := totalGap / 2
	if gap < 1 {
		gap = 1
	}

	return HeaderStyle.Width(d.width).Render(
		left + fmt.Sprintf("%*s", gap, "") + middle + fmt.Sprintf("%*s", gap, "") + right,
	)
}

func (d Dashboard) renderAuthorRow() string {
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
	help := "↑↓ nav │ enter details │ w open │ r refresh │ a author │ m main │ tab switch │ q quit"
	return FooterStyle.Align(lipgloss.Center).Width(d.width).Render(help)
}

// openBrowser opens a URL in the default browser
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
