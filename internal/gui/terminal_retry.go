package gui

import (
	"errors"
	"os/exec"
	"slices"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
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
		return SudoRunDoneMsg{ID: id, Title: title, Args: args, Err: err}
	})
}
