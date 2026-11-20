//go:build linux

package brew

// Tests for brew's terminal: a child that reads /dev/tty must fail at once,
// not stop for a password (Q-v2-11-b). The test needs a controlling terminal,
// so it runs its helper on a pty.

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

const ptyHelperEnv = "LAZYBREW_PTY_HELPER"

// openPty returns a new pty, master and slave. The slave is the terminal a
// helper process gets as its controlling terminal.
func openPty(t *testing.T) (master, slave *os.File) {
	t.Helper()
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("no pty: %v", err)
	}
	t.Cleanup(func() { _ = master.Close() })

	var unlock int32
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&unlock))); errno != 0 {
		t.Skipf("unlock pty: %v", errno)
	}
	var n uint32
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&n))); errno != 0 {
		t.Skipf("pty number: %v", errno)
	}
	slave, err = os.OpenFile("/dev/pts/"+strconv.Itoa(int(n)), os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("open pty slave: %v", err)
	}
	t.Cleanup(func() { _ = slave.Close() })
	return master, slave
}

// TestBrewCannotReadTheTerminal runs a fake brew that reads the terminal, as
// sudo does for a password. The helper runs on a pty with the terminal as its
// controlling terminal, the way a user's terminal would be. The read must fail
// at once, so the stream ends with the read's own answer.
func TestBrewCannotReadTheTerminal(t *testing.T) {
	_, slave := openPty(t)
	result := filepath.Join(t.TempDir(), "result")

	cmd := exec.Command(os.Args[0], "-test.run=^TestPtyHelper$")
	cmd.Env = append(os.Environ(), ptyHelperEnv+"="+result)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = slave, slave, slave
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("helper failed: %v", err)
		}
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("the helper did not finish: brew stopped reading the terminal")
	}

	got, err := os.ReadFile(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "no-terminal" {
		t.Fatalf("the read of the terminal should fail at once, got %q", got)
	}
}

// TestPtyHelper is the helper process of TestBrewCannotReadTheTerminal. It runs
// only when the parent sets ptyHelperEnv, and writes what the stream reported.
func TestPtyHelper(t *testing.T) {
	result := os.Getenv(ptyHelperEnv)
	if result == "" {
		t.Skip("helper process of TestBrewCannotReadTheTerminal")
	}
	r := writeFakeBrew(t, fmt.Sprintf("if read x </dev/tty; then echo terminal; else echo no-terminal; fi\n"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, errCh := r.ExecuteStream(ctx, "whatever")
	var lines []string
	for line := range out {
		lines = append(lines, line)
	}
	outcome := "err: " + fmt.Sprint(<-errCh)
	if len(lines) == 1 && lines[0] == "no-terminal" && ctx.Err() == nil {
		outcome = "no-terminal"
	}
	if err := os.WriteFile(result, []byte(strings.TrimSpace(outcome)), 0o600); err != nil {
		t.Fatal(err)
	}
}
