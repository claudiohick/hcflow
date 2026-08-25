// Package doctor implements the hcflow doctor diagnostic.
package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hickstein/hcflow/internal/config"
	"github.com/hickstein/hcflow/internal/git"
	"github.com/hickstein/hcflow/internal/github"
)

// Severity levels for diagnostic findings.
type Severity string

const (
	SeverityOK      Severity = "ok"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
	SeverityInfo    Severity = "info"
)

// Finding is a single diagnostic result.
type Finding struct {
	Category string
	Name     string
	Severity Severity
	Message  string
}

// Report is the full doctor report.
type Report struct {
	Findings []Finding
	Healthy  bool
}

// Run performs a full diagnostic.
func Run(dir string) *Report {
	r := &Report{Healthy: true}
	check := func(category, name string, fn func() (Severity, string)) {
		sev, msg := fn()
		r.Findings = append(r.Findings, Finding{
			Category: category,
			Name:     name,
			Severity: sev,
			Message:  msg,
		})
		if sev == SeverityError {
			r.Healthy = false
		}
	}

	gitSvc := git.New(dir)

	// ── Repository ───────────────────────────────────────────────────────────
	check("Repository", "git installed", func() (Severity, string) {
		if _, err := exec.LookPath("git"); err != nil {
			return SeverityError, "git not found in PATH"
		}
		return SeverityOK, "git available"
	})

	check("Repository", "git repository", func() (Severity, string) {
		if !gitSvc.IsRepo() {
			return SeverityError, "not a git repository"
		}
		return SeverityOK, "valid git repository"
	})

	// ── GitHub ───────────────────────────────────────────────────────────────
	check("GitHub", "gh installed", func() (Severity, string) {
		if _, err := exec.LookPath("gh"); err != nil {
			return SeverityError, "gh not found in PATH — install from https://cli.github.com"
		}
		return SeverityOK, "gh available"
	})

	var owner, repo string
	check("GitHub", "remote origin", func() (Severity, string) {
		url := gitSvc.RemoteOrigin()
		if url == "" {
			return SeverityError, "no remote 'origin' configured"
		}
		var err error
		owner, repo, err = git.ParseGitHubRemote(url)
		if err != nil {
			return SeverityWarning, fmt.Sprintf("remote is not a GitHub URL: %s", url)
		}
		return SeverityOK, fmt.Sprintf("github.com/%s/%s", owner, repo)
	})

	ghSvc := github.New(dir, owner, repo)
	check("GitHub", "authenticated", func() (Severity, string) {
		ok, _ := ghSvc.IsAuthenticated()
		if !ok {
			return SeverityError, "not authenticated — run: gh auth login"
		}
		return SeverityOK, "authenticated"
	})

	if owner != "" && repo != "" {
		check("GitHub", "repository accessible", func() (Severity, string) {
			ok, _ := ghSvc.RepoAccessible()
			if !ok {
				return SeverityWarning, fmt.Sprintf("cannot access %s/%s", owner, repo)
			}
			return SeverityOK, "accessible"
		})
	}

	// ── Configuration ─────────────────────────────────────────────────────────
	var cfg *config.Config
	check("Configuration", ".hcflow.yml", func() (Severity, string) {
		var cfgPath string
		var err error
		cfg, cfgPath, err = config.Load(dir)
		if err != nil {
			return SeverityError, err.Error()
		}
		_ = cfgPath
		return SeverityOK, fmt.Sprintf("schema v%d", cfg.Schema)
	})

	if cfg != nil {
		check("Configuration", "schema version", func() (Severity, string) {
			if cfg.Schema < config.CurrentSchema {
				return SeverityWarning, fmt.Sprintf("schema v%d < current v%d — run: hcflow upgrade",
					cfg.Schema, config.CurrentSchema)
			}
			if cfg.Schema > config.CurrentSchema {
				return SeverityWarning, fmt.Sprintf("schema v%d is newer than this hcflow version",
					cfg.Schema)
			}
			return SeverityOK, fmt.Sprintf("v%d (current)", cfg.Schema)
		})
	}

	// ── Workflows ─────────────────────────────────────────────────────────────
	check("Workflow", "CI workflow", func() (Severity, string) {
		path := fmt.Sprintf("%s/.github/workflows/hcflow-ci.yml", dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return SeverityWarning, "hcflow-ci.yml not found — run: hcflow init"
		}
		return SeverityOK, "configured"
	})

	check("Workflow", "Release Please workflow", func() (Severity, string) {
		path := fmt.Sprintf("%s/.github/workflows/hcflow-release.yml", dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return SeverityWarning, "hcflow-release.yml not found — run: hcflow init"
		}
		return SeverityOK, "configured"
	})

	// ── Merge strategy ────────────────────────────────────────────────────────
	if cfg != nil {
		check("Rules", "merge strategy", func() (Severity, string) {
			if cfg.Git.MergeStrategy != "squash" {
				return SeverityWarning, fmt.Sprintf("merge strategy is %q, expected 'squash'", cfg.Git.MergeStrategy)
			}
			return SeverityOK, "squash merge"
		})
	}

	// ── Template ─────────────────────────────────────────────────────────────
	check("GitHub", "PR template", func() (Severity, string) {
		path := fmt.Sprintf("%s/.github/pull_request_template.md", dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return SeverityInfo, "PR template not found — run: hcflow init"
		}
		return SeverityOK, "configured"
	})

	return r
}

// Summary returns a one-line summary of the report.
func (r *Report) Summary() string {
	if r.Healthy {
		return "healthy"
	}
	errors := 0
	warnings := 0
	for _, f := range r.Findings {
		switch f.Severity {
		case SeverityError:
			errors++
		case SeverityWarning:
			warnings++
		}
	}
	parts := []string{}
	if errors > 0 {
		parts = append(parts, fmt.Sprintf("%d error(s)", errors))
	}
	if warnings > 0 {
		parts = append(parts, fmt.Sprintf("%d warning(s)", warnings))
	}
	return strings.Join(parts, ", ")
}
