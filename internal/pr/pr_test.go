package pr_test

import (
	"testing"

	"github.com/hickstein/hcflow/internal/pr"
)

func TestIsConventionalCommit(t *testing.T) {
	valid := []string{
		"feat: add login",
		"fix(auth): resolve token expiry",
		"feat(api)!: remove legacy endpoint",
		"chore: update dependencies",
		"docs(readme): improve examples",
		"perf(db): optimise queries",
		"refactor(core): extract service",
		"test: add integration tests",
		"build(ci): update actions",
		"ci: add linting step",
	}
	for _, s := range valid {
		if !pr.IsConventionalCommit(s) {
			t.Errorf("expected %q to be valid conventional commit", s)
		}
	}

	invalid := []string{
		"add login",
		"fix auth resolve token",
		"WIP: some work",
		"initial commit",
		"",
		"feat",
		"feat:",
	}
	for _, s := range invalid {
		if pr.IsConventionalCommit(s) {
			t.Errorf("expected %q to be invalid conventional commit", s)
		}
	}
}

func TestBranchName(t *testing.T) {
	tests := []struct {
		ccType string
		desc   string
		want   string
	}{
		{"feat", "telegram callback", "feat/telegram-callback"},
		{"fix", "reconnect timeout", "fix/reconnect-timeout"},
		{"chore", "Update Deps!", "chore/update-deps"},
	}
	for _, tt := range tests {
		got := pr.BranchName(tt.ccType, tt.desc)
		if got != tt.want {
			t.Errorf("BranchName(%q, %q) = %q, want %q", tt.ccType, tt.desc, got, tt.want)
		}
	}
}

func TestSuggestTitle(t *testing.T) {
	title := pr.SuggestTitle("feat/telegram-callback", []string{"initial implementation", "fix tests"})
	if !pr.IsConventionalCommit(title) {
		t.Errorf("SuggestTitle should produce a conventional commit, got %q", title)
	}
}

func TestValidatePRTitle(t *testing.T) {
	if err := pr.ValidatePRTitle("feat(auth): add login"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := pr.ValidatePRTitle("add login"); err == nil {
		t.Error("expected error for non-conventional title")
	}
}
