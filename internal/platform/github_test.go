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
