package gui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// An uninstall runs again in the terminal, as an install does.
func TestEnterRunsAnUninstallInTheTerminal(t *testing.T) {
	m := sudoFailedModel("uninstall", "--cask", "foo")
	cmds := captureExec(m)

	updateModel(m, enterKey)

	if len(*cmds) != 1 || !slices.Equal((*cmds)[0].Args, []string{"/fake/brew", "uninstall", "--cask", "foo"}) {
		t.Fatalf("uninstall: got %v, want uninstall --cask foo", *cmds)
	}
}

// Only install, reinstall, upgrade and uninstall are run again. Anything else
// that needed sudo keeps its message and Enter does nothing. A cask zap is an
// uninstall with --zap, which deletes user data, so it is refused too.
func TestEnterDoesNothingForOtherCommands(t *testing.T) {
	for _, args := range [][]string{
		{"update"},
		{"services", "start", "moshi-hook"},
		{"tap", "foo/bar"},
		{"trust", "foo/bar"},
		{"cleanup"},
		{"autoremove"},
		{"uninstall", "--zap", "--cask", "foo"},
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

	updateModel(m, enterKey)

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

// The retry runs only from a failed operation. One still running, or cancelled,
// that carries a sudo refusal keeps Enter and its status row as they are.
func TestEnterDoesNothingUnlessTheOperationFailed(t *testing.T) {
	for _, status := range []opStatus{opRunning, opCancelled} {
		m := sudoFailedModel("install", "--cask", "foo")
		m.opState.Status = status
		m.opState.Started = time.Now()
		cmds := captureExec(m)

		m = updateModel(m, enterKey)

		if len(*cmds) != 0 {
			t.Errorf("status %v: Enter ran %q in the terminal", status, (*cmds)[0].Args)
		}
		if got := ansi.Strip(m.renderOperation(80, 10)); strings.Contains(got, "run in this terminal") {
			t.Errorf("status %v: overlay shows the retry hint:\n%s", status, got)
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

// A retry that succeeds ends the operation as a task's success does: the
// toast, the refresh, and the cache group the command writes is dropped.
func TestRetrySuccessFinishesLikeATaskAndDropsItsCache(t *testing.T) {
	m := sudoFailedModel("install", "--cask", "foo")
	m.client.Cache.SetWithTTL(brew.KeyCasksList, []brew.Cask{}, time.Minute)

	next, cmd := m.Update(SudoRunDoneMsg{ID: "install:foo", Title: "install foo", Args: []string{"install", "--cask", "foo"}})
	updated := next.(Model)
	m = &updated

	if m.opState.Status != opSuccess {
		t.Errorf("status = %v, want opSuccess", m.opState.Status)
	}
	if got := lastToast(m); got != "install foo completed" {
		t.Errorf("toast = %q, want %q", got, "install foo completed")
	}
	if _, ok := m.client.Cache.Get(brew.KeyCasksList); ok {
		t.Error("the install's cache group was not dropped")
	}
	if n := len(collectRefreshMsgs(cmd)); n != 1 {
		t.Errorf("refresh messages = %d, want 1", n)
	}
}

// A retry that fails keeps the failed operation as it was, and its toast names
// the exit status, as a failed task's toast names its error.
func TestRetryFailureKeepsTheOperationAndToastsTheExit(t *testing.T) {
	m := sudoFailedModel("install", "--cask", "foo")

	next, _ := m.Update(SudoRunDoneMsg{
		ID: "install:foo", Title: "install foo", Args: []string{"install", "--cask", "foo"},
		Err: fmt.Errorf("exited with status 3"),
	})
	updated := next.(Model)
	m = &updated

	if m.opState == nil || m.opState.Status != opError {
		t.Fatalf("the failed operation changed: %+v", m.opState)
	}
	var needs *brew.NeedsSudoError
	if !errors.As(m.opState.Err, &needs) {
		t.Errorf("the operation's error = %v, want the sudo error kept", m.opState.Err)
	}
	if got := lastToast(m); got != "install foo: exited with status 3" {
		t.Errorf("toast = %q, want the exit status", got)
	}
}

// The callback of the terminal run maps a non-zero exit to its status.
func TestRetryCallbackNamesTheExitStatus(t *testing.T) {
	m := sudoFailedModel("install", "--cask", "foo")
	var cb tea.ExecCallback
	m.execProcess = func(_ *exec.Cmd, fn tea.ExecCallback) tea.Cmd {
		cb = fn
		return nil
	}
	updateModel(m, enterKey)
	if cb == nil {
		t.Fatal("Enter did not hand the command to the terminal runner")
	}

	exitErr := exec.Command("sh", "-c", "exit 3").Run()
	failed, ok := cb(exitErr).(SudoRunDoneMsg)
	if !ok {
		t.Fatalf("callback returned %T, want SudoRunDoneMsg", cb(exitErr))
	}
	if failed.Err == nil || failed.Err.Error() != "exited with status 3" {
		t.Errorf("err = %v, want exited with status 3", failed.Err)
	}
	if clean := cb(nil).(SudoRunDoneMsg); clean.Err != nil {
		t.Errorf("a clean exit has err %v", clean.Err)
	}
}

// ptyHelperEnv marks the run of the pty test that happens inside the pseudo-
// terminal. The parent run starts the helper under script(1).
const ptyHelperEnv = "LB23_PTY_HELPER"

// fakeBrewScript stands in for brew in the pty test. It records whether it has
// a controlling terminal, which sudo needs, and exits with LB23_FAKE_BREW_EXIT.
const fakeBrewScript = `#!/bin/sh
if (exec 3</dev/tty) 2>/dev/null; then echo tty > "$LB23_MARKER"; else echo no-tty > "$LB23_MARKER"; fi
exit "$LB23_FAKE_BREW_EXIT"
`

// TestSudoRetryRunsInTheTerminalUnderAPty runs the retry for real: the brew
// command the production code builds, run with cmd.Run in a pseudo-terminal
// that script(1) provides, with its exit status coming back as SudoRunDoneMsg.
// The tea.ExecProcess wrapper is replaced by a plain cmd.Run here, because a
// Program is not running; that wrapper only sets stdio and calls Run. Skipped
// when script(1) is absent.
func TestSudoRetryRunsInTheTerminalUnderAPty(t *testing.T) {
	if os.Getenv(ptyHelperEnv) != "" {
		retryUnderPty(t)
		return
	}
	scriptBin, err := exec.LookPath("script")
	if err != nil {
		t.Skip("script(1) is not installed: the pty check needs it")
	}
	dir := t.TempDir()
	fake := filepath.Join(dir, "brew")
	if err := os.WriteFile(fake, []byte(fakeBrewScript), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "marker")
	result := filepath.Join(dir, "result.json")

	self := os.Args[0]
	runTest := "-test.run=^TestSudoRetryRunsInTheTerminalUnderAPty$"
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command(scriptBin, "-q", "/dev/null", self, runTest)
	} else {
		cmd = exec.Command(scriptBin, "-qec", "'"+self+"' "+runTest, "/dev/null")
	}
	cmd.Env = append(os.Environ(), ptyHelperEnv+"=1", "LB23_FAKE_BREW="+fake,
		"LB23_MARKER="+marker, "LB23_RESULT="+result, "LB23_FAKE_BREW_EXIT=3")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pty run failed: %v\n%s", err, out)
	}

	gotMarker, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(gotMarker)); got != "tty" {
		t.Errorf("brew ran with %q, want a controlling terminal (tty)", got)
	}
	raw, err := os.ReadFile(result)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Toast  string
		Status opStatus
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Toast != "install foo: exited with status 3" {
		t.Errorf("toast = %q, want the exit status", got.Toast)
	}
	if got.Status != opError {
		t.Errorf("status = %v, want opError", got.Status)
	}
}

// retryUnderPty is the pty run's body: the failed install is retried with the
// production command, and the outcome is written for the parent to read.
func retryUnderPty(t *testing.T) {
	runner, err := brew.NewDefaultRunnerWithPath(os.Getenv("LB23_FAKE_BREW"))
	if err != nil {
		t.Fatal(err)
	}
	m := New(brew.NewClient(runner), testConfig())
	m.opState = &Operation{
		Title: "install foo", TaskID: "install:foo", TaskTitle: "install foo",
		Status: opError, Ended: time.Now(),
		Err: &brew.NeedsSudoError{Command: "install --cask foo", Args: []string{"install", "--cask", "foo"}},
	}
	m.execProcess = func(cmd *exec.Cmd, fn tea.ExecCallback) tea.Cmd {
		return func() tea.Msg {
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			return fn(cmd.Run())
		}
	}

	next, cmd := m.Update(enterKey)
	next, _ = next.(Model).Update(cmd())
	final := next.(Model)
	out, err := json.Marshal(struct {
		Toast  string
		Status opStatus
	}{lastToast(&final), final.opState.Status})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("LB23_RESULT"), out, 0o600); err != nil {
		t.Fatal(err)
	}
}
