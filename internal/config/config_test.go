package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joelazar/kagi/internal/config"
)

func TestLoad_ValidConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `api_key: "test-key"
session_token: "test-token"
defaults:
  format: pretty
  search:
    region: us
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Test via Load by placing a .kagi.yaml in a temp dir
	t.Chdir(dir)

	if err := os.WriteFile(filepath.Join(dir, ".kagi.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if cfg.APIKey != "test-key" {
		t.Errorf("expected api_key 'test-key', got %q", cfg.APIKey)
	}
	if cfg.SessionToken != "test-token" {
		t.Errorf("expected session_token 'test-token', got %q", cfg.SessionToken)
	}
	if cfg.Defaults.Format != "pretty" {
		t.Errorf("expected format 'pretty', got %q", cfg.Defaults.Format)
	}
	if cfg.Defaults.Search.Region != "us" {
		t.Errorf("expected region 'us', got %q", cfg.Defaults.Search.Region)
	}
}

func TestLoad_NoConfigFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HOME", t.TempDir())

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config when no files exist")
	}
}

func TestLoad_LocalConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	content := `api_key: "local-key"
`
	if err := os.WriteFile(filepath.Join(dir, ".kagi.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.APIKey != "local-key" {
		t.Errorf("expected 'local-key', got %q", cfg.APIKey)
	}
}
