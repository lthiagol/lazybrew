package gui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// logStatusOf runs the log callback for one finished brew command, as the
// LoggingRunner does, and returns the status the log gives that command.
func logStatusOf(t *testing.T, args []string, err error) CommandStatus {
	t.Helper()
	m := newTestModel()
	m.CommandLogCallback()(args, err)
	entries := m.commandLog.Entries()
	if len(entries) == 0 {
		t.Fatal("callback logged no entry")
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
	m.CommandLogCallback()([]string{"doctor"}, &brew.BrewExitError{Command: "doctor", ExitCode: 1})

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
