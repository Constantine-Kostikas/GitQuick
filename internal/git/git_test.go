package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGitRepo(t *testing.T) {
	// Create temp dir without .git
	tmpDir, err := os.MkdirTemp("", "githelper-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Error(err)
		}
	}(tmpDir)

	if IsGitRepo(tmpDir) {
		t.Error("expected false for non-git directory")
	}

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	if !IsGitRepo(tmpDir) {
		t.Error("expected true for git directory")
	}
}

func TestGetRemoteURL(t *testing.T) {
	// This test requires being in an actual git repo
	// We'll test error case for non-git dir
	tmpDir, err := os.MkdirTemp("", "githelper-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Error(err)
		}
	}(tmpDir)

	_, err = GetRemoteURL(tmpDir)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestCheckout_NotGitRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "githelper-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Error(err)
		}
	}(tmpDir)

	err = Checkout(tmpDir, "main")
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}
