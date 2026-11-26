package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/lthiagol/lazybrew/internal/brew"
)

// selTarget is the target of mutType on m's selected row; it fails the test
// when no row is selected.
func selTarget(t testing.TB, m *Model, mutType mutationType, label string) mutTarget {
	t.Helper()
	tg, ok := m.selectedTarget(mutType, label)
	if !ok {
		t.Fatalf("no row selected for %s", label)
	}
	return tg
}

// formulaeModel is a model on a loaded Formulae panel listing names, cursor on
// the first, whose writes r records.
func formulaeModel(r *searchUninstallRunner, names ...string) *Model {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 120, Height: 40})
	m.client = brew.NewClient(r.mock())
	m.activePanel = PanelFormulae
	p := m.panels[PanelFormulae]
	p.loaded = true
	p.formulae, p.items = formulaRows(names...)
	p.selected = 0
	return m
}

func formulaRows(names ...string) ([]brew.Formula, []string) {
	fs := make([]brew.Formula, len(names))
	for i, n := range names {
		fs[i] = brew.Formula{Name: n}
	}
	return fs, append([]string(nil), names...)
}

func wantWrites(t *testing.T, r *searchUninstallRunner, want string) {
	t.Helper()
	got := r.writesSeen()
	if len(got) != 1 || strings.Join(got[0], " ") != want {
		t.Fatalf("writes = %v, want [%s]", got, want)
	}
}

// TestUninstallActsOnTheNamedPackageAfterCursorMoves (LB-47 FB2): x, then j
// while brew reads the dependents; the dialog names foo and the confirm
// uninstalls foo, not the row the cursor moved to.
func TestUninstallActsOnTheNamedPackageAfterCursorMoves(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{"foo": true, "bar": true}}
	m := formulaeModel(r, "foo", "bar")

	m, cmd := pressKeyOn(m, "x")
	if cmd == nil {
		t.Fatal("x started no dependency check")
	}
	m = sendKey(m, "j")
	if got := m.panels[PanelFormulae].selectedName(); got != "bar" {
		t.Fatalf("cursor = %q, want bar", got)
	}
	m = updateModel(m, cmd())
	if m.activeModal == nil || !strings.Contains(m.View().Content, "Uninstall foo?") {
		t.Fatalf("dialog does not name foo:\n%s", m.View().Content)
	}
	confirmAndRun(t, m)
	wantWrites(t, r, "uninstall foo")
}

// TestConfirmActsOnTheNamedPackageAfterReload (LB-47 FB2, FB7): a reload that
// inserts a row above the target while the dialog is open keeps both the
// cursor and the confirmed upgrade on foo.
func TestConfirmActsOnTheNamedPackageAfterReload(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{}}
	m := formulaeModel(r, "foo", "zed")

	m = sendKey(m, "u")
	if m.activeModal == nil {
		t.Fatal("u opened no confirm")
	}
	fs, items := formulaRows("aaa", "foo", "zed")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Formulae: fs, Items: items})
	if got := m.panels[PanelFormulae].selectedName(); got != "foo" {
		t.Fatalf("cursor after reload = %q, want foo", got)
	}
	confirmAndRun(t, m)
	wantWrites(t, r, "upgrade foo")
}

// TestOutdatedReloadKeepsCursorOnSameKind (LB-47 FB7): on Outdated a formula
// and a cask can share a name; the cursor stays on the cask row.
func TestOutdatedReloadKeepsCursorOnSameKind(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelOutdated
	p := m.panels[PanelOutdated]
	p.loaded = true
	p.formulae = []brew.Formula{{Name: "dup"}}
	p.casks = []brew.Cask{{Name: "dup"}}
	p.items = []string{"dup", "dup"}
	p.selected = 1

	m = updateModel(m, DataLoadedMsg{
		PanelID:  PanelOutdated,
		Formulae: []brew.Formula{{Name: "aaa"}, {Name: "dup"}},
		Casks:    []brew.Cask{{Name: "dup"}},
		Items:    []string{"aaa", "dup", "dup"},
	})
	if got := m.panels[PanelOutdated].selected; got != 2 {
		t.Fatalf("selected = %d, want 2 (the cask row)", got)
	}
}

// TestDepCheckDroppedUnderAnotherDialog (LB-47 FB2): a dependency check that
// lands while another dialog is open does not replace it.
func TestDepCheckDroppedUnderAnotherDialog(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{"foo": true}}
	m := formulaeModel(r, "foo")

	m, cmd := pressKeyOn(m, "x")
	m = sendKey(m, "r")
	if m.activeModal == nil || !strings.Contains(m.View().Content, "Reinstall foo?") {
		t.Fatalf("r opened no reinstall dialog:\n%s", m.View().Content)
	}
	m = updateModel(m, cmd())
	if !strings.Contains(m.View().Content, "Reinstall foo?") || strings.Contains(m.View().Content, "Uninstall foo?") {
		t.Fatalf("dependency check replaced the open dialog:\n%s", m.View().Content)
	}
	confirmAndRun(t, m)
	wantWrites(t, r, "reinstall foo")
}

// TestSearchLandingLeavesOtherPanel (LB-47 FB5): results landing while the
// user is on Formulae are stored without switching to Search.
func TestSearchLandingLeavesOtherPanel(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	rows := []brew.SearchResult{{Name: "jq", IsFormula: true}}
	m = updateModel(m, SearchDoneMsg{Results: []string{"jq"}, Raw: rows})
	if m.activePanel != PanelFormulae {
		t.Fatalf("activePanel = %v, want Formulae", m.activePanel)
	}
	if len(m.searchResults) != 1 || !m.panels[PanelSearch].loaded {
		t.Fatalf("results not stored: %v", m.searchResults)
	}
}

// TestSpaceOnEmptyOutdatedSelectsNothing (LB-47 FB8).
func TestSpaceOnEmptyOutdatedSelectsNothing(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelOutdated
	m = sendSpecial(m, tea.KeySpace)
	if len(m.batch.selected) != 0 {
		t.Fatalf("batch = %v, want empty", m.batch.selected)
	}
}
