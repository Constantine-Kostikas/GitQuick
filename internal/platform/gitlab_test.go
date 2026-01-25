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
