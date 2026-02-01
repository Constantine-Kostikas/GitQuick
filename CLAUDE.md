# gitQuick Project Context
**FOR CLAUDE** The project is called `gitQuick` regardless of the repo name

## Overview

**GitQuick** (`gq`) is a terminal-based user interface (TUI) utility written in Go that streamlines Git workflow by providing quick access to Merge Requests (MRs) and Pull Requests (PRs) across GitHub and GitLab.

**Current Version:** 0.1.3

## Key Features

- MR/PR browsing with author filtering (`@me` quick reference)
- One-command branch checkout with automatic fetch/pull
- Platform auto-detection (GitHub vs GitLab)
- Interactive tab-based dashboard
- Real-time checkout progress with step-by-step feedback
- Author picker showing repository contributors
- Refresh functionality for MR lists

## Technology Stack

- **Language**: Go 1.25
- **UI Framework**: Bubbletea + Lipgloss + Bubbles (Charmbracelet ecosystem)
- **External CLIs**: `git`, `gh` (GitHub CLI), `glab` (GitLab CLI)
- **Build/Release**: GoReleaser v2, GitHub Actions
- **Signing**: GPG-signed releases

## Project Structure

```
gitHelper/
├── main.go                           # Entry point (validates repo, detects platform, runs TUI)
├── go.mod / go.sum                   # Go modules (Go 1.25)
├── .goreleaser.yaml                  # Multi-platform release config (v2, GPG signed)
├── CHANGELOG.md                      # Version history
├── CLAUDE.md                         # This file
├── .gitignore
│
├── .github/workflows/
│   ├── go.yml                        # CI workflow (build & test on push/PR)
│   └── release.yml                   # GoReleaser CD workflow (on tag push)
│
├── docs/plans/
│   ├── 2026-01-25-githelper-design.md        # Architecture & design docs
│   └── 2026-01-25-githelper-implementation.md # Step-by-step implementation guide
│
├── internal/
│   ├── cmd/
│   │   └── run.go                    # Command execution wrapper with 30s timeout
│   │
│   ├── git/                          # Git operations package
│   │   ├── git.go                    # IsGitRepo, GetRemoteURL, GetCurrentBranch, Checkout
│   │   └── git_test.go               # Unit tests
│   │
│   ├── platform/                     # Platform abstraction layer
│   │   ├── platform.go               # Platform interface, MR, Author, RepoInfo types
│   │   ├── detect.go                 # DetectPlatformFromURL, NewPlatform factory
│   │   ├── detect_test.go
│   │   ├── github.go                 # GitHub impl via `gh` CLI
│   │   ├── github_test.go
│   │   ├── gitlab.go                 # GitLab impl via `glab` CLI
│   │   └── gitlab_test.go
│   │
│   └── ui/                           # Bubbletea TUI components
│       ├── styles.go                 # Lipgloss styling definitions
│       ├── dashboard.go              # Main model, tabs, state management
│       ├── mrlist.go                 # Scrollable MR list component
│       ├── checkout.go               # Checkout modal with progress states
│       └── authorpicker.go           # Author selection modal
│
└── dist/                             # Build output directory
```

## Key Types and Interfaces

### Platform Package (`internal/platform/`)

```go
type Platform interface {
    ListMRs(author string) ([]MR, error)      // Get PRs/MRs by author (@me supported)
    GetRepoInfo() (RepoInfo, error)           // Get repo metadata
    ListAuthors() ([]Author, error)           // List repository contributors
}

type MR struct {
    Number int      // PR/MR number
    Title  string   // Title
    Branch string   // Source branch
    Status string   // "open", "draft", "merged", "closed"
    URL    string   // Link to PR/MR
}

type Author struct {
    Username string // GitHub/GitLab username
    Name     string // Display name
}

type RepoInfo struct {
    Name          string // Repository name
    Description   string // Repo description
    Platform      string // "github" or "gitlab"
    DefaultBranch string // Default branch name
}
```

### Git Package (`internal/git/`)

```go
type CheckoutError struct {
    Step string // "fetch", "checkout", or "pull"
    Err  error  // Underlying error
}

func IsGitRepo(path string) bool
func GetRemoteURL(path string) (string, error)
func GetCurrentBranch(path string) (string, error)
func Checkout(path, branch string) error  // fetch → checkout → pull
```

### UI Components (`internal/ui/`)

- **Dashboard**: Root model orchestrating all UI, manages tabs and state
- **MRList**: Scrollable list with fuzzy filtering
- **CheckoutModal**: Progress overlay (Fetching → CheckingOut → Pulling → Done/Error)
- **AuthorPicker**: Modal for selecting author filter from contributors

## Application Flow

1. Get current working directory
2. Validate it's a git repository (`git.IsGitRepo()`)
3. Retrieve remote URL (`git.GetRemoteURL()`)
4. Auto-detect platform from URL (`platform.DetectPlatformFromURL()`)
5. Create appropriate platform instance (`platform.NewPlatform()`)
6. Initialize Bubbletea TUI with Dashboard
7. Async load: repo info, current branch, authors, MR list
8. Run event loop with alternative screen mode

## Keyboard Bindings

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit application |
| `Tab` | Switch between MRs/Issues/Branches tabs |
| `↑↓` / `j/k` | Navigate list |
| `Enter` | Checkout selected MR branch |
| `w` | Open selected MR in browser |
| `f` / `/` | Search/filter MR list |
| `a` | Open author picker modal |
| `r` / `R` | Refresh MR list |
| `m` | Checkout to default/main branch |
| `Esc` | Cancel/close modal or search |

### MR Detail View
| Key | Action |
|-----|--------|
| `j/k` | Scroll file list |
| `d` | View full description (scrollable) |
| `c` | View commits (scrollable) |
| `Enter` | Checkout branch |
| `Esc` | Close modal |

## External CLI Commands

**GitHub (via `gh`):**
```bash
gh pr list --author <author> --json number,title,headRefName,state,url
gh repo view --json name,description,defaultBranchRef
gh api repos/{owner}/{repo}/contributors --paginate -q .[].login
```

**GitLab (via `glab`):**
```bash
glab mr list -F json --author <author>
glab repo view -F json
glab api projects/:id/members/all
```

**Git:**
```bash
git remote get-url origin
git rev-parse --abbrev-ref HEAD
git fetch origin
git checkout <branch>
git pull
```

## Build & Test

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Build locally
go build -o dist/gq .

# Run locally
./dist/gq

# Release (requires GPG key and GITHUB_TOKEN)
goreleaser release --clean
```

## Dependencies

**Direct (go.mod):**
- `github.com/charmbracelet/bubbles` v0.21.0 - List, spinner, textinput
- `github.com/charmbracelet/bubbletea` v1.3.10 - TUI framework
- `github.com/charmbracelet/lipgloss` v1.1.0 - Styling/layout

**Runtime Requirements:**
- `git` - for repo operations
- `gh` - for GitHub repos (must be authenticated via `gh auth login`)
- `glab` - for GitLab repos (must be authenticated via `glab auth login`)

## CI/CD

**Build Workflow (`.github/workflows/go.yml`):**
- Triggers: Push to main, PRs to main
- Go 1.25 on Ubuntu latest
- Steps: checkout, setup Go, build, test

**Release Workflow (`.github/workflows/release.yml`):**
- Triggers: Tag push matching `v*`
- GoReleaser v2 with GPG signing
- Builds for: Linux, Darwin (macOS), Windows × amd64/arm64
- Requires secrets: `GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`, `GPG_FINGERPRINT`

## Testing

9 unit tests across packages (all passing):
- `internal/git/` - IsGitRepo, GetRemoteURL, Checkout error handling
- `internal/platform/` - URL detection, platform factory, JSON parsing

No UI tests (Bubbletea testing is complex).

## Version History

- **v0.1.3** - GoReleaser v2 configuration fixes
- **v0.1.2** - GoReleaser build configuration fix
- **v0.1.1** - Added GPG signing for releases
- **v0.1.0** - Initial release with full feature set

## Key Patterns

### Platform Abstraction
- Interface-based design allows adding new platforms
- Factory pattern via `NewPlatform()`
- CLI wrapping with JSON output parsing
- 30-second timeout on all commands

### Bubbletea Architecture
- Model-Update-View pattern
- Custom message types for async operations
- `tea.Batch()` for parallel loading
- Modal overlays on main dashboard

### Styling
- Lipgloss predefined styles in `styles.go`
- Night-friendly color palette (cyan, teal, green)
- Consistent active/inactive states

## Future Considerations

Per design docs, potential enhancements:
- Self-hosted GitLab/GitHub Enterprise support via config file
- Issues and Branches tab implementations
- CI status display in MR list
- Caching for faster startup
- Configurable key bindings