package brew

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Runner interface {
	Execute(ctx context.Context, args ...string) ([]byte, error)
	ExecuteJSON(ctx context.Context, result any, args ...string) error
	ExecuteStream(ctx context.Context, args ...string) (<-chan string, <-chan error)
	BrewPath() string
}

type DefaultRunner struct {
	brewPath string
}

func NewDefaultRunner() (*DefaultRunner, error) {
	path, err := findBrewPath()
	if err != nil {
		return nil, err
	}
	return &DefaultRunner{brewPath: path}, nil
}

func NewDefaultRunnerWithPath(brewPath string) (*DefaultRunner, error) {
	if brewPath == "" {
		return NewDefaultRunner()
	}
	if _, err := os.Stat(brewPath); err != nil {
		return nil, fmt.Errorf("brew path %s: %w", brewPath, err)
	}
	return &DefaultRunner{brewPath: brewPath}, nil
}

func findBrewPath() (string, error) {
	candidates := []string{
		os.Getenv("HOMEBREW_PREFIX") + "/bin/brew",
		"/opt/homebrew/bin/brew",
		"/usr/local/bin/brew",
		"/home/linuxbrew/.linuxbrew/bin/brew",
	}

	for _, c := range candidates {
		if c != "/bin/brew" {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
	}

	if path, err := exec.LookPath("brew"); err == nil {
		return path, nil
	}

	return "", &BrewNotFoundError{Searched: candidates}
}

func (r *DefaultRunner) BrewPath() string {
	return r.brewPath
}

func (r *DefaultRunner) Execute(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.brewPath, args...)
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_ASK=1")
	cmd.Stdin = bytes.NewReader(nil)
	setProcessGroup(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	Logger().Debug("brew exec", "args", strings.Join(args, " "))

	err := finishErr(cmd.Run())
	duration := time.Since(start)

	if err != nil {
		exitCode := 0
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		stderrStr := strings.TrimSpace(stderr.String())
		Logger().Warn("brew failed", "args", strings.Join(args, " "), "exit", exitCode, "duration", duration)
		return stdout.Bytes(), &BrewExitError{
			Command:  strings.Join(args, " "),
			ExitCode: exitCode,
			Stderr:   stderrStr,
		}
	}

	Logger().Debug("brew ok", "args", strings.Join(args, " "), "duration", duration)
	return stdout.Bytes(), nil
}

func (r *DefaultRunner) ExecuteJSON(ctx context.Context, result any, args ...string) error {
	stdout, err := r.Execute(ctx, args...)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(stdout, result); err != nil {
		Logger().Warn("json parse failed", "args", strings.Join(args, " "), "cause", err.Error(), "raw", string(stdout)[:min(len(stdout), 200)])
		return &JSONParseError{
			Command:   strings.Join(args, " "),
			Cause:     err,
			RawOutput: stdout,
		}
	}
	return nil
}

func (r *DefaultRunner) ExecuteStream(ctx context.Context, args ...string) (<-chan string, <-chan error) {
	stdoutChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(stdoutChan)
		defer close(errChan)

		// sendErr delivers at most one error for the whole stream: the
		// manager reads errChan exactly once, so a second send would
		// deadlock the goroutine and leak the process handle (r1
		// finding 8).
		var once sync.Once
		sendErr := func(err error) {
			if err == nil {
				return
			}
			once.Do(func() {
				// A cancel outranks any local error regardless of which
				// one arrives first: the GUI maps context.Canceled to
				// "Cancelled", and a stdout read error that lost the race
				// against the cancel must not mask it (r2 finding NEW-7).
				if errors.Is(ctx.Err(), context.Canceled) {
					err = context.Canceled
				}
				// errChan is buffered(1) and the once guarantees this is
				// the only send ever attempted, so the buffered send
				// always succeeds — no guarded fallback needed
				// (r2 finding NEW-6).
				select {
				case errChan <- err:
				default:
				}
			})
		}

		cmd := exec.CommandContext(ctx, r.brewPath, args...)
		cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_ASK=1")
		cmd.Stdin = bytes.NewReader(nil)
		setProcessGroup(cmd)

		// Stdout and stderr are writers, not pipes, so cmd.Wait owns both
		// copies and waits for them itself, bounded by WaitDelay: a stray
		// child holding the pipes open cannot hang the stream.
		lines := &lineWriter{ctx: ctx, out: stdoutChan}
		var stderr bytes.Buffer
		cmd.Stdout = lines
		cmd.Stderr = &stderr

		if err := cmd.Start(); err != nil {
			sendErr(fmt.Errorf("start: %w", err))
			return
		}

		err := finishErr(cmd.Wait())
		lines.flush()
		if err != nil {
			// A cancelled context killed the process: sendErr already
			// prefers context.Canceled over the local error, so the GUI
			// maps the completion to "Cancelled" instead of a misleading
			// "signal: killed" (r1 finding 10, r2 finding NEW-7).
			stderrStr := strings.TrimSpace(stderr.String())
			errMsg := stderrStr
			if errMsg == "" {
				errMsg = err.Error()
			}
			sendErr(fmt.Errorf("brew %s: %s", strings.Join(args, " "), errMsg))
		}
	}()

	return stdoutChan, errChan
}

// lineWriter splits brew's stdout into lines and sends each one to out. A
// send blocks until the consumer takes the line or ctx is done, in which case
// Write fails and the copy stops.
type lineWriter struct {
	ctx     context.Context
	out     chan<- string
	partial string
}

func (w *lineWriter) Write(p []byte) (int, error) {
	data := w.partial + string(p)
	lines := strings.Split(data, "\n")
	w.partial = lines[len(lines)-1]
	for _, line := range lines[:len(lines)-1] {
		select {
		case w.out <- line:
		case <-w.ctx.Done():
			return 0, w.ctx.Err()
		}
	}
	return len(p), nil
}

// flush sends the unterminated last line, if any, once the output has ended.
func (w *lineWriter) flush() {
	if w.partial == "" {
		return
	}
	select {
	case w.out <- w.partial:
	case <-w.ctx.Done():
	}
	w.partial = ""
}

// finishErr is the error of a finished brew process. exec.ErrWaitDelay means
// brew exited 0 and only a stray child still held its output open past
// WaitDelay, so that is not a failure.
func finishErr(err error) error {
	if errors.Is(err, exec.ErrWaitDelay) {
		return nil
	}
	return err
}

type MockRunner struct {
	ExecuteFn       func(ctx context.Context, args ...string) ([]byte, error)
	ExecuteJSONFn   func(ctx context.Context, result any, args ...string) error
	ExecuteStreamFn func(ctx context.Context, args ...string) (<-chan string, <-chan error)
	BrewPathFn      func() string
}

func (m *MockRunner) Execute(ctx context.Context, args ...string) ([]byte, error) {
	if m.ExecuteFn != nil {
		return m.ExecuteFn(ctx, args...)
	}
	return []byte{}, nil
}

func (m *MockRunner) ExecuteJSON(ctx context.Context, result any, args ...string) error {
	if m.ExecuteJSONFn != nil {
		return m.ExecuteJSONFn(ctx, result, args...)
	}
	stdout, err := m.Execute(ctx, args...)
	if err != nil {
		return err
	}
	return json.Unmarshal(stdout, result)
}

func (m *MockRunner) ExecuteStream(ctx context.Context, args ...string) (<-chan string, <-chan error) {
	if m.ExecuteStreamFn != nil {
		return m.ExecuteStreamFn(ctx, args...)
	}
	ch := make(chan string, 1)
	errCh := make(chan error, 1)
	close(ch)
	close(errCh)
	return ch, errCh
}

func (m *MockRunner) BrewPath() string {
	if m.BrewPathFn != nil {
		return m.BrewPathFn()
	}
	return "/fake/brew"
}

func NewMockRunner() *MockRunner {
	return &MockRunner{}
}

type CommandCallback func(args []string, err error)

type LoggingRunner struct {
	inner   Runner
	OnExec  CommandCallback
	OnStart CommandCallback
}

func NewLoggingRunner(inner Runner, onStart, onExec CommandCallback) *LoggingRunner {
	return &LoggingRunner{
		inner:   inner,
		OnStart: onStart,
		OnExec:  onExec,
	}
}

func (r *LoggingRunner) Execute(ctx context.Context, args ...string) ([]byte, error) {
	if r.OnStart != nil {
		r.OnStart(args, nil)
	}
	out, err := r.inner.Execute(ctx, args...)
	if r.OnExec != nil {
		r.OnExec(args, err)
	}
	return out, err
}

func (r *LoggingRunner) ExecuteJSON(ctx context.Context, result any, args ...string) error {
	if r.OnStart != nil {
		r.OnStart(args, nil)
	}
	err := r.inner.ExecuteJSON(ctx, result, args...)
	if r.OnExec != nil {
		r.OnExec(args, err)
	}
	return err
}

func (r *LoggingRunner) ExecuteStream(ctx context.Context, args ...string) (<-chan string, <-chan error) {
	if r.OnStart != nil {
		r.OnStart(args, nil)
	}
	stdoutChan, errChan := r.inner.ExecuteStream(ctx, args...)
	if r.OnExec == nil {
		return stdoutChan, errChan
	}
	// This goroutine is the inner channel's only reader. It logs the error,
	// then hands the same error to the caller on a channel of its own.
	out := make(chan error, 1)
	go func() {
		defer close(out)
		err := <-errChan
		r.OnExec(args, err)
		out <- err
	}()
	return stdoutChan, out
}

func (r *LoggingRunner) BrewPath() string {
	return r.inner.BrewPath()
}
