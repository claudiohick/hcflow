// upgrade.go — hcflow upgrade
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hickstein/hcflow/internal/config"
	"github.com/hickstein/hcflow/internal/migrations"
)

func newUpgradeCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the repository to the current hcflow schema",
		Long: `Upgrade .hcflow.yml and workflow files to the current supported version.

  hcflow upgrade             Apply all pending migrations
  hcflow upgrade --dry-run   Show what would change without applying

Migrations are:
  • versioned and explicit
  • idempotent (safe to run multiple times)
  • non-destructive (never silently removes customizations)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			cfg, cfgPath, err := config.Load(dir)
			if err != nil {
				return err
			}

			currentSchema := cfg.Schema
			targetSchema := config.CurrentSchema

			section("hcflow upgrade")
			kv("current schema", fmt.Sprintf("v%d", currentSchema))
			kv("supported schema", fmt.Sprintf("v%d", targetSchema))

			if migrations.IsUpToDate(cfg) {
				fmt.Println()
				success("already up to date (schema v%d)", currentSchema)
				return nil
			}

			plan := migrations.Plan(currentSchema, targetSchema)
			if len(plan) == 0 {
				success("no migrations to apply")
				return nil
			}

			section("Planned changes")
			for _, m := range plan {
				fmt.Printf("  schema v%d → v%d\n", m.FromSchema, m.ToSchema)
				for _, ch := range m.Changes {
					switch ch.Kind {
					case "add":
						fmt.Printf("    + %s\n", ch.Description)
					case "modify":
						fmt.Printf("    ~ %s\n", ch.Description)
					case "remove":
						fmt.Printf("    - %s\n", ch.Description)
					}
				}
			}
			fmt.Println()

			if dryRun {
				info("[dry-run] no changes applied")
				return nil
			}

			if err := migrations.Run(dir, cfg, plan, dryRun); err != nil {
				return err
			}

			// Save updated config
			if err := config.Write(cfgPath, cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			success("upgrade complete — schema v%d", cfg.Schema)
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would change without applying")
	return cmd
}
