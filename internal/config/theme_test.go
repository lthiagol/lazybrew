package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestThemeEveryNameLoads: each name gui.theme accepts loads as itself.
func TestThemeEveryNameLoads(t *testing.T) {
	t.Setenv(IconsEnv, "")
	for _, name := range themeNames {
		cfg, err := Load(writeConfig(t, "gui:\n  theme: "+name+"\n"))
		if err != nil {
			t.Fatalf("gui.theme %q: %v", name, err)
		}
		if cfg.GUI.Theme != name {
			t.Errorf("gui.theme = %q, want %q", cfg.GUI.Theme, name)
		}
	}
}

// TestThemeNameIsCaseInsensitive: the name is accepted in any case.
func TestThemeNameIsCaseInsensitive(t *testing.T) {
	t.Setenv(IconsEnv, "")
	if _, err := Load(writeConfig(t, "gui:\n  theme: Catppuccin-Macchiato\n")); err != nil {
		t.Fatalf("mixed-case theme should load: %v", err)
	}
}

// TestThemeUnknownIsAnErrorListingTheNames: an unknown theme is a load error
// that names every valid value.
func TestThemeUnknownIsAnErrorListingTheNames(t *testing.T) {
	t.Setenv(IconsEnv, "")
	_, err := Load(writeConfig(t, "gui:\n  theme: solarized\n"))
	if err == nil {
		t.Fatal("want an error for an unknown theme")
	}
	if !strings.Contains(err.Error(), `gui.theme: unknown value "solarized"`) {
		t.Errorf("error = %q, want the unknown value named", err)
	}
	for _, name := range themeNames {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error = %q, want it to list %q", err, name)
		}
	}
}

// TestSampleThemeCommentListsEveryName: the sample's theme comment names
// every value gui.theme accepts.
func TestSampleThemeCommentListsEveryName(t *testing.T) {
	comment := sampleComments["gui.theme"]
	for _, name := range themeNames {
		if !strings.Contains(comment, name) {
			t.Errorf("sample comment %q does not list %q", comment, name)
		}
	}
}

// TestREADMEDocumentsThemes keeps the README's Themes paragraph in step with
// the names gui.theme accepts.
func TestREADMEDocumentsThemes(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	readme := string(data)
	for _, name := range themeNames {
		if !strings.Contains(readme, name) {
			t.Errorf("README.md does not document the theme %q", name)
		}
	}
}
