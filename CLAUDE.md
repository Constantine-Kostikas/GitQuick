# gitHelper Project Context

## Overview

**gitHelper** (`gtui`) is a terminal-based user interface (TUI) utility written in Go that streamlines Git workflow by providing quick access to Merge Requests (MRs) and Pull Requests (PRs) across GitHub and GitLab.

## Key Features

- MR/PR browsing with author filtering (`@me` quick reference)
- One-command branch checkout with automatic fetch/pull
- Platform auto-detection (GitHub vs GitLab)
- Interactive tab-based dashboard
- Real-time checkout progress with step-by-step feedback

## Technology Stack

- **Language**: Go 1.25
- **UI Framework**: Bubbletea + Lipgloss + Bubbles (Charmbracelet ecosystem)
- **External CLIs**: `git`, `gh` (GitHub CLI), `glab` (GitLab CLI)
- **Build/Release**: GoReleaser v2, GitHub Actions

## Project Structure

```
gitHelper/
├── main.go                    # Entry point - validates repo, detects platform, runs TUI
├── go.mod / go.sum           # Go modules
├── .goreleaser.yaml          # Multi-platform build config
├── internal/
│   ├── git/                  # Git operations
│   │   ├── git.go            # IsGitRepo, GetRemoteURL, Checkout (fetch→checkout→pull)
│   │   └── git_test.go
│   ├── platform/             # Platform abstraction
│   │   ├── platform.go       # Interface: Platform, MR, RepoInfo types
│   │   ├── detect.go         # DetectPlatformFromURL, NewPlatform factory
│   │   ├── github.go         # GitHub impl via `gh` CLI
│   │   ├── gitlab.go         # GitLab impl via `glab` CLI
│   │   └── *_test.go
│   └── ui/                   # TUI components
│       ├── dashboard.go      # Main Bubbletea model, tabs, state management
│       ├── mrlist.go         # Scrollable MR list with search
│       ├── checkout.go       # Checkout modal with progress states
│       └── styles.go         # Lipgloss styling definitions
├── docs/plans/               # Design documents
└── .github/workflows/        # CI (go.yml) and release (release.yml)
```

## Key Patterns

### Platform Abstraction
- `Platform` interface with `ListMRs(author)` and `GetRepoInfo()`
- `MR` struct: Number, Title, Branch, Status (open/draft/merged/closed), URL
- Factory function `NewPlatform()` returns GitHub or GitLab implementation

### UI Architecture (Bubbletea)
- Dashboard is the root model orchestrating all UI
- Three tabs: MRs (active), Issues (placeholder), Branches (placeholder)
- CheckoutModal overlays dashboard with progress states: Fetching → CheckingOut → Pulling → Done/Error

### Key Bindings
- `q`/`Ctrl+C`: Quit
- `↑↓`: Navigate list
- `Enter`: Checkout selected MR branch
- `a`: Edit author filter
- `Tab`: Switch tabs

## Build & Test

```bash
# Run tests
go test ./...

# Build locally
go build -o dist/gtui .

# Release (requires GPG key)
goreleaser release --clean
```

## Dependencies

**Direct**:
- `github.com/charmbracelet/bubbles` v0.21.0
- `github.com/charmbracelet/bubbletea` v1.3.10
- `github.com/charmbracelet/lipgloss` v1.1.0

**Runtime Requirements**:
- `git` - for repo operations
- `gh` - for GitHub repos (must be authenticated)
- `glab` - for GitLab repos (must be authenticated)

## Recent Work

The project is feature-complete and release-ready. Recent commits focused on:
- GoReleaser v2 configuration fixes
- GPG signing setup for releases
- Cleanup of unused files

## Future Considerations

Per design docs, potential enhancements include:
- Self-hosted GitLab/GitHub Enterprise support via config file
- Issues and Branches tab implementations
- CI status display in MR list
- Caching for faster startup
