package brew

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	Logger().Debug("brew exec", "args", strings.Join(args, " "))

	err := cmd.Run()
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

		cmd := exec.CommandContext(ctx, r.brewPath, args...)
		cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_ASK=1")
		cmd.Stdin = bytes.NewReader(nil)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			errChan <- fmt.Errorf("stdout pipe: %w", err)
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			errChan <- fmt.Errorf("stderr pipe: %w", err)
			return
		}

		if err := cmd.Start(); err != nil {
			errChan <- fmt.Errorf("start: %w", err)
			return
		}

		go func() {
			buf := make([]byte, 4096)
			var partial string
			for {
				n, err := stdout.Read(buf)
				if n > 0 {
					data := partial + string(buf[:n])
					lines := strings.Split(data, "\n")
					partial = lines[len(lines)-1]
					for _, line := range lines[:len(lines)-1] {
						select {
						case stdoutChan <- line:
						case <-ctx.Done():
							return
						}
					}
				}
				if err == io.EOF {
					if partial != "" {
						select {
						case stdoutChan <- partial:
						case <-ctx.Done():
						}
					}
					break
				}
				if err != nil {
					errChan <- fmt.Errorf("stdout read: %w", err)
					break
				}
			}
		}()

		stderrData, _ := io.ReadAll(stderr)
		stderrStr := strings.TrimSpace(string(stderrData))

		if err := cmd.Wait(); err != nil {
			errMsg := stderrStr
			if errMsg == "" {
				errMsg = err.Error()
			}
			errChan <- fmt.Errorf("brew %s: %s", strings.Join(args, " "), errMsg)
		}
	}()

	return stdoutChan, errChan
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
	if r.OnExec != nil {
		go func() {
			err := <-errChan
			r.OnExec(args, err)
		}()
	}
	return stdoutChan, errChan
}

func (r *LoggingRunner) BrewPath() string {
	return r.inner.BrewPath()
}
