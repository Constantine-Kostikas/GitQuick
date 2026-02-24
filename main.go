package main

import (
	"fmt"
	"os"

	"github.com/Constantine-Kostikas/GitQuick/internal/boot"
	"github.com/Constantine-Kostikas/GitQuick/internal/ui"

	tea "charm.land/bubbletea/v2"
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
	prog := tea.NewProgram(dashboard)

	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
