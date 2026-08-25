// init.go — hcflow init
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/hickstein/hcflow/internal/config"
	"github.com/hickstein/hcflow/internal/git"
	"github.com/hickstein/hcflow/internal/github"
)

func newInitCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialise hcflow in the current repository",
		Long: `Configure the current GitHub repository to use hcflow.

  hcflow init             Apply all configuration
  hcflow init --dry-run   Show what would be created/changed

Creates:
  • .hcflow.yml
  • .github/workflows/hcflow-ci.yml
  • .github/workflows/hcflow-release.yml
  • .github/pull_request_template.md
  • release-please-config.json
  • .release-please-manifest.json

hcflow init is idempotent — safe to run multiple times.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			section("hcflow init")

			// 1. Validate git repo
			gitSvc := git.New(dir)
			if !gitSvc.IsRepo() {
				return fmt.Errorf("not a git repository — run 'git init' first")
			}
			success("git repository")

			// 2. Identify remote GitHub
			remoteURL := gitSvc.RemoteOrigin()
			owner, repo, err := git.ParseGitHubRemote(remoteURL)
			if err != nil {
				return fmt.Errorf("no GitHub remote found: %w\n  Set a remote with: git remote add origin https://github.com/<owner>/<repo>", err)
			}
			success("GitHub remote: %s/%s", owner, repo)

			// 3. Identify default branch
			defaultBranch, err := gitSvc.DefaultBranch("origin")
			if err != nil {
				defaultBranch = "main"
			}
			success("default branch: %s", defaultBranch)

			// 4. Verify gh CLI
			ghSvc := github.New(dir, owner, repo)
			ok, _ := ghSvc.IsAuthenticated()
			if !ok {
				return fmt.Errorf("gh CLI is not authenticated — run: gh auth login")
			}
			success("gh authenticated")

			// 5. Project name
			projectName := repo

			// ── Create files ─────────────────────────────────────────────────

			type fileToCreate struct {
				path    string
				content string
				skip    bool // skip if exists
			}

			// Compute CI commands — default to "make test" if no Makefile
			ciCommands := []string{"make test"}
			if _, err := os.Stat(filepath.Join(dir, "Makefile")); os.IsNotExist(err) {
				ciCommands = []string{"echo 'No CI commands configured — edit .hcflow.yml'"}
			}

			// Build .hcflow.yml content
			cfg := config.DefaultConfig(projectName, defaultBranch)
			cfg.CI.Commands = ciCommands
			cfgPath := filepath.Join(dir, config.FileName)

			var files []fileToCreate

			// .hcflow.yml — skip if exists (idempotent)
			cfgExists := false
			if _, err := os.Stat(cfgPath); err == nil {
				cfgExists = true
			}

			// CI workflow
			ciWorkflow, err := renderCIWorkflow(owner, repo, cfg)
			if err != nil {
				return err
			}

			// Release workflow
			releaseWorkflow, err := renderReleaseWorkflow(owner, repo, defaultBranch)
			if err != nil {
				return err
			}

			files = append(files,
				fileToCreate{
					path:    filepath.Join(dir, ".github", "workflows", "hcflow-ci.yml"),
					content: ciWorkflow,
				},
				fileToCreate{
					path:    filepath.Join(dir, ".github", "workflows", "hcflow-release.yml"),
					content: releaseWorkflow,
				},
				fileToCreate{
					path:    filepath.Join(dir, ".github", "pull_request_template.md"),
					content: prTemplate,
					skip:    true,
				},
				fileToCreate{
					path:    filepath.Join(dir, "release-please-config.json"),
					content: releasePleaseConfig(owner, repo),
					skip:    true,
				},
				fileToCreate{
					path:    filepath.Join(dir, ".release-please-manifest.json"),
					content: `{}`,
					skip:    true,
				},
			)

			// Apply
			fmt.Println()

			// Config file
			if cfgExists {
				info(".hcflow.yml already exists — skipping")
			} else {
				if !dryRun {
					if err := config.Write(cfgPath, cfg); err != nil {
						return fmt.Errorf("writing .hcflow.yml: %w", err)
					}
				}
				if dryRun {
					info("[dry-run] would create .hcflow.yml")
				} else {
					success("created .hcflow.yml")
				}
			}

			// Other files
			for _, f := range files {
				exists := false
				if _, err := os.Stat(f.path); err == nil {
					exists = true
				}

				rel, _ := filepath.Rel(dir, f.path)

				if exists && f.skip {
					info("%s already exists — skipping", rel)
					continue
				}

				if dryRun {
					if exists {
						info("[dry-run] would update %s", rel)
					} else {
						info("[dry-run] would create %s", rel)
					}
					continue
				}

				if err := writeFile(f.path, f.content); err != nil {
					return fmt.Errorf("writing %s: %w", rel, err)
				}

				if exists {
					success("updated %s", rel)
				} else {
					success("created %s", rel)
				}
			}

			fmt.Println()
			if dryRun {
				info("[dry-run] no changes applied")
				return nil
			}

			success("hcflow init complete")
			fmt.Println()
			info("Next steps:")
			info("  hcflow doctor        — verify configuration")
			info("  hcflow start feat X  — begin work")

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be created without applying")
	return cmd
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// ── Templates ──────────────────────────────────────────────────────────────

func renderCIWorkflow(owner, repo string, cfg *config.Config) (string, error) {
	tmpl := `# hcflow-workflow-version: v1
# Generated by hcflow — do not edit manually (run 'hcflow upgrade' to update)
name: CI

on:
  pull_request:
    branches: ["{{ .DefaultBranch }}"]

permissions:
  contents: read
  pull-requests: read

jobs:
  ci:
    name: CI
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Run CI commands
        run: |
{{ .Commands }}
`
	type data struct {
		DefaultBranch string
		Commands      string
	}

	var cmds []string
	for _, c := range cfg.CI.Commands {
		cmds = append(cmds, "          "+c)
	}

	d := data{
		DefaultBranch: cfg.Git.DefaultBranch,
		Commands:      strings.Join(cmds, "\n"),
	}

	t, err := template.New("ci").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := t.Execute(&sb, d); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func renderReleaseWorkflow(owner, repo, defaultBranch string) (string, error) {
	return fmt.Sprintf(`# hcflow-workflow-version: v1
# Generated by hcflow — do not edit manually
name: Release Please

on:
  push:
    branches: ["%s"]

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    name: Release Please
    runs-on: ubuntu-latest
    steps:
      - name: Release Please
        uses: googleapis/release-please-action@v4
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
`, defaultBranch), nil
}

func releasePleaseConfig(owner, repo string) string {
	return fmt.Sprintf(`{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "release-type": "simple",
  "packages": {
    ".": {}
  },
  "changelog-sections": [
    {"type": "feat",     "section": "Features"},
    {"type": "fix",      "section": "Bug Fixes"},
    {"type": "perf",     "section": "Performance Improvements"},
    {"type": "refactor", "section": "Code Refactoring", "hidden": true},
    {"type": "docs",     "section": "Documentation",    "hidden": true},
    {"type": "chore",    "section": "Miscellaneous",    "hidden": true}
  ]
}
`)
}

const prTemplate = `## Summary

<!-- Describe the change and motivation -->

## Type of change

<!-- The PR title must follow Conventional Commits: type(scope): description -->
<!-- Valid types: feat, fix, perf, refactor, docs, test, build, ci, chore -->

## Checklist

- [ ] Title follows Conventional Commits format
- [ ] Tests pass (CI will verify)
- [ ] Self-reviewed

---
_Created with [hcflow](https://github.com/hickstein/hcflow)_
`
