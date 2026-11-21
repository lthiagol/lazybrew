package flows_test

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

// launchCheckRunner is the runner of the launch check: `brew --repository`
// prints repo, `brew update` is counted in updates, and `brew outdated` reports
// one formula until the update has run, then two: the update brings in a
// package brew had not listed before.
func launchCheckRunner(repo string, updates *atomic.Int32) *brewtest.MockRunner {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		switch {
		case len(args) == 1 && args[0] == "--repository":
			return []byte(repo + "\n"), nil
		case len(args) >= 1 && args[0] == "outdated":
			if updates.Load() == 0 {
				return []byte(`{"formulae":[{"name":"ripgrep","installed_versions":["14.1.0"],"current_version":"14.1.1","pinned":false}]}`), nil
			}
			return []byte(`{"formulae":[{"name":"ripgrep","installed_versions":["14.1.0"],"current_version":"14.1.1","pinned":false},` +
				`{"name":"wget","installed_versions":["1.24.5"],"current_version":"1.25.0","pinned":false}]}`), nil
		}
		return []byte{}, nil
	}
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		if len(args) == 1 && args[0] == "update" {
			updates.Add(1)
		}
		lines := make(chan string)
		close(lines)
		errs := make(chan error, 1)
		close(errs)
		return lines, errs
	}
	return r
}

// launchRepo is a brew repository whose last update is at. A zero time leaves
// FETCH_HEAD out: a brew that has never updated.
func launchRepo(t *testing.T, at time.Time) string {
	t.Helper()
	repo := t.TempDir()
	if at.IsZero() {
		return repo
	}
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	fetchHead := filepath.Join(repo, ".git", "FETCH_HEAD")
	if err := os.WriteFile(fetchHead, []byte("abc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(fetchHead, at, at); err != nil {
		t.Fatal(err)
	}
	return repo
}

// A brew that was last updated long ago is updated at launch. The outdated
// count, which the update changed, is read again once the update ends, and it
// shows on the header. The Outdated panel lists the package the update brought in.
func TestLaunchCheckStaleFlow(t *testing.T) {
	var updates atomic.Int32
	tm := testutil.NewTestModelWithRunner(t, launchCheckRunner(launchRepo(t, time.Time{}), &updates),
		teatest.WithInitialTermSize(140, 40))

	// One wait, not two: a wait consumes the output it reads, so a text that
	// arrives in the same frame as an earlier one is never seen by a later wait.
	// The count of two only appears once the update has ended, so it is the one to wait for.
	waitForScreen(t, tm, "2 outdated")
	tm.Send(tea.KeyPressMsg{Code: '4', Text: "4"})
	waitForScreen(t, tm, "wget")
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))

	if n := updates.Load(); n != 1 {
		t.Errorf("brew update ran %d times at launch, want 1", n)
	}
}

// A brew updated within update_stale_hours is left alone at launch: no update
// runs, and the header says when it last did.
func TestLaunchCheckFreshFlow(t *testing.T) {
	var updates atomic.Int32
	tm := testutil.NewTestModelWithRunner(t, launchCheckRunner(launchRepo(t, time.Now().Add(-time.Hour)), &updates),
		teatest.WithInitialTermSize(140, 40))

	waitForScreen(t, tm, "updated 1h ago")
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))

	if n := updates.Load(); n != 0 {
		t.Errorf("brew update ran %d times on a fresh brew, want 0", n)
	}
}
