// release.go — hcflow release
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hickstein/hcflow/internal/github"
)

func newReleaseCmd() *cobra.Command {
	var publish bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "release",
		Short: "Manage the current Release Please release",
		Long: `Inspect or publish the pending Release Please release.

  hcflow release            Show release status
  hcflow release --publish  Merge the Release PR to trigger a release`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := loadContext()
			if err != nil {
				return err
			}
			if err := ctx.requireGitHub(); err != nil {
				return err
			}

			// Find release PRs
			releasePRs, err := ctx.GitHub.ReleasePRs()
			if err != nil {
				return fmt.Errorf("fetching release PRs: %w", err)
			}

			current, _ := ctx.GitHub.LatestRelease()

			section("Release")
			if current != "" {
				kv("current", current)
			} else {
				kv("current", "(no releases yet)")
			}

			if len(releasePRs) == 0 {
				fmt.Println()
				info("no pending Release Please PR")
				info("merge feature PRs to main — Release Please will create a release PR automatically")
				return nil
			}

			rpr := releasePRs[0]
			proposed := extractReleasePRVersion(rpr.Title)

			if proposed != "" {
				kvColored("proposed", proposed, green)
			}
			kv("release PR", fmt.Sprintf("#%d — %s", rpr.Number, rpr.Title))

			// CI status
			ci := github.PRCIStatus(&rpr)
			switch ci {
			case github.CIPassed:
				kvColored("CI", "passed", green)
			case github.CIFailed:
				kvColored("CI", "failed", red)
			case github.CIRunning:
				kvColored("CI", "running", yellow)
			default:
				kv("CI", string(ci))
			}

			// Mergeable
			if rpr.Mergeable == "MERGEABLE" {
				kvColored("status", "ready to release", green)
			} else {
				kvColored("status", "not yet mergeable ("+rpr.Mergeable+")", yellow)
			}

			kv("url", rpr.URL)

			if !publish {
				fmt.Println()
				info("To publish this release, run: hcflow release --publish")
				return nil
			}

			// ── Publish ──────────────────────────────────────────────────────

			fmt.Println()
			if ci == github.CIFailed {
				return fmt.Errorf("cannot publish: CI has failed on the release PR")
			}
			if ci != github.CIPassed && ci != github.CIUnknown {
				warn("CI status is %q — proceeding may not be safe", string(ci))
			}

			if rpr.Mergeable != "MERGEABLE" {
				return fmt.Errorf("release PR is not mergeable (%s)", rpr.Mergeable)
			}

			// Confirm
			fmt.Printf("About to merge Release PR #%d", rpr.Number)
			if proposed != "" {
				fmt.Printf(" (%s)", proposed)
			}
			fmt.Println()

			if dryRun {
				fmt.Println(dim("[dry-run] Would merge release PR"))
				return nil
			}

			fmt.Print("Confirm? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}

			fmt.Print("Merging release PR… ")
			if err := ctx.GitHub.MergePR(rpr.Number); err != nil {
				return fmt.Errorf("merging PR: %w", err)
			}
			fmt.Printf("%s done\n", checkMark())
			fmt.Println()
			info("Release Please will now:")
			info("  • update version in CHANGELOG.md")
			info("  • create git tag")
			info("  • create GitHub Release")
			info("Run 'hcflow status' to monitor progress.")

			return nil
		},
	}

	cmd.Flags().BoolVar(&publish, "publish", false, "merge the Release PR to trigger the release")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would happen without making changes")
	return cmd
}

// extractReleasePRVersion pulls a version from a Release Please PR title.
func extractReleasePRVersion(title string) string {
	for _, word := range strings.Fields(title) {
		word = strings.Trim(word, "(),:")
		candidate := word
		if strings.HasPrefix(candidate, "v") {
			candidate = candidate[1:]
		}
		if looksLikeReleasePRVersion(candidate) {
			if !strings.HasPrefix(word, "v") {
				return "v" + candidate
			}
			return word
		}
	}
	return ""
}

func looksLikeReleasePRVersion(s string) bool {
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
