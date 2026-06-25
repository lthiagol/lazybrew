package flows_test

import (
	"context"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/gui/testutil"
)

type installRecordingRunner struct {
	brew.MockRunner
	mu          sync.Mutex
	streamCalls [][]string
}

func (r *installRecordingRunner) recordStream(args []string) {
	r.mu.Lock()
	r.streamCalls = append(r.streamCalls, args)
	r.mu.Unlock()
}

func (r *installRecordingRunner) getStreamCalls() [][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.streamCalls
}

func newInstallRecordingRunner() *installRecordingRunner {
	rr := &installRecordingRunner{}

	rr.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "search" {
			query := args[len(args)-1]
			out := "==> Formulae\n" + query
			return []byte(out), nil
		}
		return []byte{}, nil
	}

	rr.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
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

func TestInstallFlow(t *testing.T) {
	runner := newInstallRecordingRunner()
	tm := testutil.NewTestModelWithRunner(t, runner, teatest.WithInitialTermSize(120, 40))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	tm.Type("ripgrep")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	time.Sleep(300 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))

	calls := runner.getStreamCalls()
	found := false
	for _, call := range calls {
		if len(call) >= 2 && call[0] == "install" && call[1] == "ripgrep" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected install call for ripgrep, got %v calls", calls)
		for i, call := range calls {
			t.Logf("  call %d: %v", i, call)
		}
	}
}
