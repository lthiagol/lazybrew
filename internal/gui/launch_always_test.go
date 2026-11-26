package gui

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// alwaysConfig is testConfig under update_on_start: true, the mode whose launch
// update runs at every launch, beside the launch read.
func alwaysConfig() *config.Config {
	cfg := testConfig()
	cfg.Brew.UpdateOnStart = config.UpdateAlways
	return cfg
}

// failingUpdateRunner is launchRunner with a brew update that fails, as an
// offline launch's does.
func failingUpdateRunner(repo string, updates, outdated *int32) *brewtest.MockRunner {
	r := launchRunner(repo, updates, outdated)
	readOnly := r.ExecuteStreamFn
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		if len(args) != 1 || args[0] != "update" {
			return readOnly(ctx, args...)
		}
		atomic.AddInt32(updates, 1)
		lines := make(chan string)
		close(lines)
		errs := make(chan error, 1)
		errs <- brewtest.ErrTest
		close(errs)
		return lines, errs
	}
	return r
}

// Under update_on_start: true, Init starts the launch read and the update
// together. The Status dashboard shows while the update still runs.
func TestAlwaysLaunchStatusLoadsWhileTheUpdateRuns(t *testing.T) {
	var updates, outdated int32
	m := updateModel(New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), alwaysConfig()),
		tea.WindowSizeMsg{Width: 140, Height: 40})

	var launch, start tea.Msg
	for _, msg := range batchMsgs(m.Init()) {
		switch v := msg.(type) {
		case DataLoadedMsg:
			if v.launch && v.PanelID == PanelStatus {
				launch = v
			}
		case StartUpdateMsg:
			start = v
		}
	}
	if launch == nil {
		t.Fatal("Init issues no Status launch read")
	}
	if start == nil {
		t.Fatal("Init does not start the update")
	}

	// The update's command is not run here, so the update is still running
	// when the launch read lands.
	m = updateModel(m, start)
	if !m.isUpdating {
		t.Fatal("StartUpdateMsg did not start the update")
	}
	m = updateModel(m, launch)
	if !m.isUpdating {
		t.Error("the launch read ended the update")
	}
	if !m.panels[PanelStatus].loaded {
		t.Fatal("the launch Status read was not applied while the update runs")
	}
	if out := ansi.Strip(m.renderContent(140, 40)); strings.Contains(out, "not loaded yet") {
		t.Errorf("the dashboard should show during the update, got:\n%s", out)
	}
}

// A launch read from before the update can land after the update's refresh,
// which reads the panels again. It must not overwrite the count the refresh loaded.
func TestAlwaysLaunchReadLandingAfterUpdateIsDropped(t *testing.T) {
	var updates, outdated int32
	m := New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), alwaysConfig())
	nm, cmd := m.Update(StartUpdateMsg{})
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

// A failed update still refreshes after it ends, as a successful one does. A
// launch read landing while it runs is applied; one landing after that refresh
// is dropped, so it cannot overwrite the count the refresh loaded.
func TestAlwaysLaunchReadAroundFailedUpdate(t *testing.T) {
	var updates, outdated int32
	m := New(brew.NewClient(failingUpdateRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), alwaysConfig())
	nm, cmd := m.Update(StartUpdateMsg{})
	m = asModel(nm)

	// The update's command is not run yet, so the update is still running.
	nm, _ = m.Update(launchStatusWithOutdated(3))
	m = asModel(nm)
	if got := m.status.Outdated; !got.Known || got.N != 3 {
		t.Fatalf("a launch read landing during the failing update was dropped; Outdated = %+v", got)
	}

	m = feedBack(m, cmd)
	if n := atomic.LoadInt32(&updates); n != 1 {
		t.Fatalf("the failed launch ran brew update %d times, want 1", n)
	}
	if m.isUpdating {
		t.Fatal("the failed update is still marked running")
	}
	if got := m.status.Outdated; !got.Known || got.N != 1 {
		t.Fatalf("Outdated after the failed update's refresh = %+v, want 1 known", got)
	}

	nm, _ = m.Update(launchStatusWithOutdated(3))
	m = asModel(nm)
	if got := m.status.Outdated; !got.Known || got.N != 1 {
		t.Errorf("a launch read after the failed update's refresh overwrote the count: %+v", got)
	}
}

// launchStatusWithOutdated is a launch Status read that reports n outdated.
func launchStatusWithOutdated(n int) DataLoadedMsg {
	return DataLoadedMsg{
		PanelID: PanelStatus,
		launch:  true,
		Summary: statusSummary{Outdated: knownCount{Known: true, N: n}},
	}
}

// Under update_on_start: true the preload waits for the launch update. A Status
// landing before or during the update does not start it. The Status load that
// follows the update does.
func TestAlwaysPreloadWaitsForTheLaunchUpdate(t *testing.T) {
	var updates, outdated int32
	cfg := alwaysConfig()
	cfg.GUI.Preload = true
	m := New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), cfg)

	// The launch Status read lands before the update starts.
	nm, _ := m.Update(statusLoaded())
	m = asModel(nm)
	if m.preload.started {
		t.Fatal("the preload started before the launch update")
	}

	nm, cmd := m.Update(StartUpdateMsg{})
	m = asModel(nm)
	nm, _ = m.Update(statusLoaded())
	m = asModel(nm)
	if m.preload.started {
		t.Fatal("the preload started while the launch update was running")
	}

	m = feedBack(m, cmd)
	if !m.preload.started {
		t.Error("the preload did not start once the update ended and the dashboard reloaded")
	}
}

// A failed launch update settles too, so the preload starts after the Status
// load that follows it instead of waiting forever.
func TestAlwaysPreloadStartsAfterAFailedUpdate(t *testing.T) {
	var updates, outdated int32
	cfg := alwaysConfig()
	cfg.GUI.Preload = true
	m := New(brew.NewClient(failingUpdateRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), cfg)

	nm, cmd := m.Update(StartUpdateMsg{})
	m = feedBack(asModel(nm), cmd)
	if m.isUpdating {
		t.Fatal("the failed update is still marked running")
	}
	if !m.preload.started {
		t.Error("the preload did not start after the failed update's Status load")
	}
}

// A refused launch update settles after the launch Status read has landed. No
// load follows it, so the refusal itself starts the preload.
func TestAlwaysPreloadStartsWhenARefusedUpdateFollowsStatus(t *testing.T) {
	var updates, outdated int32
	cfg := alwaysConfig()
	cfg.GUI.Preload = true
	m := updateModel(New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), cfg),
		tea.WindowSizeMsg{Width: 140, Height: 40})
	m.tasks = task.NewManager(1)
	if _, err := m.tasks.Enqueue(&task.Task{ID: "parked"}); err != nil {
		t.Fatal(err)
	}

	nm, _ := m.Update(statusLoaded())
	m = asModel(nm)
	if m.preload.started {
		t.Fatal("the preload started before the launch update settled")
	}

	nm, _ = m.Update(StartUpdateMsg{})
	m = asModel(nm)
	if m.isUpdating {
		t.Fatal("a refused update still reads as running")
	}
	if !m.preload.started {
		t.Error("the refused update left the preload waiting for an update that never runs")
	}
}

// A refused launch update never runs, so nothing reloads after it. The preload
// starts on the Status load that lands next, without waiting for an update.
func TestAlwaysPreloadStartsAfterARefusedUpdate(t *testing.T) {
	var updates, outdated int32
	cfg := alwaysConfig()
	cfg.GUI.Preload = true
	m := updateModel(New(brew.NewClient(launchRunner(brewRepoWithFetchHead(t, time.Time{}), &updates, &outdated)), cfg),
		tea.WindowSizeMsg{Width: 140, Height: 40})
	m.tasks = task.NewManager(1)
	// A parked task fills the queue, so the update's enqueue is refused.
	if _, err := m.tasks.Enqueue(&task.Task{ID: "parked"}); err != nil {
		t.Fatal(err)
	}

	nm, cmd := m.Update(StartUpdateMsg{})
	m = feedBack(asModel(nm), cmd)
	if m.isUpdating {
		t.Fatal("a refused update still reads as running")
	}

	nm, cmd = m.Update(statusLoaded())
	m = feedBack(asModel(nm), cmd)
	if !m.preload.started {
		t.Error("the preload waits for a refused update that never runs")
	}
}
