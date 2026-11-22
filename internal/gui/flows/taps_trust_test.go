package flows_test

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

// trustFlowRunner answers the Taps reads for one tap that trusts one formula. Once
// `brew untrust` has run, the trust list no longer has the formula. Every brew
// call is recorded.
type trustFlowRunner struct {
	brewtest.MockRunner
	mu        sync.Mutex
	untrusted bool
	calls     [][]string
}

func newTrustFlowRunner() *trustFlowRunner {
	r := &trustFlowRunner{}
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		r.record(args)
		switch {
		case args[0] == "tap":
			return []byte("octo/tools\n"), nil
		case args[0] == "tap-info":
			return []byte(`[{"name":"octo/tools","trusted":true}]`), nil
		case args[0] == "trust" && len(args) == 2 && args[1] == "--json=v1":
			if r.isUntrusted() {
				return []byte(`{"taps":["octo/tools"],"formulae":[],"casks":[],"commands":[]}`), nil
			}
			return []byte(`{"taps":["octo/tools"],"formulae":["octo/tools/zeta"],"casks":[],"commands":[]}`), nil
		case args[0] == "untrust":
			r.mu.Lock()
			r.untrusted = true
			r.mu.Unlock()
		}
		return []byte(""), nil
	}
	return r
}

func (r *trustFlowRunner) record(args []string) {
	r.mu.Lock()
	r.calls = append(r.calls, args)
	r.mu.Unlock()
}

func (r *trustFlowRunner) isUntrusted() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.untrusted
}

func (r *trustFlowRunner) getCalls() [][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.calls)
}

// waitForBrewCall waits until the runner has recorded the call args, so a flow
// asserts on the action only after it reached brew.
func waitForBrewCall(t *testing.T, r *trustFlowRunner, args ...string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		for _, call := range r.getCalls() {
			if slices.Equal(call, args) {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("no `brew %s` call, got %v", strings.Join(args, " "), r.getCalls())
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestTapsTrustedItemFlow (lb30 AC-03): on the Taps list a trusted formula is a row
// below its tap. x on it refuses with a toast and runs nothing; t asks, and a yes
// runs `brew untrust --formula`. After the reload the row is gone and the cursor is
// on the tap above it.
func TestTapsTrustedItemFlow(t *testing.T) {
	runner := newTrustFlowRunner()
	tm := testutil.NewTestModelWithRunner(t, runner, teatest.WithInitialTermSize(120, 40))

	tm.Send(tea.KeyPressMsg{Code: '5', Text: "5"})
	waitForScreen(t, tm, "└ ✔ zeta")
	tm.Send(tea.KeyPressMsg{Code: 'j', Text: "j"})

	tm.Send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	waitForScreen(t, tm, "x untaps a tap; t untrusts octo/tools/zeta")

	tm.Send(tea.KeyPressMsg{Code: 't', Text: "t"})
	waitForScreen(t, tm, "Untrust octo/tools/zeta?")
	tm.Send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	waitForBrewCall(t, runner, "untrust", "--formula", "octo/tools/zeta")

	// R reads the taps and the trust list again; the toast says the reload landed.
	tm.Send(tea.KeyPressMsg{Code: 'R', Text: "R"})
	waitForScreen(t, tm, "Data refresh")
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))

	for _, call := range runner.getCalls() {
		if len(call) > 0 && call[0] == "untap" {
			t.Fatalf("x on a trusted item ran brew untap: %v", runner.getCalls())
		}
	}
	// The sidebar row is checked, not the bare name: the operation overlay names
	// the formula too. The item sits under its tap, so its row shows the short name.
	view := ansi.Strip(tm.FinalModel(t).View().Content)
	if strings.Contains(view, "└ ✔ zeta") {
		t.Errorf("the untrusted formula is still on the Taps list:\n%s", view)
	}
	// The tap row leads with the locked mark, which is ✔ in the Unicode set.
	if !strings.Contains(view, "▸ ✔ octo/tools ") {
		t.Errorf("the cursor should land on the tap above the removed row:\n%s", view)
	}
}
