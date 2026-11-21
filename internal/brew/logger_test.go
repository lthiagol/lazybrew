package brew

import (
	"errors"
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

// TestResetLoggerClosesTheLogFile (LB-21): ResetLogger closes the debug log file
// EnableFileLogging opened, so a write to it after the reset fails.
func TestResetLoggerClosesTheLogFile(t *testing.T) {
	if err := EnableFileLogging(filepath.Join(t.TempDir(), "debug.log")); err != nil {
		t.Fatal(err)
	}
	logMu.Lock()
	f := logFile
	logMu.Unlock()

	ResetLogger()

	if _, err := f.WriteString("after reset"); !errors.Is(err, os.ErrClosed) {
		t.Errorf("write to the debug log after ResetLogger: err = %v, want os.ErrClosed", err)
	}
}

// TestEnableFileLoggingClosesThePreviousFile (LB-21): a second EnableFileLogging
// replaces the first file and closes it, so no file is left open behind it.
func TestEnableFileLoggingClosesThePreviousFile(t *testing.T) {
	t.Cleanup(ResetLogger)
	if err := EnableFileLogging(filepath.Join(t.TempDir(), "first.log")); err != nil {
		t.Fatal(err)
	}
	logMu.Lock()
	first := logFile
	logMu.Unlock()

	if err := EnableFileLogging(filepath.Join(t.TempDir(), "second.log")); err != nil {
		t.Fatal(err)
	}

	if _, err := first.WriteString("after replace"); !errors.Is(err, os.ErrClosed) {
		t.Errorf("write to the replaced debug log: err = %v, want os.ErrClosed", err)
	}
}
