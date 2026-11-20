//go:build unix

package brew

import (
	"os/exec"
	"syscall"
	"time"
)

// killGroupDelay bounds how long Wait keeps waiting for a child that exited
// but left its stdout or stderr open (a stray holder of the pipe). Past it,
// Wait closes the pipes and returns, so the stream always ends. Tests shorten it.
var killGroupDelay = 2 * time.Second

// setProcessGroup starts brew in a new session, so it is its own process group
// and has no controlling terminal. A cancelled context kills the whole group,
// so children brew started (curl, git, ruby helpers) die with it and release
// the pipes it holds. Without a controlling terminal, a read of /dev/tty fails
// at once, so sudo fails instead of stopping for a password nobody can type
// (Q-v2-11-b). A process in its own group under the terminal would be stopped.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = killGroupDelay
}
