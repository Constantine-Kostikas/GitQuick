package platform

import (
	"errors"
	"strings"
)

// DetectPlatformFromURL returns "github", "gitlab", or "" based on remote URL
func DetectPlatformFromURL(url string) string {
	urlLower := strings.ToLower(url)
	if strings.Contains(urlLower, "github.com") {
		return "github"
	}
	if strings.Contains(urlLower, "gitlab.com") {
		return "gitlab"
	}
	if strings.Contains(urlLower, "generation-y") {
		return "gitlab"
	}
	
	return ""
}

// ErrUnknownPlatform is returned when the remote URL doesn't match known platforms
var ErrUnknownPlatform = errors.New("unknown platform: only github.com and gitlab.com are supported")

// NewPlatform creates the appropriate Platform implementation based on remote URL
func NewPlatform(repoPath, remoteURL string) (Platform, error) {
	switch DetectPlatformFromURL(remoteURL) {
	case "github":
		return NewGitHub(repoPath), nil
	case "gitlab":
		return NewGitLab(repoPath), nil
	default:
		return nil, ErrUnknownPlatform
	}
}
