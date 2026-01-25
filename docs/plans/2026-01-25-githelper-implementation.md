# gitHelper Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a TUI tool that lists MRs from GitHub/GitLab and checks out selected branches.

**Architecture:** Platform abstraction layer wraps `gh`/`glab` CLIs. Bubbletea dashboard shows MR list with author filter. Checkout flow runs fetch/checkout/pull sequence.

**Tech Stack:** Go, Bubbletea, Lipgloss, Bubbles, gh CLI, glab CLI

---

## Task 1: Project Setup

**Files:**
- Modify: `go.mod`
- Modify: `main.go`

**Step 1: Add dependencies**

Run:
```bash
cd /home/dev/development/projects/gitHelper && go get github.com/charmbracelet/bubbletea github.com/charmbracelet/lipgloss github.com/charmbracelet/bubbles
```

**Step 2: Create minimal bubbletea app**

Replace `main.go` with:

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return "gitHelper - press q to quit\n"
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 3: Verify it runs**

Run: `go run main.go`
Expected: Shows "gitHelper - press q to quit", pressing q exits cleanly.

**Step 4: Commit**

```bash
git add go.mod go.sum main.go
git commit -m "feat: initialize bubbletea app skeleton"
```

---

## Task 2: Platform Types and Interface

**Files:**
- Create: `internal/platform/platform.go`

**Step 1: Create platform package with types**

```go
package platform

// MR represents a merge/pull request
type MR struct {
	Number int
	Title  string
	Branch string
	Status string // "open", "draft", "merged", "closed"
	URL    string
}

// RepoInfo contains repository metadata
type RepoInfo struct {
	Name          string
	Description   string
	Platform      string // "github" or "gitlab"
	DefaultBranch string
}

// Platform abstracts GitHub/GitLab operations
type Platform interface {
	ListMRs(author string) ([]MR, error)
	GetRepoInfo() (RepoInfo, error)
}
```

**Step 2: Verify it compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/platform/platform.go
git commit -m "feat: add platform interface and types"
```

---

## Task 3: Git Operations - Checkout Flow

**Files:**
- Create: `internal/git/git.go`
- Create: `internal/git/git_test.go`

**Step 1: Write failing test for IsGitRepo**

```go
package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGitRepo(t *testing.T) {
	// Create temp dir without .git
	tmpDir, err := os.MkdirTemp("", "githelper-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if IsGitRepo(tmpDir) {
		t.Error("expected false for non-git directory")
	}

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	if !IsGitRepo(tmpDir) {
		t.Error("expected true for git directory")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/git/...`
Expected: FAIL - package doesn't exist

**Step 3: Write minimal implementation**

```go
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsGitRepo checks if the given path is inside a git repository
func IsGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/git/...`
Expected: PASS

**Step 5: Write failing test for GetRemoteURL**

Add to `git_test.go`:

```go
func TestGetRemoteURL(t *testing.T) {
	// This test requires being in an actual git repo
	// We'll test error case for non-git dir
	tmpDir, err := os.MkdirTemp("", "githelper-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = GetRemoteURL(tmpDir)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}
```

**Step 6: Run test to verify it fails**

Run: `go test ./internal/git/...`
Expected: FAIL - GetRemoteURL not defined

**Step 7: Implement GetRemoteURL**

Add to `git.go`:

```go
// GetRemoteURL returns the origin remote URL for the repo at path
func GetRemoteURL(path string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
```

**Step 8: Run test to verify it passes**

Run: `go test ./internal/git/...`
Expected: PASS

**Step 9: Write failing test for Checkout**

Add to `git_test.go`:

```go
func TestCheckout_NotGitRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "githelper-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	err = Checkout(tmpDir, "main")
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}
```

**Step 10: Run test to verify it fails**

Run: `go test ./internal/git/...`
Expected: FAIL - Checkout not defined

**Step 11: Implement Checkout with fetch and pull**

Add to `git.go`:

```go
// CheckoutResult contains the result of a checkout operation
type CheckoutResult struct {
	Fetched    bool
	CheckedOut bool
	Pulled     bool
	Error      error
	Step       string // which step failed
}

// Checkout fetches, checks out the branch, and pulls
func Checkout(path, branch string) error {
	// Fetch
	fetchCmd := exec.Command("git", "fetch", "origin")
	fetchCmd.Dir = path
	if err := fetchCmd.Run(); err != nil {
		return &CheckoutError{Step: "fetch", Err: err}
	}

	// Checkout
	checkoutCmd := exec.Command("git", "checkout", branch)
	checkoutCmd.Dir = path
	if err := checkoutCmd.Run(); err != nil {
		return &CheckoutError{Step: "checkout", Err: err}
	}

	// Pull
	pullCmd := exec.Command("git", "pull")
	pullCmd.Dir = path
	if err := pullCmd.Run(); err != nil {
		return &CheckoutError{Step: "pull", Err: err}
	}

	return nil
}

// CheckoutError wraps an error with the step that failed
type CheckoutError struct {
	Step string
	Err  error
}

func (e *CheckoutError) Error() string {
	return e.Step + ": " + e.Err.Error()
}

func (e *CheckoutError) Unwrap() error {
	return e.Err
}
```

**Step 12: Run tests**

Run: `go test ./internal/git/...`
Expected: PASS

**Step 13: Commit**

```bash
git add internal/git/
git commit -m "feat: add git operations (IsGitRepo, GetRemoteURL, Checkout)"
```

---

## Task 4: Platform Detection

**Files:**
- Create: `internal/platform/detect.go`
- Create: `internal/platform/detect_test.go`

**Step 1: Write failing test for DetectPlatform**

```go
package platform

import "testing"

func TestDetectPlatformFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/user/repo.git", "github"},
		{"git@github.com:user/repo.git", "github"},
		{"https://gitlab.com/user/repo.git", "gitlab"},
		{"git@gitlab.com:user/repo.git", "gitlab"},
		{"https://example.com/user/repo.git", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := DetectPlatformFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/...`
Expected: FAIL - DetectPlatformFromURL not defined

**Step 3: Implement DetectPlatformFromURL**

```go
package platform

import "strings"

// DetectPlatformFromURL returns "github", "gitlab", or "" based on remote URL
func DetectPlatformFromURL(url string) string {
	urlLower := strings.ToLower(url)
	if strings.Contains(urlLower, "github.com") {
		return "github"
	}
	if strings.Contains(urlLower, "gitlab.com") {
		return "gitlab"
	}
	return ""
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/platform/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/platform/detect.go internal/platform/detect_test.go
git commit -m "feat: add platform detection from remote URL"
```

---

## Task 5: GitHub Platform Implementation

**Files:**
- Create: `internal/platform/github.go`
- Create: `internal/platform/github_test.go`

**Step 1: Write test for parsing gh CLI output**

```go
package platform

import (
	"reflect"
	"testing"
)

func TestGitHub_ParseMRList(t *testing.T) {
	jsonOutput := `[
		{"number": 142, "title": "Fix login timeout", "headRefName": "feature/login", "state": "OPEN", "url": "https://github.com/org/repo/pull/142"},
		{"number": 138, "title": "Add user preferences", "headRefName": "user-prefs", "state": "DRAFT", "url": "https://github.com/org/repo/pull/138"}
	]`

	mrs, err := parseGitHubMRs([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []MR{
		{Number: 142, Title: "Fix login timeout", Branch: "feature/login", Status: "open", URL: "https://github.com/org/repo/pull/142"},
		{Number: 138, Title: "Add user preferences", Branch: "user-prefs", Status: "draft", URL: "https://github.com/org/repo/pull/138"},
	}

	if !reflect.DeepEqual(mrs, expected) {
		t.Errorf("got %+v, want %+v", mrs, expected)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/...`
Expected: FAIL - parseGitHubMRs not defined

**Step 3: Implement GitHub platform**

```go
package platform

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// GitHub implements Platform for GitHub repositories
type GitHub struct {
	repoPath string
}

// NewGitHub creates a GitHub platform instance
func NewGitHub(repoPath string) *GitHub {
	return &GitHub{repoPath: repoPath}
}

// ghPR represents the JSON structure from gh pr list
type ghPR struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	HeadRefName string `json:"headRefName"`
	State       string `json:"state"`
	URL         string `json:"url"`
}

func parseGitHubMRs(data []byte) ([]MR, error) {
	var prs []ghPR
	if err := json.Unmarshal(data, &prs); err != nil {
		return nil, err
	}

	mrs := make([]MR, len(prs))
	for i, pr := range prs {
		mrs[i] = MR{
			Number: pr.Number,
			Title:  pr.Title,
			Branch: pr.HeadRefName,
			Status: strings.ToLower(pr.State),
			URL:    pr.URL,
		}
	}
	return mrs, nil
}

// ListMRs returns pull requests for the given author
func (g *GitHub) ListMRs(author string) ([]MR, error) {
	cmd := exec.Command("gh", "pr", "list",
		"--author", author,
		"--json", "number,title,headRefName,state,url",
	)
	cmd.Dir = g.repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseGitHubMRs(out)
}

// GetRepoInfo returns repository information
func (g *GitHub) GetRepoInfo() (RepoInfo, error) {
	cmd := exec.Command("gh", "repo", "view", "--json", "name,description,defaultBranchRef")
	cmd.Dir = g.repoPath
	out, err := cmd.Output()
	if err != nil {
		return RepoInfo{}, err
	}

	var result struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		DefaultBranchRef struct {
			Name string `json:"name"`
		} `json:"defaultBranchRef"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return RepoInfo{}, err
	}

	return RepoInfo{
		Name:          result.Name,
		Description:   result.Description,
		Platform:      "github",
		DefaultBranch: result.DefaultBranchRef.Name,
	}, nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/platform/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/platform/github.go internal/platform/github_test.go
git commit -m "feat: add GitHub platform implementation"
```

---

## Task 6: GitLab Platform Implementation

**Files:**
- Create: `internal/platform/gitlab.go`
- Create: `internal/platform/gitlab_test.go`

**Step 1: Write test for parsing glab CLI output**

```go
package platform

import (
	"reflect"
	"testing"
)

func TestGitLab_ParseMRList(t *testing.T) {
	jsonOutput := `[
		{"iid": 42, "title": "Update API docs", "source_branch": "docs/api", "state": "opened", "web_url": "https://gitlab.com/org/repo/-/merge_requests/42"},
		{"iid": 40, "title": "Fix CI pipeline", "source_branch": "fix/ci", "state": "merged", "web_url": "https://gitlab.com/org/repo/-/merge_requests/40"}
	]`

	mrs, err := parseGitLabMRs([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []MR{
		{Number: 42, Title: "Update API docs", Branch: "docs/api", Status: "open", URL: "https://gitlab.com/org/repo/-/merge_requests/42"},
		{Number: 40, Title: "Fix CI pipeline", Branch: "fix/ci", Status: "merged", URL: "https://gitlab.com/org/repo/-/merge_requests/40"},
	}

	if !reflect.DeepEqual(mrs, expected) {
		t.Errorf("got %+v, want %+v", mrs, expected)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/...`
Expected: FAIL - parseGitLabMRs not defined

**Step 3: Implement GitLab platform**

```go
package platform

import (
	"encoding/json"
	"os/exec"
)

// GitLab implements Platform for GitLab repositories
type GitLab struct {
	repoPath string
}

// NewGitLab creates a GitLab platform instance
func NewGitLab(repoPath string) *GitLab {
	return &GitLab{repoPath: repoPath}
}

// glabMR represents the JSON structure from glab mr list
type glabMR struct {
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	SourceBranch string `json:"source_branch"`
	State        string `json:"state"`
	WebURL       string `json:"web_url"`
}

func parseGitLabMRs(data []byte) ([]MR, error) {
	var gMRs []glabMR
	if err := json.Unmarshal(data, &gMRs); err != nil {
		return nil, err
	}

	mrs := make([]MR, len(gMRs))
	for i, mr := range gMRs {
		status := mr.State
		if status == "opened" {
			status = "open"
		}
		mrs[i] = MR{
			Number: mr.IID,
			Title:  mr.Title,
			Branch: mr.SourceBranch,
			Status: status,
			URL:    mr.WebURL,
		}
	}
	return mrs, nil
}

// ListMRs returns merge requests for the given author
func (g *GitLab) ListMRs(author string) ([]MR, error) {
	args := []string{"mr", "list", "-F", "json"}
	if author == "@me" {
		args = append(args, "--author", "@me")
	} else {
		args = append(args, "--author", author)
	}

	cmd := exec.Command("glab", args...)
	cmd.Dir = g.repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseGitLabMRs(out)
}

// GetRepoInfo returns repository information
func (g *GitLab) GetRepoInfo() (RepoInfo, error) {
	cmd := exec.Command("glab", "repo", "view", "-F", "json")
	cmd.Dir = g.repoPath
	out, err := cmd.Output()
	if err != nil {
		return RepoInfo{}, err
	}

	var result struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return RepoInfo{}, err
	}

	return RepoInfo{
		Name:          result.Name,
		Description:   result.Description,
		Platform:      "gitlab",
		DefaultBranch: result.DefaultBranch,
	}, nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/platform/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/platform/gitlab.go internal/platform/gitlab_test.go
git commit -m "feat: add GitLab platform implementation"
```

---

## Task 7: Platform Factory

**Files:**
- Modify: `internal/platform/detect.go`
- Modify: `internal/platform/detect_test.go`

**Step 1: Add factory function test**

Add to `detect_test.go`:

```go
func TestNewPlatform(t *testing.T) {
	// Test that NewPlatform returns correct type based on URL
	tests := []struct {
		url          string
		expectedType string
	}{
		{"https://github.com/user/repo.git", "*platform.GitHub"},
		{"git@gitlab.com:user/repo.git", "*platform.GitLab"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			p, err := NewPlatform("/tmp", tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			typeName := reflect.TypeOf(p).String()
			if typeName != tt.expectedType {
				t.Errorf("got %s, want %s", typeName, tt.expectedType)
			}
		})
	}
}

func TestNewPlatform_Unknown(t *testing.T) {
	_, err := NewPlatform("/tmp", "https://example.com/repo.git")
	if err == nil {
		t.Error("expected error for unknown platform")
	}
}
```

Add import for `reflect` at top of test file.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/...`
Expected: FAIL - NewPlatform not defined

**Step 3: Implement NewPlatform factory**

Add to `detect.go`:

```go
import "errors"

// ErrUnknownPlatform is returned when the remote URL doesn't match known platforms
var ErrUnknownPlatform = errors.New("unknown platform: only github.com and gitlab.com are supported")

// NewPlatform creates the appropriate Platform implementation based on remote URL
func NewPlatform(repoPath, remoteURL string) (Platform, error) {
	switch DetectPlatformFromURL(remoteURL) {
	case "github":
		return NewGitHub(repoPath), nil
	case "gitlab":
		return NewGitLab(repoPath), nil
	default:
		return nil, ErrUnknownPlatform
	}
}
```

**Step 4: Run tests**

Run: `go test ./internal/platform/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/platform/detect.go internal/platform/detect_test.go
git commit -m "feat: add platform factory function"
```

---

## Task 8: UI Styles

**Files:**
- Create: `internal/ui/styles.go`

**Step 1: Create styles with lipgloss**

```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("39")  // Blue
	secondaryColor = lipgloss.Color("245") // Gray
	successColor   = lipgloss.Color("42")  // Green
	errorColor     = lipgloss.Color("196") // Red
	selectedColor  = lipgloss.Color("170") // Purple

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Padding(0, 2)

	// List item styles
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(selectedColor).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Status styles
	StatusOpenStyle = lipgloss.NewStyle().
			Foreground(successColor)

	StatusDraftStyle = lipgloss.NewStyle().
				Foreground(secondaryColor)

	StatusMergedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("141")) // Purple

	StatusClosedStyle = lipgloss.NewStyle().
				Foreground(errorColor)

	// Footer style
	FooterStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	// Modal style
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor)
)
```

**Step 2: Verify it compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/ui/styles.go
git commit -m "feat: add UI styles with lipgloss"
```

---

## Task 9: MR List Component

**Files:**
- Create: `internal/ui/mrlist.go`

**Step 1: Create MR list component**

```go
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
```

**Step 2: Verify it compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/ui/mrlist.go
git commit -m "feat: add MR list UI component"
```

---

## Task 10: Checkout Modal Component

**Files:**
- Create: `internal/ui/checkout.go`

**Step 1: Create checkout modal component**

```go
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
```

**Step 2: Verify it compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/ui/checkout.go
git commit -m "feat: add checkout modal UI component"
```

---

## Task 11: Dashboard Component

**Files:**
- Create: `internal/ui/dashboard.go`

**Step 1: Create dashboard component**

```go
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
	platform     platform.Platform
	repoInfo     platform.RepoInfo
	repoPath     string
	author       string
	authorInput  textinput.Model
	editingAuthor bool
	activeTab    Tab
	mrList       MRList
	checkout     *CheckoutModal
	width        int
	height       int
	err          error
	loading      bool
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
```

**Step 2: Verify it compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/ui/dashboard.go
git commit -m "feat: add dashboard UI component"
```

---

## Task 12: Wire Up Main Entry Point

**Files:**
- Modify: `main.go`

**Step 1: Update main.go to use dashboard**

```go
package main

import (
	"fmt"
	"os"

	"gitHelper/internal/git"
	"gitHelper/internal/platform"
	"gitHelper/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Check if we're in a git repo
	if !git.IsGitRepo(cwd) {
		fmt.Fprintln(os.Stderr, "Error: Not a git repository. Run gitHelper from inside a git repo.")
		os.Exit(1)
	}

	// Get remote URL
	remoteURL, err := git.GetRemoteURL(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting remote URL: %v\n", err)
		os.Exit(1)
	}

	// Detect platform
	p, err := platform.NewPlatform(cwd, remoteURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Supported platforms: github.com, gitlab.com")
		os.Exit(1)
	}

	// Create and run the dashboard
	dashboard := ui.NewDashboard(p, cwd)
	prog := tea.NewProgram(dashboard, tea.WithAltScreen())

	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: Verify it compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Test manually**

Run: `go run main.go` (in a GitHub or GitLab repo)
Expected: Dashboard appears, shows MRs, can navigate and checkout

**Step 4: Commit**

```bash
git add main.go
git commit -m "feat: wire up main entry point with dashboard"
```

---

## Task 13: Final Integration Test

**Step 1: Build the binary**

Run: `go build -o githelper .`
Expected: Binary created successfully

**Step 2: Test in a GitHub repo**

```bash
cd /path/to/github/repo
/path/to/githelper
```
Expected: Dashboard shows GitHub PRs, checkout works

**Step 3: Test in a GitLab repo**

```bash
cd /path/to/gitlab/repo
/path/to/githelper
```
Expected: Dashboard shows GitLab MRs, checkout works

**Step 4: Test error cases**

```bash
cd /tmp
/path/to/githelper
```
Expected: "Not a git repository" error

**Step 5: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix: address integration test issues"
```

---

## Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Project setup | main.go, go.mod |
| 2 | Platform types | internal/platform/platform.go |
| 3 | Git operations | internal/git/git.go |
| 4 | Platform detection | internal/platform/detect.go |
| 5 | GitHub implementation | internal/platform/github.go |
| 6 | GitLab implementation | internal/platform/gitlab.go |
| 7 | Platform factory | internal/platform/detect.go |
| 8 | UI styles | internal/ui/styles.go |
| 9 | MR list component | internal/ui/mrlist.go |
| 10 | Checkout modal | internal/ui/checkout.go |
| 11 | Dashboard component | internal/ui/dashboard.go |
| 12 | Main entry point | main.go |
| 13 | Integration testing | - |
