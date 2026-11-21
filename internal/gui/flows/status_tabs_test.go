package flows_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

// TestStatusTabsFlow: m opens the Missing tab with its rows; l reaches the Vulns
// tab, which shows the hint and runs nothing; v runs the check and shows the
// report; h and l move through the five Status tabs without running it again.
func TestStatusTabsFlow(t *testing.T) {
	var vulns atomic.Int32
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		switch {
		case len(args) >= 1 && args[0] == "missing":
			return []byte("openssl@3: ca-certificates\n"), &brew.BrewExitError{Command: "missing", ExitCode: 1}
		case len(args) >= 1 && args[0] == "config":
			return []byte("HOMEBREW_VERSION: 4.9.0\n"), nil
		case len(args) >= 1 && args[0] == "vulns":
			vulns.Add(1)
			return []byte("CVE-2026-1: openssl@3\n"), nil
		}
		return []byte{}, nil
	}

	tm := testutil.NewTestModelWithRunner(t, r, teatest.WithInitialTermSize(140, 40))
	tm.Send(logsKeyPress('m'))
	waitForScreen(t, tm, "openssl@3: missing ca-certificates")

	tm.Send(logsKeyPress('l'))
	waitForScreen(t, tm, "Press v to check installed formulae")
	if n := vulns.Load(); n != 0 {
		t.Fatalf("opening the Vulns tab ran brew vulns %d times, want 0", n)
	}

	tm.Send(logsKeyPress('v'))
	waitForScreen(t, tm, "CVE-2026-1: openssl@3")

	// From Vulns, l wraps to Dashboard and h walks back through the tabs.
	for _, key := range []rune{'l', 'h', 'h', 'h'} {
		tm.Send(logsKeyPress(key))
	}
	waitForScreen(t, tm, "ready to brew")
	tm.Send(logsKeyPress('h'))
	waitForScreen(t, tm, "4.9.0")

	tm.Send(logsKeyPress('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
	if n := vulns.Load(); n != 1 {
		t.Fatalf("the Vulns check ran %d times across the tab cycle, want 1", n)
	}
}
