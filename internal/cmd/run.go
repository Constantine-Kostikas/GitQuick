package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DefaultTimeout is the default timeout for commands
const DefaultTimeout = 30 * time.Second

// Run executes a command with timeout and captures stderr on failure
func Run(dir string, name string, args ...string) ([]byte, error) {
	return RunWithTimeout(dir, DefaultTimeout, name, args...)
}

// RunWithTimeout executes a command with a specified timeout and captures stderr on failure
func RunWithTimeout(dir string, timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("command timed out after %v", timeout)
		}
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("%s: %s", err.Error(), stderrStr)
		}
		return nil, err
	}

	return stdout.Bytes(), nil
}

// RunSimple executes a command without capturing output, with timeout
func RunSimple(dir string, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command timed out after %v", DefaultTimeout)
		}
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return fmt.Errorf("%s: %s", err.Error(), stderrStr)
		}
		return err
	}

	return nil
}
