**Note:** This project was vibe coded for learning purposes. It's a playground for exploring Go, TUI development with Bubbletea, and CLI tool design. Contributions and feedback welcome, but don't expect enterprise-grade stability.
 

# gitQuick (gq)

A terminal-based user interface for streamlining Git workflow across GitHub and GitLab.

## What is this?

gitQuick (`gq`) is a TUI utility that gives you quick access to Merge Requests and Pull Requests without leaving your terminal. Instead of context-switching to a browser, you can browse, inspect, and checkout MR/PR branches directly from the command line.

**The problem it solves:** You're deep in terminal work and need to review or checkout a colleague's PR. Normally you'd open a browser, navigate to the repo, find the PR, copy the branch name, go back to terminal, fetch, checkout. With gitQuick, you press a few keys and you're done.

## Features

- **MR/PR browsing** - View open merge requests with status indicators (open, draft, merged, closed)
- **Detail view** - See PR description and file changes with additions/deletions per file
- **Quick checkout** - Select an MR and checkout its branch with automatic fetch/pull
- **Author filtering** - Filter by author with `@me` shortcut for your own PRs
- **Platform auto-detection** - Works with both GitHub and GitLab, detected from your remote URL
- **Keyboard-driven** - Navigate entirely with keyboard shortcuts

## Installation

```bash
# Using go install
go install github.com/Constantine-Kostikas/GitQuick@latest

# Or download from releases
# https://github.com/Constantine-Kostikas/gitQuick/releases
```

### macOS

1. Download the correct archive for your Mac from the [releases page](https://github.com/Constantine-Kostikas/gitQuick/releases):
   - Apple Silicon (M1/M2/M3): `GitQuick_Darwin_arm64.tar.gz`
   - Intel: `GitQuick_Darwin_x86_64.tar.gz`

2. Extract and move the binary to your PATH:

```bash
tar -xzf GitQuick_Darwin_arm64.tar.gz
mv gq /usr/local/bin/gq
```

3. Remove the macOS quarantine attribute (one-time step — required because the binary is not Apple-notarized):

```bash
xattr -d com.apple.quarantine /usr/local/bin/gq
```

Without step 3, macOS Gatekeeper will block the binary or prompt you to allow it on every run.

> **Note:** Steps 2 and 3 are not needed if you install via `go install`, since the binary is compiled locally on your machine.

### Requirements

- `git` - for repository operations
- `gh` - GitHub CLI (for GitHub repos)
- `glab` - GitLab CLI (for GitLab repos)

**Important:** gitQuick is a wrapper around these CLIs and does not handle authentication. You must authenticate with the respective CLI before using gitQuick:

```bash
# For GitHub repositories
gh auth login

# For GitLab repositories
glab auth login
```

## Usage

Run `gq` from any git repository:

```bash
cd your-repo
gq
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j/k` or `↑/↓` | Navigate list |
| `Enter` | View MR details |
| `Enter` (in detail view) | Checkout branch |
| `Esc` | Close modal / cancel |
| `a` | Open author picker |
| `r` | Refresh MR list |
| `Tab` | Switch tabs |
| `q` | Quit |

## How it works

1. Detects whether you're in a GitHub or GitLab repo from the remote URL
2. Uses `gh` or `glab` CLI to fetch MR/PR data
3. Displays an interactive list you can browse and filter
4. When you select an MR, shows details including changed files with diff stats
5. On checkout, runs `git fetch`, `git checkout`, and `git pull` automatically

## License

MIT License

Copyright (c) 2026

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

---