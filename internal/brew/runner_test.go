package brew

import (
	"context"
	"testing"
)

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
