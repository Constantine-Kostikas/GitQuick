package platform

import (
	"reflect"
	"testing"
)

func TestGitLab_ParseMRList(t *testing.T) {
	jsonOutput := `[
		{"iid": 42, "title": "Update API docs", "source_branch": "docs/api", "state": "opened", "web_url": "https://gitlab.com/org/repo/-/merge_requests/42"},
		{"iid": 40, "title": "Fix CI pipeline", "source_branch": "fix/ci", "state": "merged", "web_url": "https://gitlab.com/org/repo/-/merge_requests/40"}
	]`

	mrs, err := parseGitLabMRs([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []MR{
		{Number: 42, Title: "Update API docs", Branch: "docs/api", Status: "open", URL: "https://gitlab.com/org/repo/-/merge_requests/42"},
		{Number: 40, Title: "Fix CI pipeline", Branch: "fix/ci", Status: "merged", URL: "https://gitlab.com/org/repo/-/merge_requests/40"},
	}

	if !reflect.DeepEqual(mrs, expected) {
		t.Errorf("got %+v, want %+v", mrs, expected)
	}
}

func TestGitLab_ParseMRDetail(t *testing.T) {
	jsonOutput := `{
		"iid": 123,
		"title": "Add new feature",
		"description": "This MR adds a new feature.\n\n## Changes\n- Added foo\n- Updated bar"
	}`

	detail, err := parseGitLabMRDetail([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if detail.IID != 123 {
		t.Errorf("got IID %d, want 123", detail.IID)
	}
	if detail.Title != "Add new feature" {
		t.Errorf("got Title %q, want %q", detail.Title, "Add new feature")
	}
	expectedDesc := "This MR adds a new feature.\n\n## Changes\n- Added foo\n- Updated bar"
	if detail.Description != expectedDesc {
		t.Errorf("got Description %q, want %q", detail.Description, expectedDesc)
	}
}

func TestGitLab_ParseMRDetail_EmptyDescription(t *testing.T) {
	jsonOutput := `{
		"iid": 456,
		"title": "Quick fix",
		"description": ""
	}`

	detail, err := parseGitLabMRDetail([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if detail.IID != 456 {
		t.Errorf("got IID %d, want 456", detail.IID)
	}
	if detail.Description != "" {
		t.Errorf("got Description %q, want empty string", detail.Description)
	}
}

func TestGitLab_ParseDiffStats(t *testing.T) {
	jsonOutput := `[
		{"old_path": "README.md", "new_path": "README.md", "additions": 10, "deletions": 2},
		{"old_path": "main.go", "new_path": "main.go", "additions": 50, "deletions": 30},
		{"old_path": "old_name.go", "new_path": "new_name.go", "additions": 5, "deletions": 0}
	]`

	files, additions, deletions, err := parseGitLabDiffStats([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("got %d files, want 3", len(files))
	}

	expected := []FileChange{
		{Path: "README.md", Additions: 10, Deletions: 2},
		{Path: "main.go", Additions: 50, Deletions: 30},
		{Path: "new_name.go", Additions: 5, Deletions: 0},
	}

	if !reflect.DeepEqual(files, expected) {
		t.Errorf("got files %+v, want %+v", files, expected)
	}

	if additions != 65 {
		t.Errorf("got total additions %d, want 65", additions)
	}
	if deletions != 32 {
		t.Errorf("got total deletions %d, want 32", deletions)
	}
}

func TestGitLab_ParseDiffStats_EmptyOldPath(t *testing.T) {
	// Test case where old_path is empty (new file)
	jsonOutput := `[
		{"old_path": "", "new_path": "brand_new.go", "additions": 100, "deletions": 0}
	]`

	files, additions, deletions, err := parseGitLabDiffStats([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}

	if files[0].Path != "brand_new.go" {
		t.Errorf("got path %q, want %q", files[0].Path, "brand_new.go")
	}
	if additions != 100 {
		t.Errorf("got additions %d, want 100", additions)
	}
	if deletions != 0 {
		t.Errorf("got deletions %d, want 0", deletions)
	}
}

func TestGitLab_ParseDiffStats_EmptyArray(t *testing.T) {
	jsonOutput := `[]`

	files, additions, deletions, err := parseGitLabDiffStats([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("got %d files, want 0", len(files))
	}
	if additions != 0 {
		t.Errorf("got additions %d, want 0", additions)
	}
	if deletions != 0 {
		t.Errorf("got deletions %d, want 0", deletions)
	}
}
