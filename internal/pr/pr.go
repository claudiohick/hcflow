// Package pr contains PR-related logic including Conventional Commit validation.
package pr

import (
	"fmt"
	"regexp"
	"strings"
)

// Valid Conventional Commit types accepted by hcflow.
var validTypes = []string{
	"feat", "fix", "perf", "refactor", "docs",
	"test", "build", "ci", "chore",
}

// ccPattern is the regex for a conventional commit title.
var ccPattern = regexp.MustCompile(
	`^(feat|fix|perf|refactor|docs|test|build|ci|chore)(\([^)]+\))?(!)?: .+`,
)

// IsConventionalCommit reports whether title follows Conventional Commits.
func IsConventionalCommit(title string) bool {
	return ccPattern.MatchString(title)
}

// SuggestTitle builds a suggested PR title from branch name and commit messages.
// It tries to detect a conventional-commit-style type from the branch name.
func SuggestTitle(branch string, commits []string) string {
	// Strip branch prefix like feat/some-feature -> feat(some-feature):
	ccType, scope := parseBranchName(branch)
	if ccType == "" {
		ccType = "chore"
	}

	// Build description from last commit or scope
	desc := scope
	if desc == "" && len(commits) > 0 {
		desc = commits[len(commits)-1]
	}
	if desc == "" {
		desc = "..."
	}

	if scope != "" {
		return fmt.Sprintf("%s(%s): %s", ccType, scope, humanize(desc))
	}
	return fmt.Sprintf("%s: %s", ccType, humanize(desc))
}

// parseBranchName splits a branch like "feat/telegram-callback" into type and scope.
func parseBranchName(branch string) (ccType, scope string) {
	for _, t := range validTypes {
		prefix := t + "/"
		if strings.HasPrefix(branch, prefix) {
			rest := strings.TrimPrefix(branch, prefix)
			return t, rest
		}
	}
	return "", branch
}

// humanize converts kebab-case to space-separated words.
func humanize(s string) string {
	return strings.ReplaceAll(s, "-", " ")
}

// ValidTypes returns the list of valid conventional commit types.
func ValidTypes() []string {
	return append([]string(nil), validTypes...)
}

// BranchName constructs a branch name from type and description.
func BranchName(ccType, description string) string {
	slug := strings.ToLower(strings.TrimSpace(description))
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return fmt.Sprintf("%s/%s", ccType, slug)
}

// ValidatePRTitle returns an error if title doesn't follow Conventional Commits.
func ValidatePRTitle(title string) error {
	if !IsConventionalCommit(title) {
		types := strings.Join(validTypes, "|")
		return fmt.Errorf(
			"PR title %q is not a valid Conventional Commit.\n"+
				"  Expected format: type(scope): description\n"+
				"  Valid types: %s\n"+
				"  Example: feat(auth): add token refresh",
			title, types,
		)
	}
	return nil
}
