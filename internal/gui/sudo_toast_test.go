package gui

import (
	"context"
	"errors"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// failedInstallToast runs a brew install whose stream fails with err and returns
// the error toast the failed task leaves.
func failedInstallToast(t *testing.T, err error) string {
	t.Helper()
	mock := brewtest.NewMockRunner()
	mock.ExecuteStreamFn = func(context.Context, ...string) (<-chan string, <-chan error) {
		lines := make(chan string)
		close(lines)
		errs := make(chan error, 1)
		errs <- err
		close(errs)
		return lines, errs
	}
	m := New(brew.NewClient(mock), testConfig())
	_, errs := m.client.Runner.ExecuteStream(context.Background(), "install", "--cask", "foo")
	return lastToast(updateModel(m, TaskCompletedMsg{ID: "install:foo", Title: "install foo", Err: <-errs}))
}

// A failed task whose sudo had no terminal says to run the command in a terminal
// (Q-v2-11-b-2), with the task's title first, like the other failure toasts.
func TestSudoWithoutTerminalToastSaysRunInATerminal(t *testing.T) {
	got := failedInstallToast(t, errors.New("brew install --cask foo: sudo: a terminal is required to read the password"))
	want := `install foo: needs sudo — run "brew install --cask foo" in a terminal`
	if got != want {
		t.Errorf("toast = %q, want %q", got, want)
	}
}

// Other failures keep their toast: the title and brew's error.
func TestOtherFailureToastIsUnchangedByTheSudoLine(t *testing.T) {
	got := failedInstallToast(t, errors.New("brew install --cask foo: Error: No Cask with this name"))
	want := "install foo: brew install --cask foo: Error: No Cask with this name"
	if got != want {
		t.Errorf("toast = %q, want %q", got, want)
	}
}
