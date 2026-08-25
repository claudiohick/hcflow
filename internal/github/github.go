// Package github wraps the gh CLI for hcflow operations.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Service wraps the gh CLI.
type Service struct {
	dir   string
	owner string
	repo  string
}

// New returns a Service for the given repo.
func New(dir, owner, repo string) *Service {
	return &Service{dir: dir, owner: owner, repo: repo}
}

// gh runs a gh command and returns stdout.
func (s *Service) gh(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	cmd.Dir = s.dir
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh %s: %w — %s", strings.Join(args, " "), err, strings.TrimSpace(errBuf.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// IsAuthenticated reports whether gh is authenticated.
func (s *Service) IsAuthenticated() (bool, error) {
	_, err := s.gh("auth", "status")
	return err == nil, nil
}

// RepoAccessible checks if the repo is visible.
func (s *Service) RepoAccessible() (bool, error) {
	_, err := s.gh("repo", "view", fmt.Sprintf("%s/%s", s.owner, s.repo))
	return err == nil, nil
}

// PRState holds basic PR information.
type PRState struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	State     string `json:"state"`
	Mergeable string `json:"mergeable"`
	HeadRef   string `json:"headRefName"`
	BaseRef   string `json:"baseRefName"`
	Body      string `json:"body"`
	Labels    []struct {
		Name string `json:"name"`
	} `json:"labels"`
	StatusCheckRollup []struct {
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		Name       string `json:"name"`
	} `json:"statusCheckRollup"`
}

// PRForBranch returns the open PR for the given branch, or nil if none.
func (s *Service) PRForBranch(branch string) (*PRState, error) {
	out, err := s.gh("pr", "list",
		"--head", branch,
		"--state", "open",
		"--json", "number,title,url,state,mergeable,headRefName,baseRefName,statusCheckRollup,labels",
		"--limit", "1",
	)
	if err != nil {
		return nil, err
	}
	var prs []PRState
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return nil, fmt.Errorf("parsing PR list: %w", err)
	}
	if len(prs) == 0 {
		return nil, nil
	}
	return &prs[0], nil
}

// CreatePR creates a pull request and returns the PR state.
func (s *Service) CreatePR(title, body, base, head string) (*PRState, error) {
	// `gh pr create` outputs just the PR URL — no --json flag available.
	url, err := s.gh("pr", "create",
		"--title", title,
		"--body", body,
		"--base", base,
		"--head", head,
	)
	if err != nil {
		return nil, err
	}
	url = strings.TrimSpace(url)

	// Fetch structured data via gh pr view.
	out, err := s.gh("pr", "view", url,
		"--json", "number,title,url,state,mergeable,headRefName,baseRefName,statusCheckRollup,labels",
	)
	if err != nil {
		// View failed — return minimal state from the URL.
		return &PRState{URL: url, Title: title}, nil
	}
	var pr PRState
	if err := json.Unmarshal([]byte(out), &pr); err != nil {
		return &PRState{URL: url, Title: title}, nil
	}
	return &pr, nil
}

// MergePR squash-merges the given PR number.
func (s *Service) MergePR(number int) error {
	_, err := s.gh("pr", "merge",
		fmt.Sprintf("%d", number),
		"--squash",
		"--delete-branch",
	)
	return err
}

// ReleasePRs returns open PRs created by release-please (labelled "autorelease: pending").
func (s *Service) ReleasePRs() ([]PRState, error) {
	out, err := s.gh("pr", "list",
		"--label", "autorelease: pending",
		"--state", "open",
		"--json", "number,title,url,state,mergeable,statusCheckRollup,body,labels",
	)
	if err != nil {
		// release-please might not have run yet
		return nil, nil
	}
	var prs []PRState
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return nil, err
	}
	return prs, nil
}

// CIStatus summarises a PR's CI checks.
type CIStatus string

const (
	CIUnknown CIStatus = "unknown"
	CIPending CIStatus = "pending"
	CIRunning CIStatus = "running"
	CIPassed  CIStatus = "passed"
	CIFailed  CIStatus = "failed"
)

// PRCIStatus derives a single CIStatus from a PR's status checks.
func PRCIStatus(pr *PRState) CIStatus {
	if pr == nil || len(pr.StatusCheckRollup) == 0 {
		return CIUnknown
	}
	anyRunning := false
	anyFailed := false
	for _, c := range pr.StatusCheckRollup {
		switch strings.ToUpper(c.Status) {
		case "IN_PROGRESS", "QUEUED", "WAITING":
			anyRunning = true
		case "COMPLETED":
			switch strings.ToUpper(c.Conclusion) {
			case "FAILURE", "TIMED_OUT", "CANCELLED", "ACTION_REQUIRED":
				anyFailed = true
			}
		}
	}
	if anyFailed {
		return CIFailed
	}
	if anyRunning {
		return CIRunning
	}
	return CIPassed
}

// LatestRelease returns the latest GitHub release tag, or "" if none.
func (s *Service) LatestRelease() (string, error) {
	out, err := s.gh("release", "list",
		"--limit", "1",
		"--json", "tagName",
	)
	if err != nil {
		return "", nil
	}
	var releases []struct {
		TagName string `json:"tagName"`
	}
	if err := json.Unmarshal([]byte(out), &releases); err != nil {
		return "", nil
	}
	if len(releases) == 0 {
		return "", nil
	}
	return releases[0].TagName, nil
}

// Owner returns the repo owner.
func (s *Service) Owner() string { return s.owner }

// Repo returns the repo name.
func (s *Service) Repo() string { return s.repo }
