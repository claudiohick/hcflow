// save.go — hcflow save
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newSaveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "save [message]",
		Short: "Commit all current changes",
		Long: `Stage and commit all current changes.

  hcflow save "implement callback handler"
  hcflow save   (prompts for message interactively)

The message does NOT need to follow Conventional Commits format.
These commits are on a feature branch and will be squashed into the
PR title on merge.`,
		Args:    cobra.MaximumNArgs(1),
		Example: "  hcflow save \"fix reconnect logic\"",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := loadContext()
			if err != nil {
				return err
			}

			// Check we're not on default branch — unless this is the very first commit.
			branch, err := ctx.Git.CurrentBranch()
			if err != nil {
				return err
			}
			if branch == ctx.Config.Git.DefaultBranch {
				hasCommits, _ := ctx.Git.HasCommits()
				if hasCommits {
					return fmt.Errorf("cannot save on %q — create a branch first with 'hcflow start'", branch)
				}
				// First commit: allow it.
				warn("saving directly on %q (initial commit)", branch)
			}

			// Show current status
			status, _ := ctx.Git.StatusShort()
			if status == "" {
				warn("nothing to commit — working tree is clean")
				return nil
			}

			fmt.Println(dim("Changes to be committed:"))
			for _, line := range strings.Split(status, "\n") {
				if line != "" {
					fmt.Printf("  %s\n", dim(line))
				}
			}
			fmt.Println()

			// Get message
			var message string
			if len(args) > 0 {
				message = strings.TrimSpace(args[0])
			} else {
				message, err = promptMessage()
				if err != nil {
					return err
				}
			}

			if message == "" {
				return fmt.Errorf("commit message cannot be empty")
			}

			// Stage all
			if err := ctx.Git.AddAll(); err != nil {
				return fmt.Errorf("staging changes: %w", err)
			}

			// Commit
			if err := ctx.Git.Commit(message); err != nil {
				return fmt.Errorf("committing: %w", err)
			}

			success("saved: %s", message)
			info("branch: %s", branch)

			return nil
		},
	}
}

func promptMessage() (string, error) {
	fmt.Print("Commit message: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading message: %w", err)
	}
	return strings.TrimSpace(line), nil
}
