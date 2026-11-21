package app

import (
	"os"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui"
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

func TestOptionsDefaults(t *testing.T) {
	opts := Options{}
	if opts.Debug {
		t.Error("Debug should default to false")
	}
	if opts.ConfigPath != "" {
		t.Error("ConfigPath should default to empty")
	}
}
