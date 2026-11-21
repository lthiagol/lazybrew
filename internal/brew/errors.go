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
