package gui

import (
	"errors"
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// untapModel is the Taps panel with user/tap selected, and the Formulae and
// Casks lists loaded with the given installed packages.
func untapModel(formulae []brew.Formula, casks []brew.Cask) *Model {
	m := newTestModel()
	m.activePanel = PanelTaps
	m.panels[PanelTaps].items = []string{"user/tap  3 formulae"}
	m.panels[PanelTaps].taps = []brew.Tap{{Name: "user/tap"}}
	m.panels[PanelTaps].setTapRows()
	m.panels[PanelTaps].selected = 0
	m.panels[PanelFormulae].formulae = formulae
	m.panels[PanelCasks].casks = casks
	return m
}

// confirmView is the untap confirm as its words, in reading order: the colour
// codes go, and so does the box border, which a wrapped row carries as ┃.
func confirmView(t *testing.T, m *Model) string {
	t.Helper()
	if _, ok := m.activeModal.(*modal.ConfirmModal); !ok {
		t.Fatalf("expected the untap confirm, got %T", m.activeModal)
	}
	view := sgr.ReplaceAllString(m.activeModal.View().Content, "")
	return strings.Join(strings.Fields(strings.ReplaceAll(view, "┃", " ")), " ")
}

// TestUntapConfirmSaysWhatItDoes: with no package installed from the tap, the
// confirm names the untap and the consequence, and says nothing of installed
// packages.
func TestUntapConfirmSaysWhatItDoes(t *testing.T) {
	m := untapModel(nil, nil)
	nm, _ := m.confirmUntap()
	mm := nm.(Model)
	m = &mm
	c := m.activeModal.(*modal.ConfirmModal)
	if c.Action() != "Untap" || c.Kind() != modal.ActionDanger {
		t.Errorf("confirm = %q kind %v, want the Untap chip in the danger kind", c.Action(), c.Kind())
	}
	if c.Detail() != "The tap's formulae and casks are no longer available." {
		t.Errorf("detail = %q", c.Detail())
	}
	if view := confirmView(t, m); strings.Contains(view, "refuse") {
		t.Errorf("no package is installed from the tap, yet the confirm mentions them:\n%s", view)
	}
}

// TestUntapConfirmListsInstalledFromTheTap: the installed packages whose full
// name is under user/tap/ are listed, and only those; brew's refusal is named.
func TestUntapConfirmListsInstalledFromTheTap(t *testing.T) {
	formulae := []brew.Formula{
		{Name: "foo", FullName: "user/tap/foo"},
		{Name: "bar", FullName: "other/tap/bar"},
		{Name: "wget", FullName: "wget"},
	}
	casks := []brew.Cask{{Name: "baz", FullName: "user/tap/baz"}}
	m := untapModel(formulae, casks)

	nm, _ := m.confirmUntap()
	mm := nm.(Model)
	m = &mm
	view := confirmView(t, m)
	for _, want := range []string{"Untap user/tap?", "brew will refuse to untap while they are installed", "foo", "baz"} {
		if !strings.Contains(view, want) {
			t.Errorf("confirm lacks %q:\n%s", want, view)
		}
	}
	for _, not := range []string{"bar", "wget"} {
		if strings.Contains(view, not) {
			t.Errorf("confirm lists %q, which is not from user/tap:\n%s", not, view)
		}
	}
}

// TestUntapConfirmCountsPackagesTheLeavesFilterHides: the Formulae list narrowed
// to leaves still counts a package from the tap that the filter hides.
func TestUntapConfirmCountsPackagesTheLeavesFilterHides(t *testing.T) {
	m := untapModel(nil, nil)
	p := m.panels[PanelFormulae]
	p.leavesActive = true
	p.formulae = nil
	p.unfilteredFormulae = []brew.Formula{{Name: "dep", FullName: "user/tap/dep"}}

	nm, _ := m.confirmUntap()
	mm := nm.(Model)
	if view := confirmView(t, &mm); !strings.Contains(view, "dep") {
		t.Fatalf("the hidden package from the tap should be listed:\n%s", view)
	}
}

// TestUntapConfirmMakesNoBrewCall: the installed list comes from what is loaded.
func TestUntapConfirmMakesNoBrewCall(t *testing.T) {
	m := untapModel(nil, nil)
	m.confirmUntap()
	if !noTaskQueued(m) {
		t.Fatal("the untap confirm queued a task")
	}
}

// A failed untap says brew's line, not its exit code.
func TestUntapFailureSaysBrewsLine(t *testing.T) {
	m := untapModel(nil, nil)
	stderr := "Warning: something minor\nError: Refusing to untap user/tap because foo is installed.\n"
	m = updateModel(m, TaskCompletedMsg{
		ID: "untap:user/tap", Title: "Untap user/tap",
		Err: &brew.BrewExitError{Command: "untap user/tap", ExitCode: 1, Stderr: stderr},
	})
	want := "Untap user/tap: Refusing to untap user/tap because foo is installed."
	if got := lastToast(m); got != want {
		t.Fatalf("toast = %q, want %q", got, want)
	}
}

// A failure with no stderr falls back to the error text.
func TestTapFailureWithoutStderrKeepsTheError(t *testing.T) {
	got := tapFailureText("Untrust: user/tool", errors.New("boom"))
	if got != "Untrust: user/tool: boom" {
		t.Fatalf("got %q", got)
	}
}

// Other failures keep their toast: a services write and a package upgrade are
// not taps writes.
func TestNonTapFailureTextIsUnchanged(t *testing.T) {
	err := &brew.BrewExitError{Command: "upgrade bash", ExitCode: 1, Stderr: "Error: nope"}
	if got := failureText("upgrade-bash", "Upgrade bash", err); got != "Upgrade bash: "+err.Error() {
		t.Fatalf("non-tap failure text changed: %q", got)
	}
}

// TestUntapConfirmListsInstalledCaskFromBrewPayload (LB-17): a cask that user/tap
// owns, read from brew's installed payload (full_token, no full_name), is named
// in the untap confirm.
func TestUntapConfirmListsInstalledCaskFromBrewPayload(t *testing.T) {
	payload := `{"casks":[{"token":"gcloud","full_token":"user/tap/gcloud","tap":"user/tap","name":["gcloud"],"installed":"1.0"}]}`
	_, casks, err := brew.ParseInfo([]byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	m := untapModel(nil, casks)

	nm, _ := m.confirmUntap()
	mm := nm.(Model)
	view := confirmView(t, &mm)
	for _, want := range []string{"brew will refuse to untap while they are installed", "gcloud"} {
		if !strings.Contains(view, want) {
			t.Errorf("confirm lacks %q:\n%s", want, view)
		}
	}
}
