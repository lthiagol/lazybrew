package gui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// errBrewKilled mimics what the brew runner reports when exec.CommandContext
// kills the brew process on the deadline: the returned error is NOT
// context.DeadlineExceeded. A timeout must therefore be detected from the
// context's own error, never from the type of the error the runner returned —
// sniffing the returned error would silently never fire.
var errBrewKilled = errors.New("signal: killed")

// blockingRunner returns a mock runner whose every invocation blocks until its
// context is done and then fails the way a killed brew process fails.
func blockingRunner() *brew.MockRunner {
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		<-ctx.Done()
		return nil, errBrewKilled
	}
	return r
}

// awaitMsg runs cmd and waits at most d for its message. A test that can hang
// is worse than no test, so every wait here is bounded.
func awaitMsg(t *testing.T, cmd tea.Cmd, d time.Duration) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatal("cmd is nil")
	}
	done := make(chan tea.Msg, 1)
	go func() {
		done <- cmd()
	}()
	select {
	case msg := <-done:
		return msg
	case <-time.After(d):
		t.Fatalf("cmd hung past %s", d)
		return nil
	}
}

// assertTimeout checks that err is a *brew.TimeoutError (not a raw context
// error) and that its message names the brew command and the duration the
// user is told about.
func assertTimeout(t *testing.T, err error, command string, duration string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	var toErr *brew.TimeoutError
	if !errors.As(err, &toErr) {
		t.Fatalf("expected *brew.TimeoutError, got %T: %v", err, err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected the timeout to be reported as a *brew.TimeoutError, not the raw context error: %v", err)
	}
	if toErr.Command != command {
		t.Errorf("timeout command = %q, want %q", toErr.Command, command)
	}
	if !strings.Contains(err.Error(), "brew "+command+" timed out after "+duration) {
		t.Errorf("timeout message %q does not name the command and the duration", err.Error())
	}
}

func TestFetchPanelDataTimesOut(t *testing.T) {
	orig := brewReadTimeout
	brewReadTimeout = 50 * time.Millisecond
	defer func() { brewReadTimeout = orig }()

	client := brew.NewClient(blockingRunner())

	msg := awaitMsg(t, fetchPanelData(client, PanelFormulae), 2*time.Second)
	dMsg, ok := msg.(DataLoadedMsg)
	if !ok {
		t.Fatalf("expected DataLoadedMsg, got %T", msg)
	}
	assertTimeout(t, dMsg.Err, "list", "50ms")
}

func TestFetchStatusDataTimesOut(t *testing.T) {
	origRead := brewReadTimeout
	origTotal := brewStatusTotalTimeout
	brewReadTimeout = 50 * time.Millisecond
	brewStatusTotalTimeout = 500 * time.Millisecond
	defer func() {
		brewReadTimeout = origRead
		brewStatusTotalTimeout = origTotal
	}()

	client := brew.NewClient(blockingRunner())

	// The user-visible path: the dashboard carries the timeout as a ⚠ line.
	msg := awaitMsg(t, fetchStatusData(client, false), 2*time.Second)
	dMsg, ok := msg.(DataLoadedMsg)
	if !ok {
		t.Fatalf("expected DataLoadedMsg, got %T", msg)
	}
	joined := strings.Join(dMsg.Items, "\n")
	if !strings.Contains(joined, "brew list timed out after 50ms") {
		t.Errorf("dashboard does not report the timed-out read:\n%s", joined)
	}

	// The typed error behind that line.
	items, errs := statusDashboard(client, false)
	if len(errs) == 0 {
		t.Fatal("expected the dashboard to aggregate the timed-out reads")
	}
	assertTimeout(t, errs[0], "list", "50ms")
	if !strings.Contains(strings.Join(items, "\n"), "brew list timed out after 50ms") {
		t.Error("dashboard items do not name the command and the duration")
	}
}

func TestFetchStatusDataRespectsTotalBudget(t *testing.T) {
	origRead := brewReadTimeout
	origTotal := brewStatusTotalTimeout
	// Tiny per-read caps and a total budget below the sum of those caps: the
	// dashboard must resolve on the budget, not after every read ran to its
	// own cap.
	brewReadTimeout = 100 * time.Millisecond
	brewStatusTotalTimeout = 200 * time.Millisecond
	defer func() {
		brewReadTimeout = origRead
		brewStatusTotalTimeout = origTotal
	}()

	client := brew.NewClient(blockingRunner())

	start := time.Now()
	msg := awaitMsg(t, fetchStatusData(client, false), 5*time.Second)
	elapsed := time.Since(start)

	dMsg, ok := msg.(DataLoadedMsg)
	if !ok {
		t.Fatalf("expected DataLoadedMsg, got %T", msg)
	}
	if len(dMsg.Items) == 0 {
		t.Error("expected the dashboard to render despite every read hanging")
	}

	if elapsed > brewStatusTotalTimeout+100*time.Millisecond {
		t.Errorf("fetchStatusData took %s, past the %s total budget", elapsed, brewStatusTotalTimeout)
	}
	if elapsed >= 5*brewReadTimeout {
		t.Errorf("fetchStatusData took %s — the sum of the per-read caps, not the total budget", elapsed)
	}

	_, errs := statusDashboard(client, false)
	if len(errs) == 0 {
		t.Fatal("expected the dashboard to aggregate the timed-out reads")
	}
	for i, err := range errs {
		var toErr *brew.TimeoutError
		if !errors.As(err, &toErr) {
			t.Errorf("errs[%d] = %T: %v, want *brew.TimeoutError", i, err, err)
		}
	}
}

func TestFetchStatusDataTotalBudgetBoundsTheAggregate(t *testing.T) {
	origRead := brewReadTimeout
	origTotal := brewStatusTotalTimeout
	// A per-read cap wider than the total budget: only the total budget can
	// end the wait, so it is the bound that must bind.
	brewReadTimeout = 2 * time.Second
	brewStatusTotalTimeout = 150 * time.Millisecond
	defer func() {
		brewReadTimeout = origRead
		brewStatusTotalTimeout = origTotal
	}()

	client := brew.NewClient(blockingRunner())

	start := time.Now()
	msg := awaitMsg(t, fetchStatusData(client, false), 5*time.Second)
	elapsed := time.Since(start)

	if _, ok := msg.(DataLoadedMsg); !ok {
		t.Fatalf("expected DataLoadedMsg, got %T", msg)
	}
	if elapsed > brewStatusTotalTimeout+200*time.Millisecond {
		t.Errorf("fetchStatusData took %s, past the %s total budget", elapsed, brewStatusTotalTimeout)
	}
	// The reads waited for the total budget, so the budget — not the wider
	// per-read cap — is what ended them.
	if elapsed < brewStatusTotalTimeout*8/10 {
		t.Errorf("fetchStatusData took %s — the reads ended well before the %s total budget, so the budget was not the binding bound", elapsed, brewStatusTotalTimeout)
	}
}

func TestExecuteSearchTimesOut(t *testing.T) {
	orig := brewSearchTimeout
	brewSearchTimeout = 50 * time.Millisecond
	defer func() { brewSearchTimeout = orig }()

	m := newTestModel()
	m.client = brew.NewClient(blockingRunner())

	msg := awaitMsg(t, m.executeSearch("ripgrep"), 2*time.Second)
	sMsg, ok := msg.(SearchDoneMsg)
	if !ok {
		t.Fatalf("expected SearchDoneMsg, got %T", msg)
	}
	assertTimeout(t, sMsg.Err, "search", "50ms")
}

func TestFetchSelectedSearchInfoTimesOut(t *testing.T) {
	orig := brewSearchTimeout
	brewSearchTimeout = 50 * time.Millisecond
	defer func() { brewSearchTimeout = orig }()

	m := newTestModel()
	m.client = brew.NewClient(blockingRunner())
	m.activePanel = PanelSearch
	m.searchResults = []brew.SearchResult{{Name: "ripgrep"}}

	msg := awaitMsg(t, m.fetchSelectedSearchInfo(), 2*time.Second)
	iMsg, ok := msg.(SearchInfoLoadedMsg)
	if !ok {
		t.Fatalf("expected SearchInfoLoadedMsg, got %T", msg)
	}
	assertTimeout(t, iMsg.Err, "info", "50ms")
}
