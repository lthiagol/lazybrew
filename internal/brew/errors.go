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

func IsExitCode(err error, code int) bool {
	var e *BrewExitError
	if errors.As(err, &e) {
		return e.ExitCode == code
	}
	return false
}
