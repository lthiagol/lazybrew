package brew

import (
	"context"
	"errors"
	"strings"
)

// terminalMarkers are sudo's words for want of a terminal: it has nothing to
// ask for a password on, so it refuses to run.
var terminalMarkers = []string{"a terminal is required", "no tty present", "a password is required"}

// sudoRunner wraps a Runner. A command that failed because sudo needed a
// terminal returns a NeedsSudoError naming the command; every other error
// passes unchanged.
type sudoRunner struct {
	inner Runner
}

func (r sudoRunner) BrewPath() string {
	return r.inner.BrewPath()
}

func (r sudoRunner) Execute(ctx context.Context, args ...string) ([]byte, error) {
	out, err := r.inner.Execute(ctx, args...)
	return out, sudoError(args, err)
}

func (r sudoRunner) ExecuteJSON(ctx context.Context, result any, args ...string) error {
	return sudoError(args, r.inner.ExecuteJSON(ctx, result, args...))
}

func (r sudoRunner) ExecuteStream(ctx context.Context, args ...string) (<-chan string, <-chan error) {
	lines, errs := r.inner.ExecuteStream(ctx, args...)
	out := make(chan error, 1)
	go func() {
		defer close(out)
		if err := <-errs; err != nil {
			out <- sudoError(args, err)
		}
	}()
	return lines, out
}

// sudoError is err as NeedsSudoError when brew's words say sudo needed a
// terminal, and err unchanged otherwise.
func sudoError(args []string, err error) error {
	if err == nil || !needsTerminal(err) {
		return err
	}
	return &NeedsSudoError{Command: strings.Join(args, " "), Err: err}
}

// needsTerminal reads brew's words for a terminal refusal: the error text and
// the stderr brew prints sudo's refusal to.
func needsTerminal(err error) bool {
	text := err.Error()
	var exit *BrewExitError
	if errors.As(err, &exit) {
		text += "\n" + exit.Stderr
	}
	text = strings.ToLower(text)
	for _, marker := range terminalMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}
