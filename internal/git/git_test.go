package git_test

import (
	"testing"

	"github.com/hickstein/hcflow/internal/git"
)

func TestParseGitHubRemote(t *testing.T) {
	tests := []struct {
		url      string
		owner    string
		repo     string
		wantErr  bool
	}{
		{"https://github.com/owner/repo.git", "owner", "repo", false},
		{"https://github.com/owner/repo", "owner", "repo", false},
		{"git@github.com:owner/repo.git", "owner", "repo", false},
		{"git@github.com:owner/repo", "owner", "repo", false},
		{"https://gitlab.com/owner/repo", "", "", true},
		{"", "", "", true},
	}

	for _, tt := range tests {
		owner, repo, err := git.ParseGitHubRemote(tt.url)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseGitHubRemote(%q): expected error", tt.url)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseGitHubRemote(%q): unexpected error: %v", tt.url, err)
			continue
		}
		if owner != tt.owner || repo != tt.repo {
			t.Errorf("ParseGitHubRemote(%q): got %s/%s, want %s/%s",
				tt.url, owner, repo, tt.owner, tt.repo)
		}
	}
}
