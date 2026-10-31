package gui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew"
)

// errBrewMissing mimics a plain brew failure (brew not installed, tap
// unreadable, ...): an error that is NOT a deadline overrun. The uninstall
// dependency check must keep today's behaviour for these — no dependents
// listed, no complaint.
var errBrewMissing = errors.New("brew: command not found")

// uninstallModel returns a model with one selected formula in the formulae
// panel, so confirmUninstall has a package to check.
func uninstallModel(client *brew.Client) *Model {
	m := newTestModel()
	m.client = client
	m.activePanel = PanelFormulae
	p := m.panels[PanelFormulae]
	p.items = []string{"ripgrep  14.1.1  bottled"}
	p.selected = 0
	return m
}

// TestConfirmUninstallDepCheckTimesOut covers the point of the slice: a hung
// `brew uses` must not produce a confirm dialog that looks exactly like
// "nothing depends on this package". The dialog still opens — the user keeps
// the choice — but it says the check did not complete and that the list may
// be incomplete.
func TestConfirmUninstallDepCheckTimesOut(t *testing.T) {
	orig := brewReadTimeout
	brewReadTimeout = 50 * time.Millisecond
	defer func() { brewReadTimeout = orig }()

	m := uninstallModel(brew.NewClient(blockingRunner()))

	_, cmd := m.confirmUninstall(mutUninstall)
	msg := awaitMsg(t, cmd, 2*time.Second)
	dMsg, ok := msg.(DepCheckMsg)
	if !ok {
		t.Fatalf("expected DepCheckMsg, got %T", msg)
	}

	// The dialog still appears: the message is never empty and the modal is
	// still opened by the DepCheckMsg handler.
	if dMsg.Message == "" {
		t.Fatal("the confirm message must not be empty")
	}
	if !strings.Contains(dMsg.Message, "Dependency check timed out after 50ms") {
		t.Errorf("the message does not report the timeout and its duration:\n%s", dMsg.Message)
	}
	if !strings.Contains(dMsg.Message, "may be incomplete") {
		t.Errorf("the message does not say the list of dependents may be incomplete:\n%s", dMsg.Message)
	}
	// It must NOT present the timed-out read as "nothing depends on it".
	if strings.Contains(dMsg.Message, "The following depend on") {
		t.Errorf("a timed-out check must not list dependents as if the read succeeded:\n%s", dMsg.Message)
	}

	// The modal really opens on the timeout path.
	updated := updateModel(m, dMsg)
	if updated.activeModal == nil {
		t.Error("the confirm modal must still open after a dependency-check timeout")
	}
}

// TestConfirmUninstallDepCheckNonTimeoutError keeps today's behaviour for
// every error that is not a deadline overrun: a DepCheckMsg is still
// produced, with no dependents and no complaint.
func TestConfirmUninstallDepCheckNonTimeoutError(t *testing.T) {
	orig := brewReadTimeout
	brewReadTimeout = 50 * time.Millisecond
	defer func() { brewReadTimeout = orig }()

	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return nil, errBrewMissing
	}

	m := uninstallModel(brew.NewClient(r))

	_, cmd := m.confirmUninstall(mutUninstall)
	msg := awaitMsg(t, cmd, 2*time.Second)
	dMsg, ok := msg.(DepCheckMsg)
	if !ok {
		t.Fatalf("expected DepCheckMsg, got %T", msg)
	}

	if strings.Contains(dMsg.Message, "timed out") {
		t.Errorf("a non-timeout error must not be reported as a timeout:\n%s", dMsg.Message)
	}
	if strings.Contains(dMsg.Message, "The following depend on") {
		t.Errorf("a failed check must not list dependents:\n%s", dMsg.Message)
	}
	if !strings.Contains(dMsg.Message, "Uninstall ripgrep?") {
		t.Errorf("the confirm message lost its original wording:\n%s", dMsg.Message)
	}
}

// TestToggleLeavesTimesOut checks that the leaves read is bounded and that
// the timeout reaches the user through the existing MutationResultMsg error
// path, as a *brew.TimeoutError naming the command and the duration.
func TestToggleLeavesTimesOut(t *testing.T) {
	orig := brewReadTimeout
	brewReadTimeout = 50 * time.Millisecond
	defer func() { brewReadTimeout = orig }()

	m := newTestModel()
	m.client = brew.NewClient(blockingRunner())

	_, cmd := m.toggleLeaves()
	msg := awaitMsg(t, cmd, 2*time.Second)
	rMsg, ok := msg.(MutationResultMsg)
	if !ok {
		t.Fatalf("expected MutationResultMsg, got %T", msg)
	}
	assertTimeout(t, rMsg.Err, "leaves", "50ms")
}
