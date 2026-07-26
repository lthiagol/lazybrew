package config

import (
	"os"
	"path/filepath"
	"testing"
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
