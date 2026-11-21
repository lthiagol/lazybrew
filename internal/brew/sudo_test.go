package brew

import (
	"context"
	"errors"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// A brew command whose sudo had no terminal names the command the user runs in
// a terminal. The refusal comes as sudo words: stderr of a failed run, or the
// error text of a stream.
func TestSudoWithoutTerminalNamesTheCommand(t *testing.T) {
	cases := map[string]string{
		"terminal required": "sudo: a terminal is required to read the password",
		"no tty":            "sudo: no tty present and no askpass program specified",
		"password required": "sudo: a password is required",
	}
	for name, stderr := range cases {
		t.Run(name, func(t *testing.T) {
			mock := brewtest.NewMockRunner()
			mock.ExecuteFn = func(context.Context, ...string) ([]byte, error) {
				return nil, &BrewExitError{Command: "install --cask foo", ExitCode: 1, Stderr: stderr}
			}
			_, err := NewClient(mock).Runner.Execute(context.Background(), "install", "--cask", "foo")
			var needs *NeedsSudoError
			if !errors.As(err, &needs) {
				t.Fatalf("expected NeedsSudoError, got %T: %v", err, err)
			}
			want := `needs sudo — run "brew install --cask foo" in a terminal`
			if err.Error() != want {
				t.Errorf("message = %q, want %q", err.Error(), want)
			}
			var exit *BrewExitError
			if !errors.As(err, &exit) || exit.ExitCode != 1 {
				t.Error("the command's own error must stay reachable")
			}
		})
	}
}

func TestSudoStreamWithoutTerminalNamesTheCommand(t *testing.T) {
	mock := brewtest.NewMockRunner()
	mock.ExecuteStreamFn = func(context.Context, ...string) (<-chan string, <-chan error) {
		lines := make(chan string)
		close(lines)
		errs := make(chan error, 1)
		errs <- errors.New("brew install --cask foo: sudo: a terminal is required to read the password")
		close(errs)
		return lines, errs
	}
	_, errs := NewClient(mock).Runner.ExecuteStream(context.Background(), "install", "--cask", "foo")
	err := <-errs
	var needs *NeedsSudoError
	if !errors.As(err, &needs) || needs.Command != "install --cask foo" {
		t.Fatalf("expected NeedsSudoError for install --cask foo, got %T: %v", err, err)
	}
}

// Other failures pass through untouched: the same error value comes back.
func TestNonSudoFailureIsUnchanged(t *testing.T) {
	want := &BrewExitError{Command: "install foo", ExitCode: 1, Stderr: "Error: No available formula"}
	mock := brewtest.NewMockRunner()
	mock.ExecuteFn = func(context.Context, ...string) ([]byte, error) { return nil, want }
	_, err := NewClient(mock).Runner.Execute(context.Background(), "install", "foo")
	if err != want {
		t.Fatalf("expected the same error back, got %T: %v", err, err)
	}
}
