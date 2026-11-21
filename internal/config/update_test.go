package config

import (
	"os"
	"path/filepath"
	"testing"
)

// loadBrewYAML writes a config whose brew block is body and loads it.
func loadBrewYAML(t *testing.T, body string) (*Config, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("brew:\n"+body), 0o644); err != nil {
		t.Fatal(err)
	}
	return Load(path)
}

func TestUpdateOnStartDefaultsToAlwaysWithSixHours(t *testing.T) {
	cfg := Default()
	if cfg.Brew.UpdateOnStart != UpdateAlways {
		t.Errorf("UpdateOnStart = %q, want true (always)", cfg.Brew.UpdateOnStart)
	}
	if cfg.Brew.UpdateStaleHours != 6 {
		t.Errorf("UpdateStaleHours = %d, want 6", cfg.Brew.UpdateStaleHours)
	}

	cfg, err := loadBrewYAML(t, "  path: /custom/brew\n")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brew.UpdateOnStart != UpdateAlways || cfg.Brew.UpdateStaleHours != 6 {
		t.Errorf("a file without the keys: update_on_start=%q update_stale_hours=%d, want true and 6",
			cfg.Brew.UpdateOnStart, cfg.Brew.UpdateStaleHours)
	}
}

func TestUpdateOnStartValues(t *testing.T) {
	tests := []struct {
		value string
		want  UpdateMode
	}{
		{"true", UpdateAlways},
		{"false", UpdateNever},
		{"auto", UpdateAuto},
		// Quoted, the bools are strings in YAML; they mean the same.
		{`"true"`, UpdateAlways},
		{`"false"`, UpdateNever},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			cfg, err := loadBrewYAML(t, "  update_on_start: "+tt.value+"\n")
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Brew.UpdateOnStart != tt.want {
				t.Errorf("update_on_start: %s = %q, want %q", tt.value, cfg.Brew.UpdateOnStart, tt.want)
			}
		})
	}
}

func TestUpdateOnStartInvalidStringIsAnError(t *testing.T) {
	for _, value := range []string{"sometimes", `"yes please"`, "0"} {
		t.Run(value, func(t *testing.T) {
			if _, err := loadBrewYAML(t, "  update_on_start: "+value+"\n"); err == nil {
				t.Errorf("update_on_start: %s loaded without an error", value)
			}
		})
	}
}

func TestUpdateStaleHoursValues(t *testing.T) {
	cfg, err := loadBrewYAML(t, "  update_stale_hours: 12\n")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brew.UpdateStaleHours != 12 {
		t.Errorf("UpdateStaleHours = %d, want 12", cfg.Brew.UpdateStaleHours)
	}

	cfg, err = loadBrewYAML(t, "  update_stale_hours: 1\n")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brew.UpdateStaleHours != 1 {
		t.Errorf("UpdateStaleHours = %d, want the minimum 1", cfg.Brew.UpdateStaleHours)
	}
}

func TestUpdateStaleHoursBelowOneIsAnError(t *testing.T) {
	for _, value := range []string{"0", "-3"} {
		t.Run(value, func(t *testing.T) {
			if _, err := loadBrewYAML(t, "  update_stale_hours: "+value+"\n"); err == nil {
				t.Errorf("update_stale_hours: %s loaded without an error", value)
			}
		})
	}
}
