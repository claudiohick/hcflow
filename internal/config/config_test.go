package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hickstein/hcflow/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig("my-project", "main")
	if cfg.Project.Name != "my-project" {
		t.Errorf("expected project name 'my-project', got %q", cfg.Project.Name)
	}
	if cfg.Git.DefaultBranch != "main" {
		t.Errorf("expected default branch 'main', got %q", cfg.Git.DefaultBranch)
	}
	if cfg.Git.MergeStrategy != "squash" {
		t.Errorf("expected merge strategy 'squash', got %q", cfg.Git.MergeStrategy)
	}
	if cfg.Release.Provider != "release-please" {
		t.Errorf("expected release provider 'release-please', got %q", cfg.Release.Provider)
	}
	if cfg.Schema != config.CurrentSchema {
		t.Errorf("expected schema %d, got %d", config.CurrentSchema, cfg.Schema)
	}
}

func TestWriteAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, config.FileName)

	cfg := config.DefaultConfig("test-repo", "main")
	cfg.CI.Commands = []string{"make test", "make lint"}

	if err := config.Write(cfgPath, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	loaded, path, err := config.LoadFrom(cfgPath)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if path != cfgPath {
		t.Errorf("expected path %q, got %q", cfgPath, path)
	}
	if loaded.Project.Name != "test-repo" {
		t.Errorf("expected name 'test-repo', got %q", loaded.Project.Name)
	}
	if len(loaded.CI.Commands) != 2 {
		t.Errorf("expected 2 CI commands, got %d", len(loaded.CI.Commands))
	}
}

func TestFind(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, config.FileName)
	if err := os.WriteFile(cfgPath, []byte("schema: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Find from subdirectory
	sub := filepath.Join(dir, "sub", "pkg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	found, err := config.Find(sub)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if found != cfgPath {
		t.Errorf("expected %q, got %q", cfgPath, found)
	}
}

func TestFindNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := config.Find(dir)
	if err == nil {
		t.Error("expected error when no .hcflow.yml found")
	}
}

func TestValidate(t *testing.T) {
	cfg := &config.Config{Schema: 0}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for schema 0")
	}

	cfg2 := &config.Config{Schema: 1}
	if err := cfg2.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg2.Git.DefaultBranch != "main" {
		t.Error("expected default branch to be filled in")
	}
}
