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

// CommandEnv is the environment every brew command runs with: lazybrew's own,
// with auto-update and prompts off. The terminal retry of a sudo step uses it too.
func CommandEnv() []string {
	return append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_ASK=1")
}

// newCmd is the brew command for args. Every run sets it up the same way: the
// environment, an empty stdin, and its own process group.
func (r *DefaultRunner) newCmd(ctx context.Context, args []string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, r.brewPath, args...)
	cmd.Env = CommandEnv()
	cmd.Stdin = bytes.NewReader(nil)
	setProcessGroup(cmd)
	return cmd
}

func (r *DefaultRunner) Execute(ctx context.Context, args ...string) ([]byte, error) {
	cmd := r.newCmd(ctx, args)

	// The buffers are closed once Run returns: after exec.ErrWaitDelay, exec's
	// copy goroutines may still be writing, and must not race the reads below.
	var stdout, stderr closingBuffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmdLine := strings.Join(args, " ")
	start := time.Now()
	Logger().Debug("brew exec", "args", cmdLine)

	err := finishErr(cmd.Run())
	duration := time.Since(start)
	out := stdout.close()
	errText := stderr.close()

	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			// brew never ran (missing binary, ctx done before start, a pipe
			// error): keep the cause instead of a made-up exit code.
			Logger().Warn("brew failed to start", "args", cmdLine, "cause", err.Error())
			return out, fmt.Errorf("brew %s: %w", cmdLine, err)
		}
		exitCode := exitErr.ExitCode()
		stderrStr := strings.TrimSpace(string(errText))
		Logger().Warn("brew failed", "args", cmdLine, "exit", exitCode, "duration", duration)
		return out, &BrewExitError{
			Command:  cmdLine,
			ExitCode: exitCode,
			Stderr:   stderrStr,
		}
	}

	Logger().Debug("brew ok", "args", cmdLine, "duration", duration)
	return out, nil
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

// ExecuteStream runs brew and sends its stdout and stderr lines on one channel
// as they are written. Lines of the two streams may interleave; each stream's
// own order is kept. The error, sent once, is built from the whole stderr text.
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
		// child holding the pipes open cannot hang the stream. Each stream
		// has its own lineWriter, so a line never mixes the two. After
		// exec.ErrWaitDelay the copies may still be writing, so flush closes
		// each writer: a later write is dropped, and nothing is sent on
		// stdoutChan once it is closed.
		lines := &lineWriter{ctx: ctx, out: stdoutChan}
		errLines := &lineWriter{ctx: ctx, out: stdoutChan, keep: &bytes.Buffer{}}
		cmd.Stdout = lines
		cmd.Stderr = stderrCopy{lines: errLines}

		if err := cmd.Start(); err != nil {
			sendErr(fmt.Errorf("start: %w", err))
			return
		}

		err := finishErr(cmd.Wait())
		lines.flush()
		errLines.flush()
		if err != nil {
			// A cancelled context killed the process: sendErr already
			// prefers context.Canceled over the local error, so the GUI
			// maps the completion to "Cancelled" instead of a misleading
			// "signal: killed" (r1 finding 10, r2 finding NEW-7).
			stderrStr := strings.TrimSpace(errLines.kept())
			errMsg := stderrStr
			if errMsg == "" {
				errMsg = err.Error()
			}
			sendErr(fmt.Errorf("brew %s: %s", strings.Join(args, " "), errMsg))
		}
	}()

	return stdoutChan, errChan
}

// stderrCopy is brew's stderr for ExecuteStream. lines keeps every byte, which
// the error message is built from, and sends the lines on the stream. A failed
// line send (ctx done) is dropped here: the kept text must be whole, and a
// cancel already ends the run.
type stderrCopy struct {
	lines *lineWriter
}

func (w stderrCopy) Write(p []byte) (int, error) {
	_, _ = w.lines.Write(p)
	return len(p), nil
}

// closingBuffer is a buffer exec's copy goroutine writes to. close returns its
// bytes and drops every later write, so a copy still running after
// exec.ErrWaitDelay never races the reader.
type closingBuffer struct {
	mu     sync.Mutex
	buf    bytes.Buffer
	closed bool
}

func (b *closingBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		b.buf.Write(p)
	}
	return len(p), nil
}

func (b *closingBuffer) close() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	return b.buf.Bytes()
}

// lineWriter splits brew's output into lines and sends each one to out. Each
// stream gets its own lineWriter, since a partial line is buffered per writer.
// A line ends at \n, at \r (brew's progress bars redraw with \r), or at \r\n,
// which is one break. A send blocks until the consumer takes the line or ctx
// is done, in which case Write fails and the copy stops.
//
// flush closes the writer: later writes are dropped, since exec's copy can
// outlive cmd.Wait (exec.ErrWaitDelay) and out is closed after flush.
type lineWriter struct {
	ctx context.Context
	out chan<- string
	// keep, when set, gets every byte written before the writer closed.
	keep *bytes.Buffer

	mu     sync.Mutex
	closed bool
	buf    bytes.Buffer
	// afterCR is set when the last byte written ended a line with \r. A \n
	// that follows is the rest of the same \r\n break.
	afterCR bool
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := len(p)
	if w.closed {
		return n, nil
	}
	if w.keep != nil {
		w.keep.Write(p)
	}
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

// flush sends the unterminated last line, if any, once the output has ended,
// and closes the writer. It waits for a Write in progress to finish.
func (w *lineWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closed = true
	if w.buf.Len() == 0 {
		return
	}
	_ = w.emit()
}

// kept is the text keep holds.
func (w *lineWriter) kept() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.keep.String()
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

// CommandHook runs when a brew command starts. It returns the finisher, which
// the runner calls once with the command's final error.
type CommandHook func(args []string) (finish func(err error))

type LoggingRunner struct {
	inner     Runner
	OnCommand CommandHook
}

// NewLoggingRunner wraps inner. A nil hook logs nothing.
func NewLoggingRunner(inner Runner, onCommand CommandHook) *LoggingRunner {
	return &LoggingRunner{
		inner:     inner,
		OnCommand: onCommand,
	}
}

// start calls the hook for a new command, or returns a no-op finisher when
// there is no hook.
func (r *LoggingRunner) start(args []string) func(error) {
	if r.OnCommand == nil {
		return func(error) {}
	}
	return r.OnCommand(args)
}

func (r *LoggingRunner) Execute(ctx context.Context, args ...string) ([]byte, error) {
	finish := r.start(args)
	out, err := r.inner.Execute(ctx, args...)
	finish(err)
	return out, err
}

func (r *LoggingRunner) ExecuteJSON(ctx context.Context, result any, args ...string) error {
	finish := r.start(args)
	err := r.inner.ExecuteJSON(ctx, result, args...)
	finish(err)
	return err
}

func (r *LoggingRunner) ExecuteStream(ctx context.Context, args ...string) (<-chan string, <-chan error) {
	finish := r.start(args)
	stdoutChan, errChan := r.inner.ExecuteStream(ctx, args...)
	if r.OnCommand == nil {
		return stdoutChan, errChan
	}
	// This goroutine is the inner channel's only reader. It finishes the log
	// row, then hands the same error to the caller on a channel of its own.
	out := make(chan error, 1)
	go func() {
		defer close(out)
		err := <-errChan
		finish(err)
		out <- err
	}()
	return stdoutChan, out
}

func (r *LoggingRunner) BrewPath() string {
	return r.inner.BrewPath()
}
