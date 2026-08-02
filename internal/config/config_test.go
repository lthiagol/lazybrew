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
	// M12 per-class defaults.
	if cfg.Brew.FormulaeTTL != DefaultFormulaeTTL {
		t.Errorf("FormulaeTTL = %s, want %s", cfg.Brew.FormulaeTTL, DefaultFormulaeTTL)
	}
	if cfg.Brew.CasksTTL != DefaultCasksTTL {
		t.Errorf("CasksTTL = %s, want %s", cfg.Brew.CasksTTL, DefaultCasksTTL)
	}
	if cfg.Brew.TapsTTL != DefaultTapsTTL {
		t.Errorf("TapsTTL = %s, want %s", cfg.Brew.TapsTTL, DefaultTapsTTL)
	}
	if cfg.Brew.ServicesTTL != DefaultServicesTTL {
		t.Errorf("ServicesTTL = %s, want %s", cfg.Brew.ServicesTTL, DefaultServicesTTL)
	}
	if cfg.Brew.DoctorTTL != DefaultDoctorTTL {
		t.Errorf("DoctorTTL = %s, want %s", cfg.Brew.DoctorTTL, DefaultDoctorTTL)
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
  formulae_ttl: 10m
  casks_ttl: 15m
  taps_ttl: 2h
  services_ttl: 45s
  doctor_ttl: 90m
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
	if cfg.GUI.AutoRefreshSeconds != 30 {
		t.Errorf("AutoRefreshSeconds = %d, want 30", cfg.GUI.AutoRefreshSeconds)
	}
	if cfg.Brew.Path != "/custom/brew" {
		t.Errorf("brew path = %q", cfg.Brew.Path)
	}
	if cfg.Brew.OutdatedTTL != 5*time.Minute {
		t.Errorf("OutdatedTTL = %s, want 5m", cfg.Brew.OutdatedTTL)
	}
	if cfg.Brew.FormulaeTTL != 10*time.Minute {
		t.Errorf("FormulaeTTL = %s, want 10m", cfg.Brew.FormulaeTTL)
	}
	if cfg.Brew.CasksTTL != 15*time.Minute {
		t.Errorf("CasksTTL = %s, want 15m", cfg.Brew.CasksTTL)
	}
	if cfg.Brew.TapsTTL != 2*time.Hour {
		t.Errorf("TapsTTL = %s, want 2h", cfg.Brew.TapsTTL)
	}
	if cfg.Brew.ServicesTTL != 45*time.Second {
		t.Errorf("ServicesTTL = %s, want 45s", cfg.Brew.ServicesTTL)
	}
	if cfg.Brew.DoctorTTL != 90*time.Minute {
		t.Errorf("DoctorTTL = %s, want 90m", cfg.Brew.DoctorTTL)
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

func TestLoadZeroAllTTLsFallBackToDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	data := `brew:
  formulae_ttl: 0s
  casks_ttl: 0s
  outdated_ttl: 0s
  taps_ttl: 0s
  services_ttl: 0s
  doctor_ttl: 0s
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		got  time.Duration
		want time.Duration
	}{
		{"FormulaeTTL", cfg.Brew.FormulaeTTL, DefaultFormulaeTTL},
		{"CasksTTL", cfg.Brew.CasksTTL, DefaultCasksTTL},
		{"OutdatedTTL", cfg.Brew.OutdatedTTL, DefaultOutdatedTTL},
		{"TapsTTL", cfg.Brew.TapsTTL, DefaultTapsTTL},
		{"ServicesTTL", cfg.Brew.ServicesTTL, DefaultServicesTTL},
		{"DoctorTTL", cfg.Brew.DoctorTTL, DefaultDoctorTTL},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s = %s, want %s (default)", tc.name, tc.got, tc.want)
		}
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
