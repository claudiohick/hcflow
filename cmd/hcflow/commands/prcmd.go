// prcmd.go — hcflow pr
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hickstein/hcflow/internal/pr"
)

func newPRCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pr",
		Short: "Create a Pull Request from the current branch",
		Long: `Push the current branch and open a Pull Request on GitHub.

  hcflow pr

The PR title must follow Conventional Commits format:
  type(scope): description

hcflow will suggest a title based on the branch name and commits.
You can accept the suggestion or provide your own.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := loadContext()
			if err != nil {
				return err
			}

			if err := ctx.requireGitHub(); err != nil {
				return err
			}

			// Must not be on default branch
			branch, err := ctx.Git.CurrentBranch()
			if err != nil {
				return err
			}
			if branch == ctx.Config.Git.DefaultBranch {
				return fmt.Errorf("cannot create PR from %q — switch to a feature branch", branch)
			}

			// Check uncommitted changes
			clean, err := ctx.Git.IsClean()
			if err != nil {
				return err
			}
			if !clean {
				status, _ := ctx.Git.StatusShort()
				fmt.Println(dim("Uncommitted changes:"))
				for _, line := range strings.Split(status, "\n") {
					if line != "" {
						fmt.Printf("  %s\n", dim(line))
					}
				}
				fmt.Println()
				fmt.Print("Save these changes before creating the PR? [Y/n] ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer == "" || answer == "y" || answer == "yes" {
					msg, err := promptMessage()
					if err != nil {
						return err
					}
					if err := ctx.Git.AddAll(); err != nil {
						return fmt.Errorf("staging: %w", err)
					}
					if err := ctx.Git.Commit(msg); err != nil {
						return fmt.Errorf("committing: %w", err)
					}
					success("saved: %s", msg)
				}
			}

			// Count commits
			base := ctx.Config.Git.DefaultBranch
			count, err := ctx.Git.CommitCount(base)
			if err != nil {
				return fmt.Errorf("counting commits: %w", err)
			}

			// Get recent commits for suggestion
			commits, _ := ctx.Git.RecentCommits(base, 5)

			section("Branch")
			info(branch)
			section("Changes")
			info("%d commit(s)", count)
			fmt.Println()

			// Suggest title
			suggested := pr.SuggestTitle(branch, commits)
			fmt.Printf("Suggested title: %s\n\n", cyan.Sprint(suggested))

			// Get confirmed title
			title, err := confirmTitle(suggested)
			if err != nil {
				return err
			}

			// Validate conventional commit
			if err := pr.ValidatePRTitle(title); err != nil {
				return err
			}

			// Push
			fmt.Print("Pushing… ")
			if err := ctx.Git.PushUpstream(branch); err != nil {
				return fmt.Errorf("push failed: %w", err)
			}
			fmt.Printf("%s pushed\n", checkMark())

			// Check if PR already exists
			existing, _ := ctx.GitHub.PRForBranch(branch)
			if existing != nil {
				warn("PR #%d already exists: %s", existing.Number, existing.URL)
				return nil
			}

			// Create PR
			prState, err := ctx.GitHub.CreatePR(
				title,
				buildPRBody(branch, commits),
				ctx.Config.Git.DefaultBranch,
				branch,
			)
			if err != nil {
				return fmt.Errorf("creating PR: %w", err)
			}

			fmt.Println()
			success("PR #%d created", prState.Number)
			fmt.Printf("\n  %s\n\n", cyan.Sprint(prState.URL))

			return nil
		},
	}
}

func confirmTitle(suggested string) (string, error) {
	fmt.Printf("PR title [press Enter to accept]: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading title: %w", err)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return suggested, nil
	}
	return line, nil
}

func buildPRBody(branch string, commits []string) string {
	var sb strings.Builder
	sb.WriteString("## Changes\n\n")
	if len(commits) > 0 {
		for _, c := range commits {
			sb.WriteString(fmt.Sprintf("- %s\n", c))
		}
	} else {
		sb.WriteString("_No commits yet_\n")
	}
	sb.WriteString("\n---\n_Created with [hcflow](https://github.com/hickstein/hcflow)_")
	return sb.String()
}
