package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.GUI.Theme != "dark" {
		t.Errorf("default theme = %q, want dark", cfg.GUI.Theme)
	}
	if cfg.GUI.SidebarWidth != 35 {
		t.Errorf("default sidebar = %d, want 35", cfg.GUI.SidebarWidth)
	}
	if !cfg.GUI.ShowIcons {
		t.Error("ShowIcons should default to true")
	}
	if cfg.Brew.UpdateOnStart != UpdateAlways {
		t.Errorf("UpdateOnStart = %q, should default to true", cfg.Brew.UpdateOnStart)
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

// A TTL above 0 but under 1s is rejected, and the error names the key and the
// form to write. A value such as 300ns or 500ms parses, so it needs this check.
func TestLoadTTLUnderOneSecondIsError(t *testing.T) {
	tests := []struct {
		line string
		key  string
		eg   string
	}{
		{"formulae_ttl: 300ns", "brew.formulae_ttl", "5m"},
		{"casks_ttl: 500ms", "brew.casks_ttl", "5m"},
		{"outdated_ttl: 1ns", "brew.outdated_ttl", "30m"},
		{"taps_ttl: 999ms", "brew.taps_ttl", "1h"},
		{"services_ttl: 300us", "brew.services_ttl", "30s"},
		{"doctor_ttl: 500ms", "brew.doctor_ttl", "1h"},
	}
	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.yml")
			if err := os.WriteFile(path, []byte("brew:\n  "+tc.line+"\n"), 0644); err != nil {
				t.Fatal(err)
			}
			_, err := Load(path)
			if err == nil {
				t.Fatalf("%s: Load gave no error", tc.line)
			}
			for _, want := range []string{tc.key, "under 1s", "e.g. " + tc.eg} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}

// One second is the smallest TTL accepted; a TTL of 0 still keeps the default.
func TestLoadTTLOneSecondIsAccepted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	data := `brew:
  services_ttl: 1s
  formulae_ttl: 0s
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brew.ServicesTTL != time.Second {
		t.Errorf("ServicesTTL = %s, want 1s", cfg.Brew.ServicesTTL)
	}
	if cfg.Brew.FormulaeTTL != DefaultFormulaeTTL {
		t.Errorf("FormulaeTTL = %s, want the default %s", cfg.Brew.FormulaeTTL, DefaultFormulaeTTL)
	}
}

// A TTL written as a bare number (300) or as text that is no duration is
// rejected, and the error names the key, the value and the form to write.
// yaml's own message for these does not name the key.
func TestLoadTTLNotADurationIsError(t *testing.T) {
	keys := []struct {
		key string
		eg  string
	}{
		{"formulae_ttl", "5m"},
		{"casks_ttl", "5m"},
		{"outdated_ttl", "30m"},
		{"taps_ttl", "1h"},
		{"services_ttl", "30s"},
		{"doctor_ttl", "1h"},
	}
	for _, k := range keys {
		for _, v := range []string{"300", "abc"} {
			t.Run(k.key+" "+v, func(t *testing.T) {
				dir := t.TempDir()
				path := filepath.Join(dir, "config.yml")
				if err := os.WriteFile(path, []byte("brew:\n  "+k.key+": "+v+"\n"), 0644); err != nil {
					t.Fatal(err)
				}
				_, err := Load(path)
				if err == nil {
					t.Fatalf("%s: %s: Load gave no error", k.key, v)
				}
				for _, want := range []string{"brew." + k.key, `"` + v + `"`, "not a duration", "e.g. " + k.eg} {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error = %q, want it to contain %q", err, want)
					}
				}
			})
		}
	}
}

// A TTL given through an anchor (*d) reads as the anchored duration, and an
// alias of a bad value is refused with the value it points at.
func TestLoadTTLAliasIsReadAsItsTarget(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte("x: &d 10m\nbrew:\n  formulae_ttl: *d\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brew.FormulaeTTL != 10*time.Minute {
		t.Errorf("FormulaeTTL = %s, want 10m", cfg.Brew.FormulaeTTL)
	}

	if err := os.WriteFile(path, []byte("x: &d 300\nbrew:\n  formulae_ttl: *d\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = Load(path)
	if err == nil {
		t.Fatal("aliased 300: Load gave no error")
	}
	for _, want := range []string{"brew.formulae_ttl", `"300"`, "not a duration", "e.g. 5m"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err, want)
		}
	}
}

// A local tag on a TTL (!custom 5m) is read as its scalar, as it was before the
// form check: the TTL loads as 5m.
func TestLoadTTLLocalTagLoadsAsScalar(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte("brew:\n  formulae_ttl: !custom 5m\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brew.FormulaeTTL != 5*time.Minute {
		t.Errorf("FormulaeTTL = %s, want 5m", cfg.Brew.FormulaeTTL)
	}
}

// An empty TTL keeps the class default, as it did before the form check.
func TestLoadTTLEmptyKeepsDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte("brew:\n  formulae_ttl:\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brew.FormulaeTTL != DefaultFormulaeTTL {
		t.Errorf("FormulaeTTL = %s, want the default %s", cfg.Brew.FormulaeTTL, DefaultFormulaeTTL)
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

// TestLoadEmptyPathReadsDefaultPath pins the location: the file is written at
// the literal ~/.config/lazybrew/config.yml, not at whatever DefaultPath returns.
func TestLoadEmptyPathReadsDefaultPath(t *testing.T) {
	clearEnvOverrides(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".config", "lazybrew", "config.yml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("gui:\n  theme: light\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GUI.Theme != "light" {
		t.Errorf("Load(\"\") theme = %q, want light from %s", cfg.GUI.Theme, path)
	}
}
