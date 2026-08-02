package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.GUI.Theme != "dark" {
		t.Errorf("default theme = %q, want dark", cfg.GUI.Theme)
	}
	if cfg.GUI.SidebarWidth != 30 {
		t.Errorf("default sidebar = %d, want 30", cfg.GUI.SidebarWidth)
	}
	if !cfg.GUI.ShowIcons {
		t.Error("ShowIcons should default to true")
	}
	if cfg.Brew.UpdateOnStart {
		t.Error("UpdateOnStart should default to false")
	}
	if cfg.Brew.OutdatedTTL != DefaultOutdatedTTL {
		t.Errorf("OutdatedTTL = %s, want %s (30m default)", cfg.Brew.OutdatedTTL, DefaultOutdatedTTL)
	}
	if cfg.Brew.OutdatedTTL != 30*time.Minute {
		t.Errorf("DefaultOutdatedTTL should be 30m, got %s", cfg.Brew.OutdatedTTL)
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GUI.Theme != "dark" {
		t.Errorf("got theme %q, want dark (defaults)", cfg.GUI.Theme)
	}
}

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	data := `gui:
  theme: light
  sidebar_width: 25
  show_icons: false
  mouse: false
  auto_refresh_seconds: 30
brew:
  path: /custom/brew
  update_on_start: true
  outdated_ttl: 5m
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.GUI.Theme != "light" {
		t.Errorf("theme = %q, want light", cfg.GUI.Theme)
	}
	if cfg.GUI.SidebarWidth != 25 {
		t.Errorf("sidebar = %d, want 25", cfg.GUI.SidebarWidth)
	}
	if cfg.GUI.ShowIcons {
		t.Error("ShowIcons should be false")
	}
	if cfg.Brew.Path != "/custom/brew" {
		t.Errorf("brew path = %q", cfg.Brew.Path)
	}
	if cfg.Brew.OutdatedTTL != 5*time.Minute {
		t.Errorf("OutdatedTTL = %s, want 5m", cfg.Brew.OutdatedTTL)
	}
}

func TestLoadMissingOutdatedTTLFallsBackToDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	data := `brew:
  path: /custom/brew
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brew.OutdatedTTL != DefaultOutdatedTTL {
		t.Errorf("missing OutdatedTTL should fall back to default, got %s", cfg.Brew.OutdatedTTL)
	}
}

func TestLoadZeroOutdatedTTLFallsBackToDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	data := `brew:
  outdated_ttl: 0s
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brew.OutdatedTTL != DefaultOutdatedTTL {
		t.Errorf("zero OutdatedTTL should fall back to default, got %s", cfg.Brew.OutdatedTTL)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte("invalid: [yaml: broken"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
