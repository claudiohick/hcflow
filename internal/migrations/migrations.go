// Package migrations handles schema upgrades for .hcflow.yml and related files.
package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hickstein/hcflow/internal/config"
)

// Change describes a single migration change.
type Change struct {
	Kind        string // "add", "modify", "remove"
	Description string
}

// Migration describes a schema version upgrade.
type Migration struct {
	FromSchema int
	ToSchema   int
	Changes    []Change
	Apply      func(dir string, cfg *config.Config, dryRun bool) error
}

// all registered migrations in order.
var all = []Migration{
	// Example: v1 -> v2 (for future use)
	// {
	// 	FromSchema: 1,
	// 	ToSchema:   2,
	// 	Changes: []Change{
	// 		{Kind: "add", Description: "improved CI permissions in workflow"},
	// 		{Kind: "modify", Description: "workflows v1 → v2"},
	// 	},
	// 	Apply: migrateV1toV2,
	// },
}

// Plan computes the migrations needed to upgrade from currentSchema to targetSchema.
func Plan(currentSchema, targetSchema int) []Migration {
	var plan []Migration
	for _, m := range all {
		if m.FromSchema >= currentSchema && m.ToSchema <= targetSchema {
			plan = append(plan, m)
		}
	}
	return plan
}

// Run executes a migration plan.
func Run(dir string, cfg *config.Config, plan []Migration, dryRun bool) error {
	for _, m := range plan {
		fmt.Printf("  migrating schema v%d → v%d\n", m.FromSchema, m.ToSchema)
		if !dryRun {
			if err := m.Apply(dir, cfg, dryRun); err != nil {
				return fmt.Errorf("migration v%d→v%d: %w", m.FromSchema, m.ToSchema, err)
			}
			cfg.Schema = m.ToSchema
		}
	}
	return nil
}

// IsUpToDate reports whether the repo is at the current schema.
func IsUpToDate(cfg *config.Config) bool {
	return cfg.Schema >= config.CurrentSchema
}

// WorkflowVersion returns the hcflow workflow version embedded in a workflow file.
func WorkflowVersion(dir string) string {
	path := filepath.Join(dir, ".github", "workflows", "hcflow-ci.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# hcflow-workflow-version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# hcflow-workflow-version:"))
		}
	}
	return "v1"
}
