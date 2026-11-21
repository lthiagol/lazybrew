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

// brewCandidates are the places findBrewPath looks for brew, in order. The
// HOMEBREW_PREFIX candidate is there only when the variable is set.
func brewCandidates() []string {
	var candidates []string
	if prefix := os.Getenv("HOMEBREW_PREFIX"); prefix != "" {
		candidates = append(candidates, prefix+"/bin/brew")
	}
	return append(candidates,
		"/opt/homebrew/bin/brew",
		"/usr/local/bin/brew",
		"/home/linuxbrew/.linuxbrew/bin/brew",
	)
}

func findBrewPath() (string, error) {
	candidates := brewCandidates()

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
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

// newCmd is the brew command for args. Every run sets it up the same way: the
// environment, an empty stdin, and its own process group.
func (r *DefaultRunner) newCmd(ctx context.Context, args []string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, r.brewPath, args...)
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_ASK=1")
	cmd.Stdin = bytes.NewReader(nil)
	setProcessGroup(cmd)
	return cmd
}

func (r *DefaultRunner) Execute(ctx context.Context, args ...string) ([]byte, error) {
	cmd := r.newCmd(ctx, args)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmdLine := strings.Join(args, " ")
	start := time.Now()
	Logger().Debug("brew exec", "args", cmdLine)

	err := finishErr(cmd.Run())
	duration := time.Since(start)

	if err != nil {
		exitCode := 0
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		stderrStr := strings.TrimSpace(stderr.String())
		Logger().Warn("brew failed", "args", cmdLine, "exit", exitCode, "duration", duration)
		return stdout.Bytes(), &BrewExitError{
			Command:  cmdLine,
			ExitCode: exitCode,
			Stderr:   stderrStr,
		}
	}

	Logger().Debug("brew ok", "args", cmdLine, "duration", duration)
	return stdout.Bytes(), nil
}

func (r *DefaultRunner) ExecuteJSON(ctx context.Context, result any, args ...string) error {
	stdout, err := r.Execute(ctx, args...)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(stdout, result); err != nil {
		cmdLine := strings.Join(args, " ")
		Logger().Warn("json parse failed", "args", cmdLine, "cause", err.Error(), "raw", string(stdout)[:min(len(stdout), 200)])
		return &JSONParseError{
			Command:   cmdLine,
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

		cmd := r.newCmd(ctx, args)

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
// line ends at \n, at \r (brew's progress bars redraw with \r), or at \r\n,
// which is one break. A send blocks until the consumer takes the line or ctx
// is done, in which case Write fails and the copy stops.
type lineWriter struct {
	ctx context.Context
	out chan<- string
	buf bytes.Buffer
	// afterCR is set when the last byte written ended a line with \r. A \n
	// that follows is the rest of the same \r\n break.
	afterCR bool
}

func (w *lineWriter) Write(p []byte) (int, error) {
	n := len(p)
	for len(p) > 0 {
		if w.afterCR && p[0] == '\n' {
			p = p[1:]
			w.afterCR = false
			continue
		}
		w.afterCR = false
		end := lineEnd(p)
		if end < 0 {
			w.buf.Write(p)
			break
		}
		w.buf.Write(p[:end])
		if err := w.emit(); err != nil {
			return n - len(p), err
		}
		w.afterCR = p[end] == '\r'
		p = p[end+1:]
	}
	return n, nil
}

// lineEnd is the index of the first \r or \n in p, or -1 when p has neither.
// It scans p with bytes.IndexByte: once for \n, then for \r only up to it.
func lineEnd(p []byte) int {
	nl := bytes.IndexByte(p, '\n')
	stop := nl
	if nl < 0 {
		stop = len(p)
	}
	if cr := bytes.IndexByte(p[:stop], '\r'); cr >= 0 {
		return cr
	}
	return nl
}

// emit sends the buffered line and empties the buffer.
func (w *lineWriter) emit() error {
	line := w.buf.String()
	w.buf.Reset()
	select {
	case w.out <- line:
		return nil
	case <-w.ctx.Done():
		return w.ctx.Err()
	}
}

// flush sends the unterminated last line, if any, once the output has ended.
func (w *lineWriter) flush() {
	if w.buf.Len() == 0 {
		return
	}
	_ = w.emit()
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
