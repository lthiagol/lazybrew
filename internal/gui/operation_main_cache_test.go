package gui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// overFilesTabOutput starts an operation over the Files tab, streams one line
// of it and returns the model. The Files rows must not show while it runs.
func overFilesTabOutput(m *Model) *Model {
	m = startOp(m, "t1", "Install bash")
	return updateModel(m, TaskOutputMsg{ID: "t1", Lines: []string{"==> Pouring bash"}})
}

// TestOperationOverCachedFilesTabShowsTheOverlay (cr3 F-r1-01 a): an operation
// started over a Files tab that is already cached shows the overlay for its
// whole run, failed included, and the Files rows are back once it closes.
func TestOperationOverCachedFilesTabShowsTheOverlay(t *testing.T) {
	m := filesTabModel(t, 300)
	if out := screenText(m); !strings.Contains(out, "file0001.txt") {
		t.Fatalf("the Files rows should show before the operation, got:\n%s", out)
	}
	m = overFilesTabOutput(m)
	if out := screenText(m); strings.Contains(out, "file0001.txt") || !strings.Contains(out, "Pouring bash") {
		t.Fatalf("the overlay should show while the operation runs, got:\n%s", screenText(m))
	}
	m = finishOp(m, "t1", "Install bash", errors.New("boom"))
	if out := screenText(m); strings.Contains(out, "file0001.txt") || !strings.Contains(out, "Failed") {
		t.Fatalf("the overlay should show the failed run, got:\n%s", screenText(m))
	}
	m = sendSpecial(m, tea.KeyEsc)
	if out := screenText(m); !strings.Contains(out, "file0001.txt") || strings.Contains(out, "Pouring bash") {
		t.Fatalf("the Files rows should be back once the operation closes, got:\n%s", out)
	}
}

// TestOperationOverUncachedFilesTabDoesNotFreezeItsFrame (cr3 F-r1-01 b): an
// operation started before the Files tab has been wrapped must not leave its
// first frame behind for the Files rows to be replaced by after it closes.
func TestOperationOverUncachedFilesTabDoesNotFreezeItsFrame(t *testing.T) {
	m := filesTabModel(t, 300)
	m = overFilesTabOutput(m)
	if out := screenText(m); !strings.Contains(out, "Pouring bash") {
		t.Fatalf("the overlay should show while the operation runs, got:\n%s", out)
	}
	m = finishOp(m, "t1", "Install bash", errors.New("boom"))
	m = sendSpecial(m, tea.KeyEsc)
	if out := screenText(m); !strings.Contains(out, "file0001.txt") || strings.Contains(out, "Pouring bash") {
		t.Fatalf("the Files rows should be back once the operation closes, got:\n%s", out)
	}
}

// TestLoadingOverCachedFilesTabShowsLoading (cr3 F-r1-01 c): a panel that starts
// loading over a cached Files tab shows the loading state, not the old rows.
func TestLoadingOverCachedFilesTabShowsLoading(t *testing.T) {
	m := filesTabModel(t, 300)
	screenText(m)
	m.panels[PanelFormulae].loading = true
	if out := screenText(m); !strings.Contains(out, "Loading...") || strings.Contains(out, "file0001.txt") {
		t.Fatalf("the loading state should show over the cached tab, got:\n%s", out)
	}
	m.panels[PanelFormulae].loading = false
	if out := screenText(m); !strings.Contains(out, "file0001.txt") {
		t.Fatalf("the Files rows should be back once loading ends, got:\n%s", out)
	}
}

// TestMainRowsDuringOperationAreTheOverlayRows (cr3 F-r1-01): while an operation
// covers the Files tab, the rows the main area scrolls through (and G measures)
// are the overlay's rows, not the Files rows.
func TestMainRowsDuringOperationAreTheOverlayRows(t *testing.T) {
	m := filesTabModel(t, 300)
	screenText(m)
	m = overFilesTabOutput(m)
	w := contentWidthIn(m.screenLayout())
	want := wrapContent(m.renderContent(w, 30), w)
	if got := m.mainRows(w, 30); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("the rows during the operation should be the overlay's, got:\n%s", strings.Join(got, "\n"))
	}
}
