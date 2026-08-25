// Package commands contains all hcflow CLI command implementations.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "hcflow",
	Short: "Standardise your development flow",
	Long: `hcflow — opinionated Git/GitHub workflow orchestrator.

Orchestrates Git, GitHub CLI, GitHub Actions, Conventional Commits
and Release Please into a low-friction development experience.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	rootCmd.Version = Version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, errorStyle(err.Error()))
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(
		newInitCmd(),
		newStartCmd(),
		newSaveCmd(),
		newStatusCmd(),
		newPRCmd(),
		newReleaseCmd(),
		newDoctorCmd(),
		newUpgradeCmd(),
		newUICmd(),
	)
}
