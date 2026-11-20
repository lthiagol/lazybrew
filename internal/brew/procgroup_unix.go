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

// setProcessGroup starts brew in its own process group and makes a cancelled
// context kill the whole group, so children brew started (curl, git, ruby
// helpers) die with it and release the pipes it holds.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = killGroupDelay
}
