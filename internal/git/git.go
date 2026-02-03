package git

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Constantine-Kostikas/GitQuick/internal/cmd"
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

// GetRemoteURL returns the origin remote URL for the repo at path
func GetRemoteURL(path string) (string, error) {
	out, err := cmd.Run(path, "git", "remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(path string) (string, error) {
	out, err := cmd.Run(path, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
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

// IsDirty checks if the working tree has uncommitted changes
func IsDirty(path string) (bool, error) {
	out, err := cmd.Run(path, "git", "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// Checkout fetches, checks out the branch, and pulls
func Checkout(path, branch string) error {
	// Fetch
	if err := cmd.RunSimple(path, "git", "fetch", "origin"); err != nil {
		return &CheckoutError{Step: "fetch", Err: err}
	}

	// Checkout
	if err := cmd.RunSimple(path, "git", "checkout", branch); err != nil {
		return &CheckoutError{Step: "checkout", Err: err}
	}

	// Pull
	if err := cmd.RunSimple(path, "git", "pull"); err != nil {
		return &CheckoutError{Step: "pull", Err: err}
	}

	return nil
}
