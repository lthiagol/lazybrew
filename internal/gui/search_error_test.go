package gui

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// TestSearchFailureShowsErrorToast (LB-20 a): when one typed read finds nothing
// and the other fails, the failure is the search's error. It shows as an error
// toast and on the Search panel, not as an empty list that reads as no results.
func TestSearchFailureShowsErrorToast(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if args[1] == "--cask" {
			return nil, brewtest.ErrTest
		}
		return nil, &brew.BrewExitError{Command: "search", ExitCode: 1}
	}
	m := newTestModel()
	m.client = brew.NewClient(r)

	m = updateModel(m, awaitMsg(t, m.executeSearch("x"), 5*time.Second))

	if !errors.Is(m.panels[PanelSearch].err, brewtest.ErrTest) {
		t.Errorf("Search panel error = %v, want the failed read's error", m.panels[PanelSearch].err)
	}
	var failure bool
	for _, toast := range m.toasts {
		failure = failure || (toast.Message == "Search: "+brewtest.ErrTest.Error() && toast.Type == modal.ToastError)
	}
	if !failure {
		t.Errorf("a failed search shows no error toast; toasts = %+v", m.toasts)
	}
}
