// start.go — hcflow start
package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hickstein/hcflow/internal/pr"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <type> <description>",
		Short: "Start a new unit of work on a new branch",
		Long: `Create a new branch for a unit of work.

  hcflow start feat telegram-callback
  hcflow start fix reconnect-timeout
  hcflow start docs update-readme

Valid types: feat, fix, perf, refactor, docs, test, build, ci, chore

The branch will be named <type>/<description-as-slug>.`,
		Args:    cobra.ExactArgs(2),
		Example: "  hcflow start feat telegram-callback\n  hcflow start fix reconnect",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := loadContext()
			if err != nil {
				return err
			}

			ccType := args[0]
			desc := args[1]

			// Validate type
			validTypes := pr.ValidTypes()
			valid := false
			for _, t := range validTypes {
				if t == ccType {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid type %q — valid types: feat, fix, perf, refactor, docs, test, build, ci, chore", ccType)
			}

			branchName := pr.BranchName(ccType, desc)

			// Check if we're already on correct branch
			current, err := ctx.Git.CurrentBranch()
			if err != nil {
				return fmt.Errorf("getting current branch: %w", err)
			}

			if current == branchName {
				warn("already on branch %s", branchName)
				return nil
			}

			// Check working tree is clean
			clean, err := ctx.Git.IsClean()
			if err != nil {
				return err
			}
			if !clean {
				return fmt.Errorf("working tree has uncommitted changes — run 'hcflow save' or stash your changes first")
			}

			// Switch to default branch and sync
			defaultBranch := ctx.Config.Git.DefaultBranch
			if err := ctx.Git.Checkout(defaultBranch); err != nil {
				return fmt.Errorf("switching to %s: %w", defaultBranch, err)
			}

			if err := ctx.Git.Pull(); err != nil {
				// "no tracking information" is expected for new/unpushed repos — not an error.
				errStr := err.Error()
				if strings.Contains(errStr, "no tracking") ||
					strings.Contains(errStr, "set-upstream") ||
					strings.Contains(errStr, "tracking information") {
					info("%s has no remote tracking branch yet (push after first commit)", defaultBranch)
				} else {
					warn("could not pull latest %s: %v", defaultBranch, err)
				}
			} else {
				success("%s synchronized", defaultBranch)
			}

			// Check branch doesn't already exist
			exists, err := ctx.Git.BranchExists(branchName)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("branch %q already exists — did you mean to switch to it?", branchName)
			}

			// Create and checkout branch
			if err := ctx.Git.CreateBranch(branchName); err != nil {
				return fmt.Errorf("creating branch: %w", err)
			}

			success("created %s", branchName)
			success("switched to %s", branchName)

			return nil
		},
	}
}
