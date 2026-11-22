package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeConfig writes a config file with the given body and returns its path.
func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestIconsDefaultsToNerd(t *testing.T) {
	t.Setenv(IconsEnv, "")
	if got := Default().GUI.Icons; got != IconsNerd {
		t.Errorf("default gui.icons = %q, want %q", got, IconsNerd)
	}
	cfg, err := Load(writeConfig(t, "gui:\n  mouse: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GUI.Icons != IconsNerd {
		t.Errorf("gui.icons with no key = %q, want %q", cfg.GUI.Icons, IconsNerd)
	}
}

func TestIconsFromFile(t *testing.T) {
	t.Setenv(IconsEnv, "")
	cfg, err := Load(writeConfig(t, "gui:\n  icons: unicode\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GUI.Icons != IconsUnicode {
		t.Errorf("gui.icons = %q, want %q", cfg.GUI.Icons, IconsUnicode)
	}
}

func TestIconsEnvOverridesFile(t *testing.T) {
	t.Setenv(IconsEnv, "unicode")
	cfg, err := Load(writeConfig(t, "gui:\n  icons: nerd\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GUI.Icons != IconsUnicode {
		t.Errorf("env should override the file: gui.icons = %q, want %q", cfg.GUI.Icons, IconsUnicode)
	}
}

func TestIconsEnvEmptyIsIgnored(t *testing.T) {
	t.Setenv(IconsEnv, "")
	cfg, err := Load(writeConfig(t, "gui:\n  icons: unicode\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GUI.Icons != IconsUnicode {
		t.Errorf("empty env must not override: gui.icons = %q, want %q", cfg.GUI.Icons, IconsUnicode)
	}
}

func TestIconsUnknownValueIsAnError(t *testing.T) {
	tests := []struct {
		name string
		env  string
		body string
		want string
	}{
		{"file", "", "gui:\n  icons: emoji\n", "gui.icons: unknown value \"emoji\""},
		{"env", "emoji", "gui:\n  icons: nerd\n", IconsEnv + ": unknown value \"emoji\""},
		// A bad file value is an error even when the environment overrides it:
		// the file is still wrong.
		{"file under env", "unicode", "gui:\n  icons: emoji\n", "gui.icons: unknown value \"emoji\""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(IconsEnv, tc.env)
			_, err := Load(writeConfig(t, tc.body))
			if err == nil {
				t.Fatal("want an error for an unknown icon set")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to contain %q", err, tc.want)
			}
		})
	}
}

// TestSampleDocumentsIcons keeps the sample's icon keys in step with the two
// icon keys, so a user can find them with lazybrew --config-sample.
func TestSampleDocumentsIcons(t *testing.T) {
	sample := Sample()
	for _, key := range []string{"icons: nerd", "show_icons: true", IconsEnv} {
		if !strings.Contains(sample, key) {
			t.Errorf("the sample does not document %q", key)
		}
	}
}
