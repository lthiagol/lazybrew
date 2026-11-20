package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
)

// captureStderr points os.Stderr at a file for the rest of the test, so
// anything the app writes there is seen. The brew logger is reset afterwards.
func captureStderr(t *testing.T) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stderr
	os.Stderr = f
	t.Cleanup(func() {
		os.Stderr = orig
		brew.SetDebug(false)
	})
	return f
}

func readAll(t *testing.T, f *os.File) string {
	t.Helper()
	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Debug output must never reach stderr while the TUI runs (C13). With no home
// directory the log file cannot be opened: the app prints one warning before
// the TUI starts, and nothing a brew call logs afterwards goes to stderr.
func TestDebugWithoutHomeLogsNothingToStderr(t *testing.T) {
	t.Setenv("HOME", "")
	stderr := captureStderr(t)

	if _, err := New(Options{Debug: true, ConfigPath: "/nonexistent/config.yml", Runner: brew.NewMockRunner()}); err != nil {
		t.Fatal("New() returned error:", err)
	}
	brew.Logger().Debug("probe-after-setup")

	got := readAll(t, stderr)
	if strings.Contains(got, "probe-after-setup") {
		t.Errorf("a debug line reached stderr after setup:\n%s", got)
	}
	if n := strings.Count(got, "\n"); n != 1 {
		t.Errorf("stderr has %d lines, want the one warning:\n%s", n, got)
	}
}

// With a home directory the debug log goes to its file and stderr stays quiet.
func TestDebugLogGoesToTheFileNotStderr(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	stderr := captureStderr(t)

	if _, err := New(Options{Debug: true, ConfigPath: "/nonexistent/config.yml", Runner: brew.NewMockRunner()}); err != nil {
		t.Fatal("New() returned error:", err)
	}
	brew.Logger().Debug("probe-to-file")

	if got := readAll(t, stderr); got != "" {
		t.Errorf("stderr should stay empty with the log file open, got:\n%s", got)
	}
	logged, err := os.ReadFile(filepath.Join(home, ".config", "lazybrew", "debug.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logged), "probe-to-file") {
		t.Errorf("debug.log lacks the probe line:\n%s", logged)
	}
}
