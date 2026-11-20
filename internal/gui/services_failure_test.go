package gui

import (
	"context"
	"errors"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// failedServiceToast restarts moshi-hook through the confirm, with brew failing
// with runErr, and returns the toast the failure leaves.
func failedServiceToast(t *testing.T, runErr error) string {
	t.Helper()
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return nil, runErr
	}
	m := New(brew.NewClient(r), config.Default())
	m.switchPanel(PanelServices)
	p := m.panels[PanelServices]
	p.services = []brew.Service{runningService("moshi-hook")}
	p.items = []string{"moshi-hook  started"}
	p.selected = 0

	m = sendKey(m, "r")
	m.activeModal = nil
	res, cmd := m.handleModalResult(&modal.ConfirmResult{Confirmed: true}, nil)
	nm := res.(Model)
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected TaskStartedMsg from the confirmed restart")
	}
	comp, ok := nm.tasks.RunNext()().(TaskCompletedMsg)
	if !ok || comp.Err == nil {
		t.Fatalf("expected a failed TaskCompletedMsg, got %+v", comp)
	}
	return lastToast(updateModel(&nm, comp))
}

// A restart brew refuses for want of root says what to run, in full (AC-04.1).
func TestServiceNeedsSudoToastNamesTheCommand(t *testing.T) {
	stderr := "Error: Service `moshi-hook` is started as `root`. Try:\n  sudo brew services restart moshi-hook"
	got := failedServiceToast(t, &brew.BrewExitError{Command: "services restart", ExitCode: 1, Stderr: stderr})
	want := `moshi-hook: needs sudo — run "sudo brew services restart moshi-hook" in a terminal`
	if got != want {
		t.Errorf("toast = %q, want %q", got, want)
	}
}

// Other failures keep the toast they had: the task title and brew's error
// (AC-04.2).
func TestServiceOtherFailureToastIsUnchanged(t *testing.T) {
	got := failedServiceToast(t, errors.New("brew services restart failed"))
	want := "restart moshi-hook: brew services restart failed"
	if got != want {
		t.Errorf("toast = %q, want %q", got, want)
	}
}
