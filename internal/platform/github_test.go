package platform

import (
	"reflect"
	"testing"
)

func TestGitHub_ParseMRList(t *testing.T) {
	jsonOutput := `[
		{"number": 142, "title": "Fix login timeout", "headRefName": "feature/login", "state": "OPEN", "url": "https://github.com/org/repo/pull/142"},
		{"number": 138, "title": "Add user preferences", "headRefName": "user-prefs", "state": "DRAFT", "url": "https://github.com/org/repo/pull/138"}
	]`

	mrs, err := parseGitHubMRs([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []MR{
		{Number: 142, Title: "Fix login timeout", Branch: "feature/login", Status: "open", URL: "https://github.com/org/repo/pull/142"},
		{Number: 138, Title: "Add user preferences", Branch: "user-prefs", Status: "draft", URL: "https://github.com/org/repo/pull/138"},
	}

	if !reflect.DeepEqual(mrs, expected) {
		t.Errorf("got %+v, want %+v", mrs, expected)
	}
}

func TestGitHub_ParseMRDetail(t *testing.T) {
	jsonOutput := `{
		"title": "Add feature X",
		"body": "This PR adds feature X.\n\n## Changes\n- Added new module",
		"additions": 150,
		"deletions": 23,
		"files": [
			{"path": "internal/feature/feature.go", "additions": 100, "deletions": 10},
			{"path": "internal/feature/feature_test.go", "additions": 50, "deletions": 13}
		]
	}`

	detail, err := parseGitHubMRDetail(42, []byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := MRDetail{
		Number:    42,
		Title:     "Add feature X",
		Body:      "This PR adds feature X.\n\n## Changes\n- Added new module",
		Additions: 150,
		Deletions: 23,
		Files: []FileChange{
			{Path: "internal/feature/feature.go", Additions: 100, Deletions: 10},
			{Path: "internal/feature/feature_test.go", Additions: 50, Deletions: 13},
		},
	}

	if !reflect.DeepEqual(detail, expected) {
		t.Errorf("got %+v, want %+v", detail, expected)
	}
}

func TestGitHub_ParseMRDetail_EmptyFiles(t *testing.T) {
	jsonOutput := `{
		"title": "Update docs",
		"body": "Documentation updates",
		"additions": 0,
		"deletions": 0,
		"files": []
	}`

	detail, err := parseGitHubMRDetail(99, []byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if detail.Number != 99 {
		t.Errorf("Number: got %d, want 99", detail.Number)
	}
	if detail.Title != "Update docs" {
		t.Errorf("Title: got %q, want %q", detail.Title, "Update docs")
	}
	if len(detail.Files) != 0 {
		t.Errorf("Files: got %d files, want 0", len(detail.Files))
	}
}
