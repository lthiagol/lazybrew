package gui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// logStatusOf starts one brew command through the log hook and finishes it
// with err, as the LoggingRunner does, and returns the status the log gives it.
func logStatusOf(t *testing.T, args []string, err error) CommandStatus {
	t.Helper()
	m := newTestModel()
	finish := m.CommandLogHook()(args)
	finish(err)
	entries := m.commandLog.Entries()
	if len(entries) == 0 {
		t.Fatal("hook logged no entry")
	}
	return entries[len(entries)-1].Status
}

// TestCommandLogStatusOfExitCodes: exit 1 of a report command is a warning, exit
// 1 of search (no results) is a success, and every other error is an error.
func TestCommandLogStatusOfExitCodes(t *testing.T) {
	exit := func(cmd string, code int) error {
		return &brew.BrewExitError{Command: cmd, ExitCode: code}
	}
	cases := []struct {
		name string
		args []string
		err  error
		want CommandStatus
	}{
		{"install ok", []string{"install", "bash"}, nil, CommandSuccess},
		{"doctor exit 1 warns", []string{"doctor"}, exit("doctor", 1), CommandWarn},
		{"missing exit 1 warns", []string{"missing"}, exit("missing", 1), CommandWarn},
		{"vulns exit 1 warns", []string{"vulns", "--json", "--max-summary", "0"}, exit("vulns", 1), CommandWarn},
		{"vulns exit 1 with an Error line errors", []string{"vulns", "--json", "--max-summary", "0"},
			&brew.BrewExitError{Command: "vulns", ExitCode: 1, Stderr: "Error: Unknown command: vulns"}, CommandError},
		{"vulns exit 1 with a warning line warns", []string{"vulns", "--json", "--max-summary", "0"},
			&brew.BrewExitError{Command: "vulns", ExitCode: 1, Stderr: "Warning: stale cache"}, CommandWarn},
		{"search exit 1 is a success", []string{"search", "--formula", "nope"}, exit("search", 1), CommandSuccess},
		{"doctor exit 2 errors", []string{"doctor"}, exit("doctor", 2), CommandError},
		{"missing exit 2 errors", []string{"missing"}, exit("missing", 2), CommandError},
		{"search exit 2 errors", []string{"search", "--cask", "x"}, exit("search", 2), CommandError},
		{"install exit 1 errors", []string{"install", "bash"}, exit("install", 1), CommandError},
		{"list exit 1 errors", []string{"list", "--formula"}, exit("list", 1), CommandError},
		{"doctor timeout errors", []string{"doctor"}, context.DeadlineExceeded, CommandError},
		{"doctor non-exit error errors", []string{"doctor"}, errors.New("exec: brew not found"), CommandError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := logStatusOf(t, tc.args, tc.err); got != tc.want {
				t.Fatalf("status = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestCommandLogWarnRowRendersWarnGlyph: a warning row shows the Warn glyph in
// the warning style, not the Error mark.
func TestCommandLogWarnRowRendersWarnGlyph(t *testing.T) {
	m := newTestModel()
	m.commandLog.SetIcons(style.UnicodeIcons)
	finish := m.CommandLogHook()([]string{"doctor"})
	finish(&brew.BrewExitError{Command: "doctor", ExitCode: 1})

	view := m.commandLog.View(40, 4)
	want := presentation.WithIcon(style.UnicodeIcons.Warn, "brew doctor")
	if got := plainLine(view); !strings.HasPrefix(got, want) {
		t.Fatalf("row = %q, want it to start %q", got, want)
	}
	if strings.Contains(plainLine(view), style.UnicodeIcons.Error) {
		t.Fatalf("warning row shows the Error mark: %q", plainLine(view))
	}
	if !strings.HasPrefix(view, style.WarningText.Render(want)) {
		t.Fatalf("warning row is not in the warning style: %q", view)
	}
}

// loggedModel wires a model the way app.New does: a LoggingRunner over runner,
// with the command log hook.
func loggedModel(runner brew.Runner) (*Model, *brew.LoggingRunner) {
	logRunner := brew.NewLoggingRunner(runner, nil)
	m := New(brew.NewClient(logRunner), testConfig())
	logRunner.OnCommand = m.CommandLogHook()
	return m, logRunner
}

// TestCommandLogOneRowPerBrewCommand: each brew command, whichever Execute kind
// runs it, adds exactly one row, which ends with its final status and is never
// left running.
func TestCommandLogOneRowPerBrewCommand(t *testing.T) {
	doctorErr := &brew.BrewExitError{Command: "doctor", ExitCode: 1}
	installErr := &brew.BrewExitError{Command: "install", ExitCode: 1}
	var m *Model
	// inCall runs inside the fake runner's call, before the command returns. It
	// checks that the log already holds the command's row, still running: the
	// row is written at start, not when the command ends.
	inCall := func(t *testing.T, n int, cmd string) {
		t.Helper()
		entries := m.commandLog.Entries()
		if len(entries) != n {
			t.Fatalf("while %q runs the log has %d rows, want %d: %+v", cmd, len(entries), n, entries)
		}
		if last := entries[n-1]; last.Command != cmd || last.Status != CommandRunning {
			t.Fatalf("while %q runs its row is %q %v, want it running", cmd, last.Command, last.Status)
		}
		for _, e := range entries[:n-1] {
			if e.Status == CommandRunning {
				t.Fatalf("while %q runs an earlier row is still running: %+v", cmd, entries)
			}
		}
	}
	fake := &brewtest.MockRunner{
		ExecuteFn: func(ctx context.Context, args ...string) ([]byte, error) {
			inCall(t, 1, "doctor")
			return nil, doctorErr
		},
		ExecuteJSONFn: func(ctx context.Context, result any, args ...string) error {
			inCall(t, 2, "info --json=v2 wget")
			return nil
		},
		ExecuteStreamFn: func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
			inCall(t, 3, "install bash")
			lines := make(chan string)
			close(lines)
			errs := make(chan error, 1)
			errs <- installErr
			close(errs)
			return lines, errs
		},
	}
	m, logRunner := loggedModel(fake)
	m.commandLog.SetIcons(style.UnicodeIcons)
	ctx := context.Background()

	_, _ = logRunner.Execute(ctx, "doctor")
	var info any
	_ = logRunner.ExecuteJSON(ctx, &info, "info", "--json=v2", "wget")
	lines, errs := logRunner.ExecuteStream(ctx, "install", "bash")
	for range lines {
	}
	<-errs

	want := []struct {
		cmd    string
		status CommandStatus
		glyph  string
	}{
		{"doctor", CommandWarn, style.UnicodeIcons.Warn},
		{"info --json=v2 wget", CommandSuccess, style.UnicodeIcons.OK},
		{"install bash", CommandError, style.UnicodeIcons.Error},
	}
	entries := m.commandLog.Entries()
	if len(entries) != len(want) {
		t.Fatalf("log has %d rows, want %d (one per command): %+v", len(entries), len(want), entries)
	}
	for i, w := range want {
		if entries[i].Command != w.cmd || entries[i].Status != w.status {
			t.Fatalf("row %d = %q %v, want %q %v", i, entries[i].Command, entries[i].Status, w.cmd, w.status)
		}
	}

	view := m.commandLog.View(60, len(want))
	if strings.Contains(view, style.UnicodeIcons.Refresh) {
		t.Fatalf("a row is still running: %q", view)
	}
	rows := plainRows(view)
	if len(rows) != len(want) {
		t.Fatalf("view shows %d rows, want %d: %q", len(rows), len(want), view)
	}
	for i, w := range want {
		if got := string(rows[i]); !strings.HasPrefix(got, presentation.WithIcon(w.glyph, "brew "+w.cmd)) {
			t.Fatalf("view row %d = %q, want glyph %q for %q", i, got, w.glyph, w.cmd)
		}
	}
}
