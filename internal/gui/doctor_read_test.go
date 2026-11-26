package gui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// LB-45: brew doctor is its own read, so the Status dashboard lands without it.

func dashView(t *testing.T, m *Model) string {
	t.Helper()
	return ansi.Strip(m.renderDashboard(140))
}

// TestDashboardLandsBeforeDoctor: with a doctor that blocks, the dashboard read
// still returns, its counts draw, and Health shows doctor running.
func TestDashboardLandsBeforeDoctor(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	r := newStatusMockRunner(func(ctx context.Context, args ...string) ([]byte, error) {
		select {
		case <-release:
		case <-ctx.Done():
		}
		return nil, ctx.Err()
	})
	client := brew.NewClient(r)

	done := make(chan tea.Msg, 1)
	go func() { done <- fetchStatusData(client, false)() }()
	var msg tea.Msg
	select {
	case msg = <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the dashboard read waited for brew doctor")
	}

	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, msg)
	out := dashView(t, m)
	if !strings.Contains(out, "Formulae:  0 installed") {
		t.Fatalf("the dashboard counts should draw before doctor lands, got:\n%s", out)
	}
	if !strings.Contains(out, "running doctor…") {
		t.Fatalf("Health should show doctor running, got:\n%s", out)
	}
	if strings.Contains(out, "press d to run doctor") {
		t.Fatalf("a pending doctor must not ask for a run, got:\n%s", out)
	}
}

// TestDoctorLandingFillsHealth: the doctor read fills the Health card, and a
// later dashboard landing keeps it.
func TestDoctorLandingFillsHealth(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, doctorLoadedMsg{gen: m.doctorGen, warnings: 2})
	if out := dashView(t, m); !strings.Contains(out, "2 warnings") || strings.Contains(out, "running doctor") {
		t.Fatalf("Health should show the doctor's 2 warnings, got:\n%s", out)
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus, Summary: dashboardSummary(), Items: statusRows(dashboardSummary())})
	if out := dashView(t, m); !strings.Contains(out, "2 warnings") {
		t.Fatalf("a dashboard landing wiped the doctor result, got:\n%s", out)
	}
}

// TestDoctorFailureReadsUnavailable: a failed or timed-out doctor read shows
// as unavailable with its reason.
func TestDoctorFailureReadsUnavailable(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, doctorLoadedMsg{gen: m.doctorGen, err: errors.New("brew doctor timed out after 1m0s")})
	words := strings.Join(strings.Fields(ansi.Strip(m.healthBody())), " ")
	if !strings.Contains(words, "doctor unavailable: brew doctor timed out") {
		t.Fatalf("a timed-out doctor should read unavailable, got:\n%s", words)
	}
}

// TestStaleDoctorDoesNotOverwrite: a doctor result from an earlier read than
// the latest issued is not applied, before or after the newer one lands.
func TestStaleDoctorDoesNotOverwrite(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.cfg.GUI.AutoRefreshSeconds = 0
	m = updateModel(m, RefreshMsg{})
	if m.doctorGen != 1 {
		t.Fatalf("a refresh should issue doctor read generation 1, got %d", m.doctorGen)
	}
	m = updateModel(m, doctorLoadedMsg{gen: 0, launch: true, warnings: 5})
	if out := dashView(t, m); strings.Contains(out, "5 warnings") || !strings.Contains(out, "running doctor…") {
		t.Fatalf("a stale doctor read was applied, got:\n%s", out)
	}
	m = updateModel(m, doctorLoadedMsg{gen: 1, refresh: true})
	m = updateModel(m, doctorLoadedMsg{gen: 0, launch: true, warnings: 5})
	if out := dashView(t, m); !strings.Contains(out, "ready to brew") {
		t.Fatalf("the newer doctor result should stay, got:\n%s", out)
	}
}

// TestLaunchDoctorDroppedAfterUpdateRefresh: like the launch dashboard read, a
// launch doctor read that lands after a stale update's refresh is dropped.
func TestLaunchDoctorDroppedAfterUpdateRefresh(t *testing.T) {
	m := newTestModel()
	m.launchStale, m.updateRefreshed = true, true
	m = updateModel(m, doctorLoadedMsg{gen: m.doctorGen, launch: true, warnings: 3})
	if !m.doctorPending || m.status.Doctor.Known {
		t.Fatal("a launch doctor read after the update's refresh was applied")
	}
}

// TestPreloadStartsWithoutDoctor: the preload starts on the Status landing,
// while the doctor read is still out.
func TestPreloadStartsWithoutDoctor(t *testing.T) {
	m := newTestModel()
	m.cfg.GUI.Preload = true
	m = updateModel(m, statusLoaded())
	if !m.doctorPending {
		t.Fatal("doctor should still be pending")
	}
	if !m.preload.started {
		t.Fatal("the preload should start on the Status landing without doctor")
	}
}

// TestRefreshToastWaitsForDoctor: "Data refreshed" fires only once the
// refresh's doctor read has landed too.
func TestRefreshToastWaitsForDoctor(t *testing.T) {
	m := newTestModel()
	m.cfg.GUI.AutoRefreshSeconds = 0
	m = updateModel(m, RefreshMsg{})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus, refresh: true})
	if lastToast(m) == "Data refreshed" {
		t.Fatal("\"Data refreshed\" toasted before doctor landed")
	}
	m = updateModel(m, doctorLoadedMsg{gen: m.doctorGen, refresh: true})
	if lastToast(m) != "Data refreshed" {
		t.Fatalf("last toast = %q after doctor landed, want \"Data refreshed\"", lastToast(m))
	}
}
