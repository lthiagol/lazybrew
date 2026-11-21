package gui

import (
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
)

// filterFormulae is a Formulae panel loaded with one formula per name, its rows
// laid out the way landData lays them out.
func filterFormulae(names ...string) *panelData {
	p := &panelData{id: PanelFormulae}
	for _, n := range names {
		p.items = append(p.items, n+"  1.0")
		p.formulae = append(p.formulae, brew.Formula{Name: n})
	}
	return p
}

// filterCasks is a Casks panel loaded with one cask per name.
func filterCasks(names ...string) *panelData {
	p := &panelData{id: PanelCasks}
	for _, n := range names {
		p.items = append(p.items, n+"  1.0")
		p.casks = append(p.casks, brew.Cask{Name: n})
	}
	return p
}

// names lists the names of the panel's rows, read through nameAt.
func names(p *panelData) []string {
	var out []string
	for i := range p.items {
		out = append(out, p.nameAt(i))
	}
	return out
}

func sameNames(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// TestQueryNarrowsItemsFormulaeAndCasksTogether: the query keeps the rows of
// items, formulae and casks that match, in step, case-insensitively.
func TestQueryNarrowsItemsFormulaeAndCasksTogether(t *testing.T) {
	p := filterFormulae("ripgrep", "rip-all", "wget")
	p.setQuery("RIP")
	if want := []string{"ripgrep", "rip-all"}; !sameNames(names(p), want) {
		t.Fatalf("items name %v, want %v", names(p), want)
	}
	if len(p.formulae) != 2 || p.formulae[1].Name != "rip-all" {
		t.Fatalf("formulae = %v, want the two matches", p.formulae)
	}

	c := filterCasks("firefox", "fire-tools", "chrome")
	c.setQuery("fire")
	if want := []string{"firefox", "fire-tools"}; !sameNames(names(c), want) {
		t.Fatalf("cask items name %v, want %v", names(c), want)
	}
	if len(c.casks) != 2 || c.casks[0].Name != "firefox" {
		t.Fatalf("casks = %v, want the two matches", c.casks)
	}
}

// TestQueryIndexesTheFilteredRows: the cursor and every selected-row lookup read
// the filtered list, not the full one.
func TestQueryIndexesTheFilteredRows(t *testing.T) {
	p := filterFormulae("ripgrep", "wget", "zstd")
	p.setQuery("zs")
	if p.selected != 0 || p.selectedName() != "zstd" {
		t.Fatalf("cursor on %q at %d, want zstd at 0", p.selectedName(), p.selected)
	}
	if f := p.selectedFormula(); f == nil || f.Name != "zstd" {
		t.Fatalf("selectedFormula = %v, want zstd", f)
	}
	if got := p.nameAt(0); got != "zstd" {
		t.Fatalf("nameAt(0) = %q, want zstd", got)
	}

	c := filterCasks("firefox", "chrome")
	c.setQuery("chr")
	if got := c.selectedCask(); got == nil || got.Name != "chrome" {
		t.Fatalf("selectedCask = %v, want chrome", got)
	}
}

// TestQueryWithNoMatchEmptiesTheList: a query nothing matches leaves no rows and
// no selection, and clearing it brings the list back.
func TestQueryWithNoMatchEmptiesTheList(t *testing.T) {
	p := filterFormulae("ripgrep", "wget")
	p.setQuery("nothing-like-it")
	if len(p.items) != 0 || len(p.formulae) != 0 {
		t.Fatalf("rows left after a no-match query: items %v, formulae %v", p.items, p.formulae)
	}
	if p.selectedName() != "" || p.selectedFormula() != nil {
		t.Fatal("a no-match query should select nothing")
	}
	if p.items == nil {
		t.Fatal("items must stay non-nil: a nil list reads as never loaded")
	}
	p.setQuery("")
	if want := []string{"ripgrep", "wget"}; !sameNames(names(p), want) {
		t.Fatalf("after clearing, items name %v, want %v", names(p), want)
	}
}

// TestQueryComposesWithLeaves: the leaves filter and the query both narrow the
// list, and each can be cleared without losing the other.
func TestQueryComposesWithLeaves(t *testing.T) {
	p := filterFormulae("a1", "a2", "b1")
	p.toggleLeaves([]string{"a1", "b1"})
	if want := []string{"a1", "b1"}; !sameNames(names(p), want) {
		t.Fatalf("leaves: items name %v, want %v", names(p), want)
	}

	p.setQuery("a")
	if want := []string{"a1"}; !sameNames(names(p), want) {
		t.Fatalf("leaves and query: items name %v, want %v", names(p), want)
	}

	p.toggleLeaves(nil)
	if want := []string{"a1", "a2"}; !sameNames(names(p), want) {
		t.Fatalf("leaves off, query on: items name %v, want %v", names(p), want)
	}

	p.setQuery("")
	if want := []string{"a1", "a2", "b1"}; !sameNames(names(p), want) {
		t.Fatalf("both cleared: items name %v, want %v", names(p), want)
	}
}

// TestLeavesToggledWithQueryActive: toggling the leaves filter re-derives from
// the full list and keeps the query.
func TestLeavesToggledWithQueryActive(t *testing.T) {
	p := filterFormulae("a1", "a2", "b1")
	p.setQuery("a")
	p.toggleLeaves([]string{"a2"})
	if want := []string{"a2"}; !sameNames(names(p), want) {
		t.Fatalf("leaves on with query: items name %v, want %v", names(p), want)
	}
	p.toggleLeaves([]string{"a2"})
	if want := []string{"a1", "a2"}; !sameNames(names(p), want) {
		t.Fatalf("leaves off with query: items name %v, want %v", names(p), want)
	}
	if p.leavesActive {
		t.Fatal("leaves should be off")
	}
}

// TestClearingQueryKeepsCursorOnSameItem: the cursor stays on the item it was on
// when the query is cleared.
func TestClearingQueryKeepsCursorOnSameItem(t *testing.T) {
	p := filterFormulae("ripgrep", "wget", "zstd")
	p.setQuery("ge")
	if p.selectedName() != "wget" {
		t.Fatalf("typing should land on the first match, got %q", p.selectedName())
	}
	p.setQuery("")
	if p.selectedName() != "wget" || p.selected != 1 {
		t.Fatalf("after clearing, cursor on %q at %d, want wget at 1", p.selectedName(), p.selected)
	}
}

// TestLandDataReappliesQueryAndKeepsCursor: a refresh lands the new data under
// the query, and the cursor stays on its name while the name is still listed.
func TestLandDataReappliesQueryAndKeepsCursor(t *testing.T) {
	m := newTestModel()
	p := m.panels[PanelFormulae]
	*p = *filterFormulae("ripgrep", "rip-tool", "wget")
	p.setQuery("rip")
	p.selected = 1
	if p.selectedName() != "rip-tool" {
		t.Fatalf("setup: cursor on %q, want rip-tool", p.selectedName())
	}

	m.landData(DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"rip-tool  1.0", "ripgrep  1.0", "zz-rip  1.0", "wget  1.0"},
		Formulae: []brew.Formula{{Name: "rip-tool"}, {Name: "ripgrep"}, {Name: "zz-rip"}, {Name: "wget"}},
	}, false)
	if want := []string{"rip-tool", "ripgrep", "zz-rip"}; !sameNames(names(p), want) {
		t.Fatalf("after landing, items name %v, want %v", names(p), want)
	}
	if p.selectedName() != "rip-tool" {
		t.Fatalf("cursor on %q, want rip-tool (kept by name)", p.selectedName())
	}
}

// TestLandDataClampsCursorWhenItemIsGone: a refresh that drops the item under the
// cursor clamps the cursor, as an unfiltered list does.
func TestLandDataClampsCursorWhenItemIsGone(t *testing.T) {
	m := newTestModel()
	p := m.panels[PanelFormulae]
	*p = *filterFormulae("ripgrep", "rip-tool")
	p.setQuery("rip")
	p.selected = 1

	m.landData(DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"ripgrep  1.0"},
		Formulae: []brew.Formula{{Name: "ripgrep"}},
	}, false)
	if p.selected != 0 || p.selectedName() != "ripgrep" {
		t.Fatalf("cursor on %q at %d, want ripgrep at 0", p.selectedName(), p.selected)
	}
}

// TestLandDataKeepsFullListWhenNoQuery: with no filter on, a landing sets the
// list as it always has.
func TestLandDataKeepsFullListWhenNoQuery(t *testing.T) {
	m := newTestModel()
	m.landData(DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"a  1.0", "b  1.0"},
		Formulae: []brew.Formula{{Name: "a"}, {Name: "b"}},
	}, false)
	p := m.panels[PanelFormulae]
	if want := []string{"a", "b"}; !sameNames(names(p), want) || p.narrowed() {
		t.Fatalf("items name %v (narrowed %v), want %v unnarrowed", names(p), p.narrowed(), want)
	}
}
