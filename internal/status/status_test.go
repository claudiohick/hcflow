package status

import (
	"testing"
)

func TestExtractVersionFromTitle(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"chore(main): release 1.5.0", "v1.5.0"},
		{"Release v2.0.0", "v2.0.0"},
		{"chore: release 0.1.2", "v0.1.2"},
		{"no version here", ""},
	}
	for _, tt := range tests {
		got := extractVersionFromTitle(tt.title)
		if got != tt.want {
			t.Errorf("extractVersionFromTitle(%q) = %q, want %q", tt.title, got, tt.want)
		}
	}
}

func TestLooksLikeVersion(t *testing.T) {
	valid := []string{"1.0.0", "1.2.3", "10.20.30", "0.1.2"}
	for _, s := range valid {
		if !looksLikeVersion(s) {
			t.Errorf("expected %q to be a valid version", s)
		}
	}
	invalid := []string{"1", "1.x.0", "abc", "", "v1.0.0"}
	for _, s := range invalid {
		if looksLikeVersion(s) {
			t.Errorf("expected %q to NOT be a valid version", s)
		}
	}
}
