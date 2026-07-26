package brew

import (
	"context"
	"strings"
	"testing"
	"time"
)

// RecordingRunner wraps MockRunner and records all Execute calls.
type RecordingRunner struct {
	MockRunner
	Calls []RecordedCall
}

type RecordedCall struct {
	Args []string
	At   time.Time
}

func NewRecordingRunner() *RecordingRunner {
	r := &RecordingRunner{}
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		r.Calls = append(r.Calls, RecordedCall{
			Args: args,
			At:   time.Now(),
		})
		return []byte{}, nil
	}
	return r
}

func AssertCalled(t *testing.T, r *RecordingRunner, args ...string) {
	t.Helper()
	want := strings.Join(args, " ")
	for _, call := range r.Calls {
		if strings.Join(call.Args, " ") == want {
			return
		}
	}
	t.Errorf("expected call %q not found in %d recorded calls", want, len(r.Calls))
	for i, call := range r.Calls {
		t.Logf("  call %d: %s", i, strings.Join(call.Args, " "))
	}
}

func AssertNotCalled(t *testing.T, r *RecordingRunner, args ...string) {
	t.Helper()
	want := strings.Join(args, " ")
	for _, call := range r.Calls {
		if strings.Join(call.Args, " ") == want {
			t.Errorf("unexpected call %q found", want)
			return
		}
	}
}

func TestRecordingRunnerCapturesCalls(t *testing.T) {
	r := NewRecordingRunner()
	ctx := context.Background()

	r.Execute(ctx, "install", "ripgrep")
	r.Execute(ctx, "install", "neovim")
	r.Execute(ctx, "uninstall", "ripgrep")

	if len(r.Calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(r.Calls))
	}

	AssertCalled(t, r, "install", "ripgrep")
	AssertCalled(t, r, "install", "neovim")
	AssertCalled(t, r, "uninstall", "ripgrep")
	AssertNotCalled(t, r, "install", "git")

	if r.Calls[0].At.IsZero() {
		t.Error("expected timestamp to be set")
	}

	// ExecuteJSON delegates to ExecuteFn internally
	var s string
	_ = r.ExecuteJSON(ctx, &s, "info", "ripgrep")
	AssertCalled(t, r, "info", "ripgrep")
}

func TestMockRunnerExecute(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("hello world"), nil
	}
	out, err := r.Execute(context.Background(), "version")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "hello world" {
		t.Errorf("got %q, want %q", string(out), "hello world")
	}
}

func TestMockRunnerExecuteJSON(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		if s, ok := result.(*string); ok {
			*s = "parsed"
		}
		return nil
	}
	var result string
	err := r.ExecuteJSON(context.Background(), &result, "info", "--json=v2", "test")
	if err != nil {
		t.Fatal(err)
	}
	if result != "parsed" {
		t.Errorf("got %q, want %q", result, "parsed")
	}
}

func TestMockRunnerExecuteStream(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		ch := make(chan string, 2)
		errCh := make(chan error, 1)
		ch <- "line1"
		ch <- "line2"
		close(ch)
		close(errCh)
		return ch, errCh
	}
	outChan, errChan := r.ExecuteStream(context.Background(), "install", "test")
	var got []string
	for line := range outChan {
		got = append(got, line)
	}
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "line1" || got[1] != "line2" {
		t.Errorf("got %v, want [line1 line2]", got)
	}
}

func TestMockRunnerBrewPath(t *testing.T) {
	r := NewMockRunner()
	r.BrewPathFn = func() string {
		return "/custom/path/brew"
	}
	if got := r.BrewPath(); got != "/custom/path/brew" {
		t.Errorf("got %q, want /custom/path/brew", got)
	}
}

func TestMockRunnerDefaults(t *testing.T) {
	r := NewMockRunner()
	out, err := r.Execute(context.Background(), "version")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "" {
		t.Errorf("expected empty output, got %q", string(out))
	}
}

func TestMockRunnerBrewPathDefault(t *testing.T) {
	r := NewMockRunner()
	if got := r.BrewPath(); got != "/fake/brew" {
		t.Errorf("got %q, want /fake/brew", got)
	}
}
