package flows_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

// logsKeyPress is a key press of one character.
func logsKeyPress(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

// logsRunner serves the Services panel one running service, moshi-hook, whose
// stdout and stderr logs are the given files. It answers the two reads only.
func logsRunner(t *testing.T, stdout, stderr string) *brewtest.MockRunner {
	t.Helper()
	list, err := json.Marshal([]map[string]any{{
		"name": "moshi-hook", "status": "started", "user": "agent",
		"file": "/x/moshi-hook.service", "exit_code": 0,
	}})
	if err != nil {
		t.Fatal(err)
	}
	info, err := json.Marshal([]map[string]any{{
		"name": "moshi-hook", "service_name": "sh.brew.moshi-hook",
		"running": true, "loaded": true, "registered": true,
		"status": "started", "user": "agent", "file": "/x/moshi-hook.service",
		"command": "moshi-hook serve", "log_path": stdout, "error_log_path": stderr,
	}})
	if err != nil {
		t.Fatal(err)
	}
	return &brewtest.MockRunner{ExecuteFn: func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "services" && args[1] == "list" {
			return list, nil
		}
		if len(args) >= 2 && args[0] == "services" && args[1] == "info" {
			return info, nil
		}
		return []byte{}, nil
	}}
}

// TestLogsFlow opens Services, then its Logs tab: the stdout lines show, a line
// appended to the file shows on the next poll, and e shows the stderr log.
func TestLogsFlow(t *testing.T) {
	dir := t.TempDir()
	stdout := filepath.Join(dir, "moshi-hook.log")
	stderr := filepath.Join(dir, "moshi-hook.err")
	if err := os.WriteFile(stdout, []byte("hook started\nhook ready\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stderr, []byte("hook warning\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tm := testutil.NewTestModelWithRunner(t, logsRunner(t, stdout, stderr), teatest.WithInitialTermSize(120, 40))
	tm.Send(logsKeyPress('6'))
	waitForScreen(t, tm, "started")
	tm.Send(logsKeyPress(']'))
	waitForScreen(t, tm, "ready")

	f, err := os.OpenFile(stdout, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("hook tick\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	waitForScreen(t, tm, "tick")

	tm.Send(logsKeyPress('e'))
	waitForScreen(t, tm, "warning")

	tm.Send(logsKeyPress('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}
