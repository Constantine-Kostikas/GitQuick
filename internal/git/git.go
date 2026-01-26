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

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	out, err := cmd.Output()
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
