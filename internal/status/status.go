// Package status aggregates the operational state of a repository.
package status

import (
	"strings"

	"github.com/hickstein/hcflow/internal/config"
	"github.com/hickstein/hcflow/internal/git"
	"github.com/hickstein/hcflow/internal/github"
)

// State is a snapshot of the repository's operational state.
type State struct {
	Config  *config.Config
	Git     GitState
	PR      *github.PRState // current branch PR, may be nil
	Release *ReleaseState   // may be nil if no release PR
	HCFlow  HCFlowState
}

// GitState holds git-derived status.
type GitState struct {
	Branch    string
	IsClean   bool
	Ahead     int
	Behind    int
	LatestTag string
}

// ReleaseState holds release-derived status.
type ReleaseState struct {
	Current  string          // current released version
	Proposed string          // version proposed by Release Please
	PR       *github.PRState // the release PR
	CIStatus github.CIStatus
}

// HCFlowState holds hcflow meta status.
type HCFlowState struct {
	Schema          int
	WorkflowVersion string
}

// Gather collects status from git and GitHub.
// ghSvc may be nil when offline/unauthenticated.
func Gather(cfg *config.Config, gitSvc *git.Service, ghSvc *github.Service) *State {
	s := &State{
		Config: cfg,
		HCFlow: HCFlowState{Schema: cfg.Schema, WorkflowVersion: "v1"},
	}

	// Git state
	branch, _ := gitSvc.CurrentBranch()
	clean, _ := gitSvc.IsClean()
	ahead, behind, _ := gitSvc.AheadBehind()
	tag, _ := gitSvc.LatestTag()
	s.Git = GitState{
		Branch:    branch,
		IsClean:   clean,
		Ahead:     ahead,
		Behind:    behind,
		LatestTag: tag,
	}

	if ghSvc == nil {
		return s
	}

	// Current branch PR
	if branch != "" && branch != cfg.Git.DefaultBranch {
		s.PR, _ = ghSvc.PRForBranch(branch)
	}

	// Release state
	releasePRs, _ := ghSvc.ReleasePRs()
	current, _ := ghSvc.LatestRelease()

	if len(releasePRs) > 0 {
		rpr := releasePRs[0]
		s.Release = &ReleaseState{
			Current:  current,
			Proposed: extractVersionFromTitle(rpr.Title),
			PR:       &rpr,
			CIStatus: github.PRCIStatus(&rpr),
		}
	} else if current != "" {
		s.Release = &ReleaseState{Current: current}
	}

	return s
}

// extractVersionFromTitle extracts a version string from a Release Please PR title.
// e.g. "chore(main): release 1.2.3" or "Release v1.2.3"
func extractVersionFromTitle(title string) string {
	for _, word := range strings.Fields(title) {
		word = strings.Trim(word, "(),:")
		candidate := word
		if strings.HasPrefix(candidate, "v") {
			candidate = candidate[1:]
		}
		if looksLikeVersion(candidate) {
			if !strings.HasPrefix(word, "v") {
				return "v" + candidate
			}
			return word
		}
	}
	return ""
}

func looksLikeVersion(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}
