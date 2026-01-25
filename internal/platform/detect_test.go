package platform

import (
	"reflect"
	"testing"
)

func TestDetectPlatformFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/user/repo.git", "github"},
		{"git@github.com:user/repo.git", "github"},
		{"https://gitlab.com/user/repo.git", "gitlab"},
		{"git@gitlab.com:user/repo.git", "gitlab"},
		{"https://example.com/user/repo.git", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := DetectPlatformFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewPlatform(t *testing.T) {
	// Test that NewPlatform returns correct type based on URL
	tests := []struct {
		url          string
		expectedType string
	}{
		{"https://github.com/user/repo.git", "*platform.GitHub"},
		{"git@gitlab.com:user/repo.git", "*platform.GitLab"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			p, err := NewPlatform("/tmp", tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			typeName := reflect.TypeOf(p).String()
			if typeName != tt.expectedType {
				t.Errorf("got %s, want %s", typeName, tt.expectedType)
			}
		})
	}
}

func TestNewPlatform_Unknown(t *testing.T) {
	_, err := NewPlatform("/tmp", "https://example.com/repo.git")
	if err == nil {
		t.Error("expected error for unknown platform")
	}
}
