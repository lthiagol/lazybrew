package flows_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

// previewRunner answers a dry run with listing until the real run has been
// streamed, and with nothing after it, as brew does once there is nothing left.
// It records the real runs' arguments.
type previewRunner struct {
	command string // the brew command whose dry run is answered: cleanup or autoremove
	listing string
	mu      sync.Mutex
	streams [][]string
	ran     atomic.Bool
}

func (p *previewRunner) runner() *brewtest.MockRunner {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) == 2 && args[0] == p.command && args[1] == "-n" && !p.ran.Load() {
			return []byte(p.listing), nil
		}
		return []byte{}, nil
	}
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		p.mu.Lock()
		p.streams = append(p.streams, args)
		p.mu.Unlock()
		p.ran.Store(true)
		ch := make(chan string)
		close(ch)
		errCh := make(chan error, 1)
		errCh <- nil
		close(errCh)
		return ch, errCh
	}
	return r
}

// TestCleanupTabFlow: c opens the Cleanup tab with brew's listing; c again asks to
// run it; y runs `brew cleanup` once, and the tab then reads brew again.
func TestCleanupTabFlow(t *testing.T) {
	p := &previewRunner{
		command: "cleanup",
		listing: "Would remove: /home/linuxbrew/.linuxbrew/Cellar/old/1.0 (3 files, 230.3MB)\n==> This operation would free approximately 230.3MB of disk space.\n",
	}
	tm := testutil.NewTestModelWithRunner(t, p.runner(), teatest.WithInitialTermSize(140, 40))

	tm.Send(tea.KeyPressMsg{Code: 'c', Text: "c"})
	waitForScreenAll(t, tm, "Would free ~230.3MB", "Press c again to run brew cleanup")
	tm.Send(tea.KeyPressMsg{Code: 'c', Text: "c"})
	waitForScreen(t, tm, "Run brew cleanup, freeing about 230.3MB?")
	tm.Send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	waitForScreen(t, tm, "Nothing to clean up")

	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.streams) != 1 || len(p.streams[0]) != 1 || p.streams[0][0] != "cleanup" {
		t.Fatalf("real runs = %v, want one `brew cleanup`", p.streams)
	}
}

// TestAutoremoveTabFlow: A opens the Unused tab with the formulae; A again asks to
// uninstall them; y runs `brew autoremove` once, and the tab then reads brew again.
func TestAutoremoveTabFlow(t *testing.T) {
	p := &previewRunner{
		command: "autoremove",
		listing: "==> Would autoremove 2 unneeded formulae:\ncmake\nopenssl@3\n",
	}
	tm := testutil.NewTestModelWithRunner(t, p.runner(), teatest.WithInitialTermSize(140, 40))

	tm.Send(tea.KeyPressMsg{Code: 'A', Text: "A"})
	waitForScreenAll(t, tm, "2 unused formulae would be removed", "Press A again to run brew autoremove")
	tm.Send(tea.KeyPressMsg{Code: 'A', Text: "A"})
	waitForScreen(t, tm, "Uninstall 2 unused formulae?")
	tm.Send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	waitForScreen(t, tm, "Nothing unused")

	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.streams) != 1 || len(p.streams[0]) != 1 || p.streams[0][0] != "autoremove" {
		t.Fatalf("real runs = %v, want one `brew autoremove`", p.streams)
	}
}
