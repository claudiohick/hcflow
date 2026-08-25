// Package git wraps Git CLI operations needed by hcflow.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Service provides Git operations.
type Service struct {
	dir string // working directory
}

// New returns a Service rooted at dir.
func New(dir string) *Service {
	return &Service{dir: dir}
}

// run executes a git command and returns trimmed stdout.
func (s *Service) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.dir
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, errBuf.String())
	}
	return strings.TrimSpace(out.String()), nil
}

// runInteractive runs git with stdio attached to the terminal.
func (s *Service) runInteractive(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// IsRepo reports whether dir is inside a git repository.
func (s *Service) IsRepo() bool {
	_, err := s.run("rev-parse", "--git-dir")
	return err == nil
}

// CurrentBranch returns the current branch name.
func (s *Service) CurrentBranch() (string, error) {
	return s.run("symbolic-ref", "--short", "HEAD")
}

// HasCommits reports whether the repo has at least one commit.
func (s *Service) HasCommits() (bool, error) {
	_, err := s.run("rev-parse", "HEAD")
	return err == nil, nil
}


func (s *Service) DefaultBranch(remote string) (string, error) {
	out, err := s.run("remote", "show", remote)
	if err != nil {
		return s.localDefaultBranch(), nil
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "HEAD branch:") {
			branch := strings.TrimSpace(strings.TrimPrefix(line, "HEAD branch:"))
			// Remote reports "(unknown)" when it has no commits yet.
			if branch == "" || branch == "(unknown)" {
				return s.localDefaultBranch(), nil
			}
			return branch, nil
		}
	}
	return s.localDefaultBranch(), nil
}

// localDefaultBranch returns the current local branch, or "master"/"main" as last resort.
func (s *Service) localDefaultBranch() string {
	if branch, err := s.CurrentBranch(); err == nil && branch != "" {
		return branch
	}
	// Check which of master/main exists locally.
	if _, err := s.run("rev-parse", "--verify", "master"); err == nil {
		return "master"
	}
	return "main"
}


// RemoteURL returns the URL of the given remote.
func (s *Service) RemoteURL(remote string) (string, error) {
	return s.run("remote", "get-url", remote)
}

// RemoteOrigin returns origin URL or empty string.
func (s *Service) RemoteOrigin() string {
	url, _ := s.RemoteURL("origin")
	return url
}

// IsClean reports whether the working tree has no uncommitted changes.
func (s *Service) IsClean() (bool, error) {
	out, err := s.run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out == "", nil
}

// StatusShort returns a short status summary.
func (s *Service) StatusShort() (string, error) {
	return s.run("status", "--short")
}

// AheadBehind returns how many commits ahead and behind the tracking branch.
func (s *Service) AheadBehind() (ahead, behind int, err error) {
	out, e := s.run("rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if e != nil {
		return 0, 0, nil // no upstream — not an error
	}
	parts := strings.Fields(out)
	if len(parts) != 2 {
		return 0, 0, nil
	}
	fmt.Sscan(parts[0], &ahead)
	fmt.Sscan(parts[1], &behind)
	return
}

// BranchExists reports whether a local branch exists.
func (s *Service) BranchExists(name string) (bool, error) {
	_, err := s.run("rev-parse", "--verify", name)
	return err == nil, nil
}

// CreateBranch creates a new branch from HEAD.
func (s *Service) CreateBranch(name string) error {
	_, err := s.run("checkout", "-b", name)
	return err
}

// Checkout switches to the named branch.
func (s *Service) Checkout(branch string) error {
	_, err := s.run("checkout", branch)
	return err
}

// Pull runs git pull --ff-only.
func (s *Service) Pull() error {
	_, err := s.run("pull", "--ff-only")
	return err
}

// PushUpstream pushes the current branch to origin with tracking.
func (s *Service) PushUpstream(branch string) error {
	_, err := s.run("push", "--set-upstream", "origin", branch)
	return err
}

// Add stages all changes.
func (s *Service) AddAll() error {
	_, err := s.run("add", "-A")
	return err
}

// Commit creates a commit with msg.
func (s *Service) Commit(msg string) error {
	_, err := s.run("commit", "-m", msg)
	return err
}

// CommitCount returns number of commits on current branch not on base.
func (s *Service) CommitCount(base string) (int, error) {
	out, err := s.run("rev-list", "--count", base+"..HEAD")
	if err != nil {
		return 0, err
	}
	var n int
	fmt.Sscan(out, &n)
	return n, nil
}

// RecentCommits returns the last n commit subjects on current branch not on base.
func (s *Service) RecentCommits(base string, n int) ([]string, error) {
	out, err := s.run("log", fmt.Sprintf("%s..HEAD", base),
		"--format=%s", fmt.Sprintf("-n%d", n))
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// LatestTag returns the most recent semver tag, or "" if none.
func (s *Service) LatestTag() (string, error) {
	out, err := s.run("describe", "--tags", "--abbrev=0", "--match", "v*")
	if err != nil {
		return "", nil
	}
	return out, nil
}

// ParseGitHubRemote extracts "owner/repo" from a GitHub remote URL.
func ParseGitHubRemote(url string) (owner, repo string, err error) {
	// supports https://github.com/owner/repo.git and git@github.com:owner/repo.git
	url = strings.TrimSuffix(url, ".git")
	if strings.Contains(url, "github.com") {
		var path string
		if strings.HasPrefix(url, "git@") {
			// git@github.com:owner/repo
			parts := strings.SplitN(url, ":", 2)
			if len(parts) == 2 {
				path = parts[1]
			}
		} else {
			// https://github.com/owner/repo
			parts := strings.SplitN(url, "github.com/", 2)
			if len(parts) == 2 {
				path = parts[1]
			}
		}
		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}
	return "", "", fmt.Errorf("cannot parse GitHub remote from: %s", url)
}
