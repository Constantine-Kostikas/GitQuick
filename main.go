package main

import (
	"fmt"
	"os"

	"github.com/Constantine-Kostikas/GitQuick/internal/boot"
	"github.com/Constantine-Kostikas/GitQuick/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	system := boot.Bootstrap()

	if system.Errors != nil {
		for _, msg := range system.Errors {
			fmt.Println(msg)
		}

		os.Exit(1)
	}

	// Create and run the dashboard
	dashboard := ui.NewDashboard(system.Platform, system.WorkingDir)
	prog := tea.NewProgram(dashboard, tea.WithAltScreen())

	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
