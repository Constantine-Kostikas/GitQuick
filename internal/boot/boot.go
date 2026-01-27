package boot

import (
	"fmt"
	"gitHelper/internal/git"
	"gitHelper/internal/platform"
	"os"
)

type System struct {
	WorkingDir string
	isGitRepo  bool
	remoteUrl  string
	Platform   platform.Platform
	Errors     []error
}

func Bootstrap() System {
	system := System{}

	workingDir, err := os.Getwd()
	if err != nil {
		system.Errors = append(system.Errors, fmt.Errorf("error getting current directory: %v\n", err))
	}

	isGitRepo := git.IsGitRepo(workingDir)
	if !isGitRepo {
		system.Errors = append(system.Errors, fmt.Errorf("error: Not a git repository. Run gitHelper from inside a git repo"))
	}

	remoteURL, err := git.GetRemoteURL(workingDir)
	if err != nil {
		system.Errors = append(system.Errors, fmt.Errorf("error getting remote URL: %v", err))
	}

	gitPlatform, err := platform.NewPlatform(workingDir, remoteURL)
	if err != nil {
		system.Errors = append(system.Errors, fmt.Errorf("error getting platform: %v", err))
	}

	system.WorkingDir = workingDir
	system.isGitRepo = isGitRepo
	system.remoteUrl = remoteURL
	system.Platform = gitPlatform

	return system
}
