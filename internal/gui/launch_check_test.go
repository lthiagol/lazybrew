package gui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// launchRunner is the runner of the launch check. `brew --repository` prints
// repo (or fails when repo is empty), `brew update` is counted in updates, and
// `brew outdated` reports one formula and is counted in outdated.
func launchRunner(repo string, updates, outdated *int32) *brewtest.MockRunner {
	r := outdatedCountingRunner(outdated)
	base := r.ExecuteFn
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "--repository" {
			if repo == "" {
				return nil, brewtest.ErrTest
			}
			return []byte(repo + "\n"), nil
		}
		return base(ctx, args...)
	}
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		if len(args) == 1 && args[0] == "update" {
			atomic.AddInt32(updates, 1)
		}
		lines := make(chan string)
		close(lines)
		errs := make(chan error, 1)
		close(errs)
		return lines, errs
	}
	return r
}

// brewRepoWithFetchHead is a brew repository whose last update is at. A zero
// time leaves FETCH_HEAD out, the state of a brew that has never updated.
func brewRepoWithFetchHead(t *testing.T, at time.Time) string {
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

// feedBack feeds the messages of cmd back through Update, and the commands those
// produce, until none are left. The timers are left out: a command still waiting
// after feedWait, the spinner's ticks, and a toast's expiry, which would run its
// whole life as a chain of ticks.
func feedBack(m *Model, cmd tea.Cmd) *Model {
	if cmd == nil {
		return m
	}
	msg, ok := runWithin(cmd, feedWait)
	if !ok {
		return m
	}
	if batch, isBatch := msg.(tea.BatchMsg); isBatch {
		for _, c := range batch {
			m = feedBack(m, c)
		}
		return m
	}
	switch msg.(type) {
	case spinner.TickMsg, modal.ToastTickMsg:
		return m
	}
	nm, next := m.Update(msg)
	return feedBack(asModel(nm), next)
}

// feedWait is how long feedBack waits for a command before it counts it as a
// timer.
const feedWait = 100 * time.Millisecond

// runWithin runs cmd and returns its message, or false when it has not returned
// within wait.
func runWithin(cmd tea.Cmd, wait time.Duration) (tea.Msg, bool) {
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		return msg, true
	case <-time.After(wait):
		return nil, false
	}
}

// launchOf runs the launch: Init, then every message it leads to.
func launchOf(m *Model) *Model {
	return feedBack(m, m.Init())
}

func TestLaunchCheckStaleRunsOneUpdateAndReloadsOutdated(t *testing.T) {
	var updates, outdated int32
	repo := brewRepoWithFetchHead(t, time.Time{})
	m := New(brew.NewClient(launchRunner(repo, &updates, &outdated)), testConfig())

	m = launchOf(m)

	if n := atomic.LoadInt32(&updates); n != 1 {
		t.Fatalf("brew update ran %d times at launch, want 1", n)
	}
	if m.isUpdating {
		t.Error("the update is still marked running after it finished")
	}
	if m.lastUpdate.IsZero() {
		t.Error("the update finished but lastUpdate is still unset")
	}
	// The completion reloads the outdated count, so it is present after the
	// update ends, not left at the count from before it.
	if got := m.status.Outdated; !got.Known || got.N != 1 {
		t.Errorf("Outdated after the update = %+v, want 1 known", got)
	}
	if atomic.LoadInt32(&outdated) == 0 {
		t.Error("the completion did not read brew outdated")
	}
}

func TestLaunchCheckFreshRunsNoUpdateAndShowsItsAge(t *testing.T) {
	var updates, outdated int32
	at := time.Now().Add(-time.Hour)
	repo := brewRepoWithFetchHead(t, at)
	m := updateModel(New(brew.NewClient(launchRunner(repo, &updates, &outdated)), testConfig()),
		tea.WindowSizeMsg{Width: 140, Height: 40})

	m = launchOf(m)

	if n := atomic.LoadInt32(&updates); n != 0 {
		t.Errorf("brew update ran %d times on a fresh brew, want 0", n)
	}
	if n := atomic.LoadInt32(&outdated); n != 0 {
		t.Errorf("a fresh launch read brew outdated %d times, want 0 (M9)", n)
	}
	if d := m.lastUpdate.Sub(at); d > time.Second || d < -time.Second {
		t.Errorf("lastUpdate = %s, want the FETCH_HEAD time %s", m.lastUpdate, at)
	}
	if plain := plainLine(m.renderHeader()); !strings.Contains(plain, "updated 1h ago") {
		t.Errorf("fresh launch header = %q, want it to say updated 1h ago", plain)
	}
}

func TestLaunchCheckUnreadableTimeIsStale(t *testing.T) {
	var updates, outdated int32
	m := New(brew.NewClient(launchRunner("", &updates, &outdated)), testConfig())

	m = launchOf(m)

	if n := atomic.LoadInt32(&updates); n != 1 {
		t.Errorf("brew update ran %d times when the time could not be read, want 1", n)
	}
}

func TestLaunchCheckStaleBadgeReadsUpdatingWhileItRuns(t *testing.T) {
	var updates, outdated int32
	m := updateModel(New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), testConfig()),
		tea.WindowSizeMsg{Width: 140, Height: 40})

	nm, _ := m.Update(lastUpdateMsg{})
	m = asModel(nm)

	if !m.isUpdating {
		t.Fatal("a stale launch check did not start the update")
	}
	if plain := plainLine(m.renderHeader()); !strings.Contains(plain, "updating…") {
		t.Errorf("header while the stale update runs = %q, want updating…", plain)
	}
	if strings.Contains(plainLine(m.renderHeader()), "never updated") {
		t.Error("the stale update still says never updated")
	}
}

func TestLaunchCheckRespectsUpdateStaleHours(t *testing.T) {
	at := time.Now().Add(-3 * time.Hour)
	for _, tt := range []struct {
		hours     int
		wantStale bool
	}{
		{hours: 2, wantStale: true},
		{hours: 6, wantStale: false},
	} {
		var updates, outdated int32
		cfg := testConfig()
		cfg.Brew.UpdateStaleHours = tt.hours
		m := New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, at), &updates, &outdated)), cfg)

		m = launchOf(m)

		got := atomic.LoadInt32(&updates) == 1
		if got != tt.wantStale {
			t.Errorf("3h-old brew with update_stale_hours=%d: update ran=%v, want %v", tt.hours, got, tt.wantStale)
		}
	}
}

// update_on_start: true runs the update at every launch, fresh or not. The
// launch starts it through StartUpdateMsg, which is driven directly: Init's
// update ticker would run for its full interval here.
func TestLaunchCheckAlwaysUpdatesOnEveryLaunch(t *testing.T) {
	var updates, outdated int32
	cfg := testConfig()
	cfg.Brew.UpdateOnStart = config.UpdateAlways
	repo := brewRepoWithFetchHead(t, time.Now().Add(-time.Hour))
	m := New(brew.NewClient(launchRunner(repo, &updates, &outdated)), cfg)

	nm, cmd := m.Update(StartUpdateMsg{})
	feedBack(asModel(nm), cmd)
	if n := atomic.LoadInt32(&updates); n != 1 {
		t.Errorf("update_on_start: true ran brew update %d times on a fresh brew, want 1", n)
	}
}

// update_on_start: false never updates, at launch or from StartUpdateMsg.
func TestLaunchCheckNeverUpdatesAtLaunch(t *testing.T) {
	var updates, outdated int32
	cfg := testConfig()
	cfg.Brew.UpdateOnStart = config.UpdateNever
	m := New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), cfg)

	m = launchOf(m)
	nm, cmd := m.Update(StartUpdateMsg{})
	feedBack(asModel(nm), cmd)
	if n := atomic.LoadInt32(&updates); n != 0 {
		t.Errorf("update_on_start: false ran brew update %d times, want 0", n)
	}
}

func TestLaunchCheckPreloadWaitsForTheCheck(t *testing.T) {
	var updates, outdated int32
	cfg := testConfig()
	cfg.GUI.Preload = true
	at := time.Now().Add(-time.Hour)
	m := New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, at), &updates, &outdated)), cfg)

	// Status lands before the check does: the preload must not start yet.
	nm, _ := m.Update(statusLoaded())
	m = asModel(nm)
	if m.preload.started {
		t.Fatal("the preload started before the launch check had read brew's update time")
	}

	// A fresh check starts it, now that Status has landed.
	nm, cmd := m.Update(lastUpdateMsg{at: at})
	m = asModel(nm)
	if !m.preload.started {
		t.Fatal("a fresh launch check did not start the preload once Status had landed")
	}
	feedBack(m, cmd)
}

func TestLaunchCheckPreloadWaitsForTheUpdate(t *testing.T) {
	var updates, outdated int32
	cfg := testConfig()
	cfg.GUI.Preload = true
	m := New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), cfg)

	nm, cmd := m.Update(lastUpdateMsg{})
	m = asModel(nm)
	// The update is running, and Status lands in the middle of it.
	nm, _ = m.Update(statusLoaded())
	m = asModel(nm)
	if m.preload.started {
		t.Fatal("the preload started while the stale update was still running")
	}

	m = feedBack(m, cmd)
	if !m.preload.started {
		t.Error("the preload did not start once the update ended and the dashboard reloaded")
	}
}

// A launch read from before a stale update can land after the update's refresh
// has loaded the outdated count. It must not overwrite that count.
func TestLaunchReadLandingAfterStaleUpdateIsDropped(t *testing.T) {
	var updates, outdated int32
	m := New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), testConfig())
	nm, cmd := m.Update(lastUpdateMsg{})
	m = feedBack(asModel(nm), cmd)
	if got := m.status.Outdated; !got.Known || got.N != 1 {
		t.Fatalf("Outdated after the update = %+v, want 1 known", got)
	}

	nm, _ = m.Update(DataLoadedMsg{PanelID: PanelStatus, launch: true})
	m = asModel(nm)
	if got := m.status.Outdated; !got.Known || got.N != 1 {
		t.Errorf("a launch read from before the update overwrote the outdated count: %+v", got)
	}
}

// The launch Status read can land while the stale update is still running. It
// is applied then, so the dashboard shows at once rather than "not loaded yet"
// until the update ends.
func TestLaunchStatusLandingDuringStaleUpdateShowsDashboard(t *testing.T) {
	var updates, outdated int32
	m := updateModel(New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), testConfig()),
		tea.WindowSizeMsg{Width: 140, Height: 40})

	// The stale check lands first and starts the update. Its command is not run
	// here, so the update is still running when the Status read lands.
	nm, _ := m.Update(lastUpdateMsg{})
	m = asModel(nm)
	if !m.isUpdating {
		t.Fatal("a stale launch check did not start the update")
	}

	nm, _ = m.Update(launchRead(fetchStatusData(m.client, false))())
	m = asModel(nm)
	if !m.panels[PanelStatus].loaded {
		t.Fatal("the launch Status read was dropped while the stale update ran")
	}
	if out := ansi.Strip(m.renderContent(140, 40)); strings.Contains(out, "not loaded yet") {
		t.Errorf("the dashboard should show during the update, got:\n%s", out)
	}
}

// A stale update that cannot be queued never runs, so nothing reloads after it.
// The launch Status read still applies, and the panel is not left blank until R.
func TestLaunchStatusLandingAfterFailedEnqueueIsApplied(t *testing.T) {
	var updates, outdated int32
	m := updateModel(New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), testConfig()),
		tea.WindowSizeMsg{Width: 140, Height: 40})
	m.tasks = task.NewManager(1)
	// A parked task fills the queue, so the update's enqueue is refused.
	if _, err := m.tasks.Enqueue(&task.Task{ID: "parked"}); err != nil {
		t.Fatal(err)
	}

	nm, cmd := m.Update(lastUpdateMsg{})
	m = feedBack(asModel(nm), cmd)
	if m.isUpdating {
		t.Fatal("a refused update still reads as running")
	}

	nm, _ = m.Update(launchRead(fetchStatusData(m.client, false))())
	m = asModel(nm)
	if !m.panels[PanelStatus].loaded {
		t.Fatal("the launch Status read was dropped after a refused update")
	}
	if out := ansi.Strip(m.renderContent(140, 40)); strings.Contains(out, "not loaded yet") {
		t.Errorf("the dashboard should show after a refused update, got:\n%s", out)
	}
}

// The report behind lb16: brew outdated listed one package before the update and
// two after it, and the count stayed at one. Once the stale update ends, the
// count, the Outdated list and the header all show both packages.
func TestStaleUpdateShowsEveryOutdatedPackage(t *testing.T) {
	var updates, outdated int32
	r := launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)
	base := r.ExecuteFn
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "outdated" {
			atomic.AddInt32(&outdated, 1)
			if atomic.LoadInt32(&updates) == 0 {
				return []byte(`{"formulae":[{"name":"ripgrep","installed_versions":["14.1.0"],"current_version":"14.1.1","pinned":false}]}`), nil
			}
			return []byte(`{"formulae":[{"name":"ripgrep","installed_versions":["14.1.0"],"current_version":"14.1.1","pinned":false},` +
				`{"name":"wget","installed_versions":["1.24.5"],"current_version":"1.25.0","pinned":false}]}`), nil
		}
		return base(ctx, args...)
	}
	m := updateModel(New(brew.NewClient(r), testConfig()), tea.WindowSizeMsg{Width: 140, Height: 40})

	// Before the update brew lists one package. The user has opened the Outdated
	// panel, which loads it lazily, and R reads the count.
	m = feedBack(m, m.lazyLoadPanel(PanelOutdated))
	nm, cmd := m.Update(RefreshMsg{LoadOutdated: true, ReloadDiagnostics: true})
	m = feedBack(asModel(nm), cmd)
	if got := m.status.Outdated; !got.Known || got.N != 1 {
		t.Fatalf("Outdated before the update = %+v, want 1 known", got)
	}
	if n := len(m.panels[PanelOutdated].formulae); n != 1 {
		t.Fatalf("the Outdated list has %d packages before the update, want 1", n)
	}

	nm, cmd = m.Update(lastUpdateMsg{})
	m = feedBack(asModel(nm), cmd)

	if got := m.status.Outdated; !got.Known || got.N != 2 {
		t.Errorf("Outdated after the update = %+v, want 2 known", got)
	}
	if n := len(m.panels[PanelOutdated].formulae); n != 2 {
		t.Errorf("the Outdated list has %d packages after the update, want 2", n)
	}
	if plain := plainLine(m.renderHeader()); !strings.Contains(plain, "2 outdated") {
		t.Errorf("header after the update = %q, want it to say 2 outdated", plain)
	}
}

// R is an explicit request, so under auto it runs brew update too, as it does
// under true (Q-lb16-r-key-auto).
func TestRKeyUnderAutoRunsOneUpdateAndLoadsOutdated(t *testing.T) {
	var updates, outdated int32
	repo := brewRepoWithFetchHead(t, time.Now().Add(-time.Hour))
	m := New(brew.NewClient(launchRunner(repo, &updates, &outdated)), testConfig())

	nm, cmd := m.Update(tea.KeyPressMsg{Code: 'R', Text: "R"})
	m = feedBack(asModel(nm), cmd)

	if n := atomic.LoadInt32(&updates); n != 1 {
		t.Fatalf("R under auto ran brew update %d times, want 1", n)
	}
	if m.isUpdating {
		t.Error("the update is still marked running after it finished")
	}
	if got := m.status.Outdated; !got.Known || got.N != 1 {
		t.Errorf("Outdated after R's update = %+v, want 1 known", got)
	}
}

// update_on_start: false keeps R to a reload: no brew update, the outdated count
// is still read.
func TestRKeyUnderNeverOnlyReloads(t *testing.T) {
	var updates, outdated int32
	cfg := testConfig()
	cfg.Brew.UpdateOnStart = config.UpdateNever
	m := New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), cfg)

	nm, cmd := m.Update(tea.KeyPressMsg{Code: 'R', Text: "R"})
	m = feedBack(asModel(nm), cmd)

	if n := atomic.LoadInt32(&updates); n != 0 {
		t.Errorf("R under never ran brew update %d times, want 0", n)
	}
	if n := atomic.LoadInt32(&outdated); n == 0 {
		t.Error("R under never did not read brew outdated")
	}
}

// R while an update is already running enqueues no second one.
func TestRKeyWhileUpdatingEnqueuesNothing(t *testing.T) {
	var updates, outdated int32
	repo := brewRepoWithFetchHead(t, time.Time{})
	m := New(brew.NewClient(launchRunner(repo, &updates, &outdated)), testConfig())
	m.isUpdating = true

	nm, cmd := m.Update(tea.KeyPressMsg{Code: 'R', Text: "R"})
	m = feedBack(asModel(nm), cmd)

	if n := atomic.LoadInt32(&updates); n != 0 {
		t.Errorf("R while an update runs enqueued %d brew updates, want 0", n)
	}
	if !m.isUpdating {
		t.Error("R while an update runs cleared isUpdating")
	}
}
