package ui

import "testing"

func TestExtractJiraTicket(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"feat: add login #JUM-271", "JUM-271"},
		{"#ABC-1 fix null pointer", "ABC-1"},
		{"no ticket here", ""},
		{"", ""},
		{"#JUM-271 and #JUM-272 both present", "JUM-271"}, // first match
		{"lowercase #jum-271 ignored", ""},                // case-sensitive
	}

	for _, tc := range tests {
		got := extractJiraTicket(tc.title)
		if got != tc.expected {
			t.Errorf("extractJiraTicket(%q) = %q, want %q", tc.title, got, tc.expected)
		}
	}
}
