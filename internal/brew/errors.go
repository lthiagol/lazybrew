package brew

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type BrewNotFoundError struct {
	Searched []string
}

func (e *BrewNotFoundError) Error() string {
	return fmt.Sprintf("brew not found in standard locations: %s", strings.Join(e.Searched, ", "))
}

type BrewExitError struct {
	Command  string
	ExitCode int
	Stderr   string
}

func (e *BrewExitError) Error() string {
	return fmt.Sprintf("brew %s exited with code %d", e.Command, e.ExitCode)
}

type JSONParseError struct {
	Command   string
	Cause     error
	RawOutput []byte
}

func (e *JSONParseError) Error() string {
	return fmt.Sprintf("failed to parse brew %s output: %s", e.Command, e.Cause)
}

func (e *JSONParseError) Unwrap() error {
	return e.Cause
}

type TimeoutError struct {
	Command string
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("brew %s timed out after %s", e.Command, e.Timeout)
}

// ServiceNeedsRootError is a services write brew refused for want of root: the
// service is a system one, and needs sudo. lazybrew never runs sudo; the error
// names the command for the user to run in a terminal.
type ServiceNeedsRootError struct {
	Action string
	Name   string
}

func (e *ServiceNeedsRootError) Error() string {
	return fmt.Sprintf("%s: needs sudo — run \"sudo brew services %s %s\" in a terminal", e.Name, e.Action, e.Name)
}

// NeedsSudoError is a brew command that needed a terminal for sudo: sudo had no
// terminal to ask for a password on, so it failed. lazybrew never runs sudo; the
// error names the command for the user to run in a terminal. Err is the
// command's own error.
type NeedsSudoError struct {
	Command string
	// Args are the brew arguments of the command, for a retry in the terminal.
	Args []string
	Err  error
}

func (e *NeedsSudoError) Error() string {
	return fmt.Sprintf("needs sudo — run \"brew %s\" in a terminal", e.Command)
}

func (e *NeedsSudoError) Unwrap() error { return e.Err }

// errorMentions reports whether the error's text, or the stderr of a failed brew
// command, contains one of markers, ignoring case. brew's refusals are read this
// way: the message names the problem, and its stderr says what to do.
func errorMentions(err error, markers []string) bool {
	text := err.Error()
	var exit *BrewExitError
	if errors.As(err, &exit) {
		text += "\n" + exit.Stderr
	}
	text = strings.ToLower(text)
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func IsExitCode(err error, code int) bool {
	var e *BrewExitError
	if errors.As(err, &e) {
		return e.ExitCode == code
	}
	return false
}

// ExitOneReport says what the error of a brew command means, for a command whose
// exit 1 is a result, as the readers above read it. warn is true for the
// commands that print a report on exit 1 (doctor, missing, vulns, and the
// cleanup and autoremove dry runs, -n or --dry-run): the readers parse that
// report. success is true for search, where exit 1 means no results. Exit 1 with
// an Error: line on stderr is a failure for vulns and the dry runs. VulnsReport
// reads vulns that way; for the dry runs it is the Log's own rule. The real
// cleanup and autoremove, without a dry-run flag, and any other error or
// command's exit 1, is a failure: both are false.
func ExitOneReport(args []string, err error) (warn, success bool) {
	var exit *BrewExitError
	if len(args) == 0 || !errors.As(err, &exit) || exit.ExitCode != 1 {
		return false, false
	}
	switch args[0] {
	case "doctor", "missing":
		return true, false
	case "vulns":
		return !hasErrorLine(exit.Stderr), false
	case "cleanup", "autoremove":
		return isDryRun(args) && !hasErrorLine(exit.Stderr), false
	case "search":
		return false, true
	}
	return false, false
}

// isDryRun reports whether a brew command's args ask for a dry run: -n or
// --dry-run among them.
func isDryRun(args []string) bool {
	for _, arg := range args {
		if arg == "-n" || arg == "--dry-run" {
			return true
		}
	}
	return false
}
