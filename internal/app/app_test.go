package app

import (
	"os"
	"path/filepath"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// TestMain points HOME at an empty directory, so the config the app loads is
// never the developer's own ~/.config/lazybrew/config.yml, and runs with the
// animations off, so the models these tests build draw resting frames.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "lazybrew-app-home")
	if err != nil {
		panic(err)
	}
	os.Setenv("HOME", home)
	os.Setenv(config.AnimationsEnv, "0")
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}

func TestAppModelIsStill(t *testing.T) {
	t.Setenv(config.AnimationsEnv, "0")
	m, err := New(Options{Runner: brewtest.NewMockRunner()})
	if err != nil {
		t.Fatal("New() returned error:", err)
	}
	if m.(*gui.Model).MotionOn() {
		t.Fatal("the app's model animates under LAZYBREW_ANIMATIONS=0")
	}
}

func TestNewWithDefaults(t *testing.T) {
	m, err := New(Options{Runner: brewtest.NewMockRunner()})
	if err != nil {
		t.Fatal("New() returned error:", err)
	}
	if m == nil {
		t.Fatal("New() returned nil model")
	}
}

func TestNewWithDebug(t *testing.T) {
	m, err := New(Options{Debug: true, Runner: brewtest.NewMockRunner()})
	if err != nil {
		t.Fatal("New() with debug returned error:", err)
	}
	if m == nil {
		t.Fatal("New() returned nil model")
	}
}

func TestNewWithConfigPath(t *testing.T) {
	m, err := New(Options{ConfigPath: "/nonexistent/config.yml", Runner: brewtest.NewMockRunner()})
	if err != nil {
		t.Fatal("New() with config path returned error:", err)
	}
	if m == nil {
		t.Fatal("New() returned nil model")
	}
}

// TestNewAppliesTheConfiguredTheme: gui.theme picks the palette the app draws
// with, by name, whatever the case.
func TestNewAppliesTheConfiguredTheme(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	cases := []struct {
		theme  string
		accent string
	}{
		{"dark", "#cba6f7"},
		{"light", "#8839ef"},
		{"Catppuccin-Macchiato", "#c6a0f6"},
		{"dracula", "#bd93f9"},
	}
	for _, tc := range cases {
		path := filepath.Join(t.TempDir(), "config.yml")
		if err := os.WriteFile(path, []byte("gui:\n  theme: "+tc.theme+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := New(Options{ConfigPath: path, Runner: brewtest.NewMockRunner()}); err != nil {
			t.Fatalf("theme %q: New() returned error: %v", tc.theme, err)
		}
		if style.AccentColor != lipgloss.Color(tc.accent) {
			t.Errorf("theme %q: accent = %v, want %s", tc.theme, style.AccentColor, tc.accent)
		}
	}
}

func TestOptionsDefaults(t *testing.T) {
	opts := Options{}
	if opts.Debug {
		t.Error("Debug should default to false")
	}
	if opts.ConfigPath != "" {
		t.Error("ConfigPath should default to empty")
	}
}
