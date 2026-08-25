// Package config handles .hcflow.yml parsing, validation and writing.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	FileName      = ".hcflow.yml"
	CurrentSchema = 1
)

// Config is the root configuration structure.
type Config struct {
	Schema  int           `yaml:"schema"`
	Project ProjectConfig `yaml:"project"`
	Git     GitConfig     `yaml:"git"`
	CI      CIConfig      `yaml:"ci"`
	Release ReleaseConfig `yaml:"release"`
	GitHub  GitHubConfig  `yaml:"github"`
	Deploy  DeployConfig  `yaml:"deploy"`
}

type ProjectConfig struct {
	Name string `yaml:"name"`
}

type GitConfig struct {
	DefaultBranch string `yaml:"default_branch"`
	MergeStrategy string `yaml:"merge_strategy"`
}

type CIConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Commands []string `yaml:"commands"`
}

type ReleaseConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Provider  string `yaml:"provider"`
	Strategy  string `yaml:"strategy"`
}

type GitHubConfig struct {
	PullRequests     bool `yaml:"pull_requests"`
	LinearHistory    bool `yaml:"linear_history"`
	RequiredApprovals int  `yaml:"required_approvals"`
}

type DeployConfig struct {
	Enabled bool `yaml:"enabled"`
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig(name, branch string) *Config {
	return &Config{
		Schema: CurrentSchema,
		Project: ProjectConfig{
			Name: name,
		},
		Git: GitConfig{
			DefaultBranch: branch,
			MergeStrategy: "squash",
		},
		CI: CIConfig{
			Enabled:  true,
			Commands: []string{"make test"},
		},
		Release: ReleaseConfig{
			Enabled:  true,
			Provider: "release-please",
			Strategy: "semver",
		},
		GitHub: GitHubConfig{
			PullRequests:     true,
			LinearHistory:    true,
			RequiredApprovals: 0,
		},
		Deploy: DeployConfig{
			Enabled: false,
		},
	}
}

// Load reads and parses .hcflow.yml from the given directory (or any parent).
func Load(dir string) (*Config, string, error) {
	path, err := Find(dir)
	if err != nil {
		return nil, "", err
	}
	return LoadFrom(path)
}

// LoadFrom reads and parses .hcflow.yml from an explicit path.
func LoadFrom(path string) (*Config, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, path, fmt.Errorf("reading %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, path, fmt.Errorf("parsing %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, path, err
	}
	return &cfg, path, nil
}

// Find walks up the directory tree looking for .hcflow.yml.
func Find(start string) (string, error) {
	dir := start
	for {
		candidate := filepath.Join(dir, FileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("%s not found (run 'hcflow init' first)", FileName)
}

// Exists reports whether .hcflow.yml exists in dir.
func Exists(dir string) bool {
	_, err := Find(dir)
	return err == nil
}

// Validate checks configuration consistency.
func (c *Config) Validate() error {
	if c.Schema < 1 {
		return fmt.Errorf("invalid schema version: %d", c.Schema)
	}
	if c.Git.DefaultBranch == "" {
		c.Git.DefaultBranch = "main"
	}
	if c.Git.MergeStrategy == "" {
		c.Git.MergeStrategy = "squash"
	}
	if c.Release.Provider == "" {
		c.Release.Provider = "release-please"
	}
	if c.Release.Strategy == "" {
		c.Release.Strategy = "semver"
	}
	return nil
}

// Write serialises cfg to path (creating or overwriting).
func Write(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("serialising config: %w", err)
	}
	header := "# hcflow configuration — https://github.com/hickstein/hcflow\n"
	return os.WriteFile(path, append([]byte(header), data...), 0o644)
}

// RootDir returns the directory containing .hcflow.yml.
func RootDir(cfgPath string) string {
	return filepath.Dir(cfgPath)
}
