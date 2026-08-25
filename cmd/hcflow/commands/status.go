// status.go — hcflow status
package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hickstein/hcflow/internal/github"
	"github.com/hickstein/hcflow/internal/status"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the operational state of this repository",
		Long:  `Display a concise view of git state, current PR, release status, and hcflow configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := loadContext()
			if err != nil {
				return err
			}

			st := status.Gather(ctx.Config, ctx.Git, ctx.GitHub)
			printStatus(ctx.Config.Project.Name, st)
			return nil
		},
	}
}

func printStatus(name string, st *status.State) {
	fmt.Printf("\n%s\n", header(name))

	// Git
	section("Git")
	kv("branch", st.Git.Branch)
	if st.Git.IsClean {
		kvColored("working tree", "clean", green)
	} else {
		kvColored("working tree", "dirty", yellow)
	}
	if st.Git.Ahead > 0 || st.Git.Behind > 0 {
		kv("ahead/behind", fmt.Sprintf("+%d / -%d", st.Git.Ahead, st.Git.Behind))
	}
	if st.Git.LatestTag != "" {
		kv("latest tag", st.Git.LatestTag)
	}

	// PR
	section("Pull Request")
	if st.PR != nil {
		kv("number", fmt.Sprintf("#%d", st.PR.Number))
		kv("title", st.PR.Title)
		ci := github.PRCIStatus(st.PR)
		switch ci {
		case github.CIPassed:
			kvColored("CI", "passed", green)
		case github.CIFailed:
			kvColored("CI", "failed", red)
		case github.CIRunning:
			kvColored("CI", "running", yellow)
		default:
			kv("CI", "unknown")
		}
		if st.PR.Mergeable == "MERGEABLE" {
			kvColored("merge", "ready", green)
		} else {
			kvColored("merge", st.PR.Mergeable, yellow)
		}
		kv("url", st.PR.URL)
	} else {
		info("no open PR for current branch")
	}

	// Release
	section("Release")
	if st.Release != nil {
		if st.Release.Current != "" {
			kv("current", st.Release.Current)
		}
		if st.Release.Proposed != "" {
			kvColored("proposed", st.Release.Proposed, green)
		}
		if st.Release.PR != nil {
			kv("release PR", fmt.Sprintf("#%d", st.Release.PR.Number))
			switch st.Release.CIStatus {
			case github.CIPassed:
				kvColored("status", "ready", green)
			case github.CIFailed:
				kvColored("status", "CI failed", red)
			default:
				kv("status", "pending")
			}
		} else if st.Release.Current != "" {
			kv("status", "no pending release")
		}
	} else {
		info("no release data")
	}

	// hcflow
	section("hcflow")
	kv("schema", fmt.Sprintf("v%d", st.HCFlow.Schema))
	kv("workflows", st.HCFlow.WorkflowVersion)

	fmt.Println()
}
