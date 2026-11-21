package flows_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

type uninstallRecordingRunner struct {
	brewtest.MockRunner
	mu          sync.Mutex
	streamCalls [][]string
}

func (r *uninstallRecordingRunner) recordStream(args []string) {
	r.mu.Lock()
	r.streamCalls = append(r.streamCalls, args)
	r.mu.Unlock()
}

func (r *uninstallRecordingRunner) getStreamCalls() [][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.streamCalls
}

func newUninstallRecordingRunner() *uninstallRecordingRunner {
	rr := &uninstallRecordingRunner{}

	rr.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		if len(args) >= 2 && args[0] == "info" {
			type formulaeJSON struct {
				Formulae []struct {
					Name     string `json:"name"`
					FullName string `json:"full_name"`
					Desc     string `json:"desc"`
					Tap      string `json:"tap"`
				} `json:"formulae"`
			}
			data := formulaeJSON{
				Formulae: []struct {
					Name     string `json:"name"`
					FullName string `json:"full_name"`
					Desc     string `json:"desc"`
					Tap      string `json:"tap"`
				}{
					{Name: "testformula", FullName: "testformula", Desc: "Test formula", Tap: "homebrew/core"},
				},
			}
			raw, _ := json.Marshal(data)
			return json.Unmarshal(raw, result)
		}
		return nil
	}

	rr.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		rr.recordStream(args)
		ch := make(chan string, 1)
		errCh := make(chan error, 1)
		close(ch)
		close(errCh)
		return ch, errCh
	}

	return rr
}

func TestUninstallFlow(t *testing.T) {
	runner := newUninstallRecordingRunner()
	tm := testutil.NewTestModelWithRunner(t, runner, teatest.WithInitialTermSize(120, 40))

	tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
	waitForScreen(t, tm, "▸ ✔ testformula")
	tm.Send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	waitForScreen(t, tm, "Uninstall testformula?")
	tm.Send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	waitForCall(t, runner.getStreamCalls, "uninstall", "testformula")
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))

	calls := runner.getStreamCalls()
	found := false
	for _, call := range calls {
		if len(call) >= 2 && call[0] == "uninstall" && call[1] == "testformula" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected uninstall call for testformula, got %v calls", calls)
		for i, call := range calls {
			t.Logf("  call %d: %v", i, call)
		}
	}
}
