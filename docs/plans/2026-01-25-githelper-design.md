# gitHelper TUI Design

## Overview

A terminal UI helper for Git workflows, built with bubbletea. Primary use case: quickly find MRs, select one, and checkout the branch.

## Requirements

- Works with both GitHub (via `gh` CLI) and GitLab (via `glab` CLI)
- Auto-detects platform from current repo's remote URL
- Scoped to current repository only
- Dashboard view with drill-down capability

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     gitHelper TUI                       │
├─────────────────────────────────────────────────────────┤
│  main.go          → Entry point, initializes bubbletea  │
│  internal/                                              │
│    platform/      → GitHub/GitLab abstraction layer     │
│      platform.go  → Interface + detection logic         │
│      github.go    → gh CLI wrapper                      │
│      gitlab.go    → glab CLI wrapper                    │
│    ui/            → Bubbletea components                │
│      dashboard.go → Main dashboard view                 │
│      mrlist.go    → MR list component                   │
│      styles.go    → Lipgloss styling                    │
│    git/           → Local git operations                │
│      checkout.go  → Fetch + checkout + pull logic       │
└─────────────────────────────────────────────────────────┘
```

## Platform Abstraction

```go
// internal/platform/platform.go

type MR struct {
    Number    int
    Title     string
    Branch    string
    Status    string  // "open", "draft", "merged", "closed"
    URL       string
}

type RepoInfo struct {
    Name          string
    Description   string
    Platform      string  // "github" or "gitlab"
    DefaultBranch string
}

type Platform interface {
    ListMRs(author string) ([]MR, error)  // "@me" or username
    GetRepoInfo() (RepoInfo, error)
}

func Detect() (Platform, error)
```

### Detection Logic

1. Run `git remote get-url origin`
2. Parse URL - `github.com` → GitHub, `gitlab.com` → GitLab
3. Self-hosted instances: future config file support (`~/.githelper.yaml`)

### CLI Commands

GitHub:
```bash
gh pr list --author @me --json number,title,headRefName,state
```

GitLab:
```bash
glab mr list --author @me -F json
```

## UI Design

### Dashboard Layout

```
┌─ gitHelper ─────────────────────────────────────────────┐
│  repo: myorg/myproject          platform: github        │
├─────────────────────────────────────────────────────────┤
│  Author: [@me ▼]                                        │
├─────────────────────────────────────────────────────────┤
│  My MRs (3)           Issues (12)           Branches    │
├─────────────────────────────────────────────────────────┤
│  ● #142 Fix login timeout              feature/login    │
│  ○ #138 Add user preferences           user-prefs       │
│  ○ #135 Update dependencies            chore/deps       │
│                                                         │
├─────────────────────────────────────────────────────────┤
│  ↑↓ navigate  │  enter checkout  │  a author  │  q quit │
└─────────────────────────────────────────────────────────┘
```

- Header: repo name, detected platform
- Author selector: dropdown/input to filter by author
- Navigation tabs: MRs, Issues, Branches (Issues/Branches are placeholders)
- Main area: scrollable list with selection highlight
- Footer: keybinding hints

### MR Information (Minimal)

- Title
- Branch name
- Status (open/draft/merged/closed)

Can iterate to add more fields later (approvals, CI status, reviewers).

## Checkout Flow

When user presses `enter` on an MR:

```
┌─ Checkout ─────────────────────────────────────────┐
│                                                    │
│  #142 Fix login timeout                            │
│  Branch: feature/login                             │
│                                                    │
│  ⠋ Fetching origin...                              │
│  ✓ Fetched                                         │
│  ⠋ Checking out feature/login...                   │
│  ✓ Checked out                                     │
│  ⠋ Pulling latest changes...                       │
│  ✓ Pulled (3 commits)                              │
│                                                    │
│                               [cancel: esc]        │
└────────────────────────────────────────────────────┘
```

Steps:
1. `git fetch origin`
2. `git checkout <branch>`
3. `git pull`
4. Return to dashboard with success message

## Error Handling

| Error | Behavior |
|-------|----------|
| Uncommitted changes | Show error, return to dashboard |
| CLI not installed | Detect at startup, show install instructions, exit |
| Not authenticated | Show CLI's auth error, exit or return to dashboard |
| Not a git repo | Detect at startup, show message, exit |

Keep it simple - show the message, let user fix externally. No auto-stashing.

## Dependencies

```go
require (
    github.com/charmbracelet/bubbletea
    github.com/charmbracelet/lipgloss
    github.com/charmbracelet/bubbles
)
```

External tools (shelled out):
- `git` (assumed installed)
- `gh` (for GitHub repos)
- `glab` (for GitLab repos)

## Future Considerations

- Self-hosted GitLab/GitHub Enterprise domain mapping
- Issue management tab
- Branch management tab
- Configurable keybindings
- Caching for faster startup
