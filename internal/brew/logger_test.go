package brew

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEnableFileLoggingWritesToTheFileNotStderr: with the debug log on, a debug
// line lands in the file and stderr stays empty (C13: the TUI owns the terminal).
func TestEnableFileLoggingWritesToTheFileNotStderr(t *testing.T) {
	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stderr
	os.Stderr = stderr
	t.Cleanup(func() {
		os.Stderr = orig
		ResetLogger()
	})

	path := filepath.Join(t.TempDir(), "debug.log")
	if err := EnableFileLogging(path); err != nil {
		t.Fatal(err)
	}
	Logger().Debug("probe-to-file")

	if got, _ := os.ReadFile(stderr.Name()); len(got) != 0 {
		t.Errorf("stderr should stay empty, got %q", got)
	}
	logged, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logged), "probe-to-file") {
		t.Errorf("debug.log lacks the probe line: %q", logged)
	}
}

// TestResetLoggerDiscardsDebug: after ResetLogger a debug line goes nowhere.
func TestResetLoggerDiscardsDebug(t *testing.T) {
	path := filepath.Join(t.TempDir(), "debug.log")
	if err := EnableFileLogging(path); err != nil {
		t.Fatal(err)
	}
	ResetLogger()
	Logger().Debug("probe-after-reset")

	if logged, _ := os.ReadFile(path); strings.Contains(string(logged), "probe-after-reset") {
		t.Errorf("a debug line reached the file after ResetLogger: %q", logged)
	}
}

func TestBrewNotFoundError(t *testing.T) {
	err := &BrewNotFoundError{Searched: []string{"/opt/homebrew/bin/brew"}}
	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
}

func TestBrewExitError(t *testing.T) {
	err := &BrewExitError{Command: "install", ExitCode: 1, Stderr: "not found"}
	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
}

func TestJSONParseError(t *testing.T) {
	inner := &BrewNotFoundError{}
	err := &JSONParseError{Command: "info", Cause: inner, RawOutput: []byte("bad json")}
	if err.Unwrap() != inner {
		t.Error("Unwrap should return the cause")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{Command: "search"}
	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
}
