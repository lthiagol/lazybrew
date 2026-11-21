package gui

import (
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// terminalRetryCommands are the brew commands whose sudo step a failed
// operation runs again in the terminal. Every other command keeps its message.
var terminalRetryCommands = []string{"install", "reinstall", "upgrade", "uninstall"}

// terminalRetry is the sudo refusal the failed operation can run again in the
// terminal, when it has one: the command is one of terminalRetryCommands, no
// task is running or queued, and no batch is in flight. Then the TUI can
// suspend without a write of lazybrew's in progress.
func (m Model) terminalRetry() (*brew.NeedsSudoError, bool) {
	if m.opState == nil || m.opState.Status != opError || m.batchCount > 0 || !m.tasks.Idle() {
		return nil, false
	}
	var needs *brew.NeedsSudoError
	if !errors.As(m.opState.Err, &needs) || len(needs.Args) == 0 {
		return nil, false
	}
	if !slices.Contains(terminalRetryCommands, needs.Args[0]) {
		return nil, false
	}
	return needs, true
}

// retryInTerminal runs the failed brew command with the TUI suspended, attached
// to the terminal so sudo can ask for its password. It runs outside the task
// manager on purpose: terminalRetry has checked that nothing else can run.
// brew keeps the environment of every brew command but not its own session,
// which is what gives sudo the terminal.
func (m Model) retryInTerminal(needs *brew.NeedsSudoError) (tea.Model, tea.Cmd) {
	cmd := exec.Command(m.client.Runner.BrewPath(), needs.Args...)
	cmd.Env = brew.CommandEnv()
	id, title, args := m.opState.TaskID, m.opState.TaskTitle, needs.Args
	return m, m.execProcess(cmd, func(err error) tea.Msg {
		// A failed command is its exit status, as the toast names it.
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			err = fmt.Errorf("exited with status %d", exit.ExitCode())
		}
		return SudoRunDoneMsg{ID: id, Title: title, Args: args, Err: err}
	})
}

// sudoRetryDone is the end of a retry in the terminal. As for every write, the
// cache group the command writes is dropped, success or not. A success ends the
// operation as a task's success does. A failure leaves the failed operation as
// it was, so Esc still closes it, and says the exit status in a toast, as a
// failed task does.
func (m Model) sudoRetryDone(msg SudoRunDoneMsg) (tea.Model, tea.Cmd) {
	if len(msg.Args) > 0 {
		m.invalidateCacheGroup(msg.Args[0])
	}
	if msg.Err != nil {
		return m, m.setToast(failureText(msg.ID, msg.Title, msg.Err), modal.ToastError)
	}
	if m.opState != nil {
		m.opState.Status = opSuccess
		m.opState.Ended = time.Now()
	}
	return m.finishedMutation(msg.ID, msg.Title)
}
