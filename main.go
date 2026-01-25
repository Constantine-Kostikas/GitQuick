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
