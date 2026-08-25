// Package ci provides utilities for CI configuration validation.
package ci

import (
	"fmt"
	"strings"
)

// Config holds CI configuration from .hcflow.yml.
type Config struct {
	Enabled  bool
	Commands []string
}

// Validate checks CI configuration is usable.
func Validate(cfg Config) error {
	if !cfg.Enabled {
		return nil
	}
	if len(cfg.Commands) == 0 {
		return fmt.Errorf("ci.enabled is true but ci.commands is empty")
	}
	return nil
}

// CommandsAsMultiline renders commands as a multiline string for workflow templates.
func CommandsAsMultiline(commands []string) string {
	return strings.Join(commands, "\n")
}
