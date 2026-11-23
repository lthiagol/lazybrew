package brew

import (
	"errors"
	"testing"
)

// TestExitOneReport: exit 1 is a warning for the dry runs, unless stderr holds an
// Error: line; the real cleanup and autoremove, and every other exit, fail.
func TestExitOneReport(t *testing.T) {
	exit := func(code int, stderr string) error {
		return &BrewExitError{Command: "cmd", ExitCode: code, Stderr: stderr}
	}
	cases := []struct {
		name        string
		args        []string
		err         error
		wantWarn    bool
		wantSuccess bool
	}{
		{"cleanup -n exit 1 warns", []string{"cleanup", "-n"}, exit(1, ""), true, false},
		{"cleanup -n with a warning line warns", []string{"cleanup", "-n"}, exit(1, "Warning: stale"), true, false},
		{"cleanup -n with an Error line fails", []string{"cleanup", "-n"}, exit(1, "Error: boom"), false, false},
		{"cleanup --dry-run exit 1 warns", []string{"cleanup", "--dry-run"}, exit(1, ""), true, false},
		{"cleanup -n --prune=all exit 1 warns", []string{"cleanup", "-n", "--prune=all"}, exit(1, ""), true, false},
		{"autoremove -n exit 1 warns", []string{"autoremove", "-n"}, exit(1, ""), true, false},
		{"autoremove -n with an Error line fails", []string{"autoremove", "-n"}, exit(1, "Error: boom"), false, false},
		{"autoremove --dry-run exit 1 warns", []string{"autoremove", "--dry-run"}, exit(1, ""), true, false},
		{"cleanup exit 1 fails", []string{"cleanup"}, exit(1, ""), false, false},
		{"autoremove exit 1 fails", []string{"autoremove"}, exit(1, ""), false, false},
		{"cleanup -n exit 2 fails", []string{"cleanup", "-n"}, exit(2, ""), false, false},
		{"autoremove -n exit 2 fails", []string{"autoremove", "-n"}, exit(2, ""), false, false},
		{"cleanup -n non-exit error fails", []string{"cleanup", "-n"}, errors.New("exec: brew not found"), false, false},
		{"cleanup -n no error fails", []string{"cleanup", "-n"}, nil, false, false},
		{"install exit 1 fails", []string{"install", "-n"}, exit(1, ""), false, false},
		{"doctor exit 1 still warns", []string{"doctor"}, exit(1, ""), true, false},
		{"search exit 1 still succeeds", []string{"search", "x"}, exit(1, ""), false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			warn, success := ExitOneReport(tc.args, tc.err)
			if warn != tc.wantWarn || success != tc.wantSuccess {
				t.Fatalf("ExitOneReport(%v) = warn %v, success %v; want warn %v, success %v",
					tc.args, warn, success, tc.wantWarn, tc.wantSuccess)
			}
		})
	}
}
