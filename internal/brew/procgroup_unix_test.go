//go:build unix

package brew

// Tests for cancel killing brew's whole process group: a child brew started
// must not keep the stream open, and must be gone once the cancel is done.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// writeGrandchildBrew is a fake brew that starts a background child holding
// its stdout and stderr, records the child's pid in pidFile, then waits.
func writeGrandchildBrew(t *testing.T) (*DefaultRunner, string) {
	t.Helper()
	pidFile := filepath.Join(t.TempDir(), "pid")
	r := writeFakeBrew(t, fmt.Sprintf("sleep 30 &\necho $! > %s\nwait\n", pidFile))
	return r, pidFile
}

// waitForPid polls until pidFile holds a pid and returns it.
func waitForPid(t *testing.T, pidFile string) int {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(pidFile); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
				t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
				return pid
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("fake brew never recorded its child pid")
	return 0
}

// waitGone polls until the process no longer exists.
func waitGone(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("child pid %d still alive after cancel", pid)
}

// TestExecuteStreamCancelKillsGrandchild locks AC-01: cancelling a stream
// kills brew's children too, so the stream closes and reports the cancel
// without waiting for the child's sleep.
func TestExecuteStreamCancelKillsGrandchild(t *testing.T) {
	r, pidFile := writeGrandchildBrew(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out, errCh := r.ExecuteStream(ctx, "whatever")

	pid := waitForPid(t, pidFile)
	cancel()

	drained := make(chan struct{})
	go func() {
		for range out {
		}
		close(drained)
	}()
	select {
	case <-drained:
	case <-time.After(5 * time.Second):
		t.Fatal("ExecuteStream did not close after cancel: a child kept the pipe open")
	}

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled after cancel, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ExecuteStream did not report the cancel in time")
	}
	waitGone(t, pid)
}

// TestExecuteCancelKillsGrandchild locks AC-01 for Execute: a cancel returns
// promptly and leaves no child running.
func TestExecuteCancelKillsGrandchild(t *testing.T) {
	r, pidFile := writeGrandchildBrew(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := r.Execute(ctx, "whatever")
		done <- err
	}()

	pid := waitForPid(t, pidFile)
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected an error after cancel, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Execute did not return after cancel: a child kept the pipe open")
	}
	waitGone(t, pid)
}

// TestExecuteStreamStrayHolderDoesNotHang locks the WaitDelay half of AC-01: a
// child that leaves the process group (setsid) keeps the pipes open and
// escapes the group kill, yet the cancelled stream still closes within
// killGroupDelay.
func TestExecuteStreamStrayHolderDoesNotHang(t *testing.T) {
	if _, err := exec.LookPath("setsid"); err != nil {
		t.Skip("setsid not available")
	}
	pidFile := filepath.Join(t.TempDir(), "pid")
	r := writeFakeBrew(t, fmt.Sprintf("setsid sleep 30 &\necho $! > %s\nwait\n", pidFile))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out, errCh := r.ExecuteStream(ctx, "whatever")
	waitForPid(t, pidFile)
	cancel()

	drained := make(chan struct{})
	go func() {
		for range out {
		}
		close(drained)
	}()
	select {
	case <-drained:
	case <-time.After(killGroupDelay + 5*time.Second):
		t.Fatal("ExecuteStream hung on a stray pipe holder")
	}

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled after cancel, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ExecuteStream did not report the cancel in time")
	}
}

// TestExecuteStrayHolderAfterExitIsNotAFailure locks finishErr: brew exits 0
// while a setsid child still holds stdout, so Wait returns ErrWaitDelay after
// killGroupDelay. The command succeeded, so Execute reports no error and keeps
// its output.
func TestExecuteStrayHolderAfterExitIsNotAFailure(t *testing.T) {
	if _, err := exec.LookPath("setsid"); err != nil {
		t.Skip("setsid not available")
	}
	pidFile := filepath.Join(t.TempDir(), "pid")
	r := writeFakeBrew(t, fmt.Sprintf("setsid sleep 30 &\necho $! > %s\necho done\nexit 0\n", pidFile))

	out, err := r.Execute(context.Background(), "whatever")
	waitForPid(t, pidFile)
	if err != nil {
		t.Fatalf("expected no error for a successful brew, got %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "done" {
		t.Fatalf("expected output %q, got %q", "done", got)
	}
}
