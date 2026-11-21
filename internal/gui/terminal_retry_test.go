package gui

import (
	"errors"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// sudoFailedModel is a model whose last operation failed because sudo had no
// terminal for the brew command args. Its brew binary is /fake/brew.
func sudoFailedModel(args ...string) *Model {
	mock := brewtest.NewMockRunner()
	mock.BrewPathFn = func() string { return "/fake/brew" }
	m := New(brew.NewClient(mock), testConfig())
	m.opState = &Operation{
		Title:     "install foo",
		TaskID:    "install:foo",
		TaskTitle: "install foo",
		Status:    opError,
		Ended:     time.Now(),
		Err: &brew.NeedsSudoError{
			Command: strings.Join(args, " "),
			Args:    args,
			Err:     errors.New("sudo: a terminal is required to read the password"),
		},
	}
	return m
}

// captureExec swaps the terminal runner of m for one that records the commands
// it is asked to run, and runs none of them.
func captureExec(m *Model) *[]*exec.Cmd {
	var cmds []*exec.Cmd
	m.execProcess = func(cmd *exec.Cmd, _ tea.ExecCallback) tea.Cmd {
		cmds = append(cmds, cmd)
		return nil
	}
	return &cmds
}

var enterKey = tea.KeyPressMsg{Code: tea.KeyEnter}

// Enter on a failed install runs the same brew command attached to the
// terminal: the brew path, the args, the runner's environment, and no session
// of its own, so the terminal is its controlling tty.
func TestEnterRunsTheFailedCommandInTheTerminal(t *testing.T) {
	m := sudoFailedModel("install", "--cask", "foo")
	cmds := captureExec(m)

	m = updateModel(m, enterKey)

	if len(*cmds) != 1 {
		t.Fatalf("Enter ran %d commands in the terminal, want 1", len(*cmds))
	}
	cmd := (*cmds)[0]
	if want := []string{"/fake/brew", "install", "--cask", "foo"}; !slices.Equal(cmd.Args, want) {
		t.Errorf("args = %q, want %q", cmd.Args, want)
	}
	if cmd.Path != "/fake/brew" {
		t.Errorf("path = %q, want /fake/brew", cmd.Path)
	}
	for _, want := range []string{"HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_ASK=1"} {
		if !slices.Contains(cmd.Env, want) {
			t.Errorf("env lacks %s", want)
		}
	}
	if cmd.SysProcAttr != nil {
		t.Errorf("SysProcAttr = %+v, want none: brew's session or process group must not apply", cmd.SysProcAttr)
	}
	if m.opState == nil || m.opState.Status != opError {
		t.Errorf("the failed operation changed: %+v", m.opState)
	}
}

// Only install, reinstall, upgrade and uninstall are run again. Anything else
// that needed sudo keeps its message and Enter does nothing.
func TestEnterDoesNothingForOtherCommands(t *testing.T) {
	for _, args := range [][]string{
		{"update"},
		{"services", "start", "moshi-hook"},
		{"tap", "foo/bar"},
		{"trust", "foo/bar"},
		{"cleanup"},
		{"autoremove"},
		{"zap", "--cask", "foo"},
	} {
		m := sudoFailedModel(args...)
		cmds := captureExec(m)
		m = updateModel(m, enterKey)
		if len(*cmds) != 0 {
			t.Errorf("%q: Enter ran %q in the terminal", args, (*cmds)[0].Args)
		}
		if m.opState == nil {
			t.Errorf("%q: Enter closed the operation", args)
		}
	}
}

// A batch member that failed on sudo runs again with its own args.
func TestEnterRunsTheFailedBatchMemberWithItsArgs(t *testing.T) {
	m := sudoFailedModel("install", "--cask", "bar")
	m.opState.Title = "install bar (2/2 done)"
	m.opState.BatchDone, m.opState.BatchTotal = 1, 2
	cmds := captureExec(m)

	m = updateModel(m, enterKey)

	if len(*cmds) != 1 || !slices.Equal((*cmds)[0].Args, []string{"/fake/brew", "install", "--cask", "bar"}) {
		t.Fatalf("batch member args: got %v, want install --cask bar", *cmds)
	}
}

// The retry runs only when nothing else can: a queued task, or a batch still in
// flight, leaves Enter doing nothing.
func TestEnterDoesNothingWhileTasksAreQueuedOrABatchRuns(t *testing.T) {
	queued := sudoFailedModel("install", "--cask", "foo")
	_, _ = queued.tasks.Enqueue(&task.Task{ID: "next", Title: "next"})

	inBatch := sudoFailedModel("install", "--cask", "foo")
	inBatch.batchCount = 1

	for name, m := range map[string]*Model{"queued task": queued, "batch in flight": inBatch} {
		cmds := captureExec(m)
		m = updateModel(m, enterKey)
		if len(*cmds) != 0 {
			t.Errorf("%s: Enter ran %q in the terminal", name, (*cmds)[0].Args)
		}
		if m.opState == nil {
			t.Errorf("%s: Enter closed the operation", name)
		}
	}
}

// The failed overlay names the retry on its status row, at both widths, only
// when Enter would run it. Esc closes it as before.
func TestFailedOverlayShowsTheRetryOnlyWhenItRuns(t *testing.T) {
	const retry = "Enter  run in this terminal · Esc  close"
	for _, width := range []int{80, 140} {
		m := sudoFailedModel("install", "--cask", "foo")
		if got := ansi.Strip(m.renderOperation(width, 10)); !strings.Contains(got, retry) {
			t.Errorf("width %d: overlay lacks the retry hint:\n%s", width, got)
		}

		refused := sudoFailedModel("services", "start", "moshi-hook")
		if got := ansi.Strip(refused.renderOperation(width, 10)); strings.Contains(got, "run in this terminal") {
			t.Errorf("width %d: refused command shows the retry hint:\n%s", width, got)
		}

		queued := sudoFailedModel("install", "--cask", "foo")
		_, _ = queued.tasks.Enqueue(&task.Task{ID: "next", Title: "next"})
		if got := ansi.Strip(queued.renderOperation(width, 10)); strings.Contains(got, "run in this terminal") {
			t.Errorf("width %d: queued work shows the retry hint:\n%s", width, got)
		}
	}

	m := sudoFailedModel("install", "--cask", "foo")
	m = updateModel(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.opState != nil {
		t.Error("Esc did not close the failed operation")
	}
}
