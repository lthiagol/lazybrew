package gui

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// TestNeedsTabFetchTable (cr3 AC-07): exactly the tabs that brew fetches.
func TestNeedsTabFetchTable(t *testing.T) {
	want := map[PanelID][]int{
		PanelStatus:   {statusConfigTab, statusDoctorTab},
		PanelFormulae: {formulaDepsTab, formulaUsesTab, formulaFilesTab},
	}
	for panel := PanelStatus; panel <= PanelSearch; panel++ {
		for tab := 0; tab < 8; tab++ {
			got := needsTabFetch(panel, tab)
			wantFetch := false
			for _, w := range want[panel] {
				wantFetch = wantFetch || w == tab
			}
			if got != wantFetch {
				t.Errorf("needsTabFetch(%v, %d) = %v, want %v", panel, tab, got, wantFetch)
			}
		}
	}
}

// TestNeedsTabFetchAllocatesNothing: it is called on every j, k and data
// landing, so it builds nothing per call.
func TestNeedsTabFetchAllocatesNothing(t *testing.T) {
	if n := testing.AllocsPerRun(100, func() { needsTabFetch(PanelFormulae, formulaFilesTab) }); n != 0 {
		t.Fatalf("needsTabFetch allocates %v times per call, want 0", n)
	}
}

// oldSortFormulae is the comparator sortFormulae used before cr3 AC-07. The new
// sort must give the same order on every input, ties included.
func oldSortFormulae(fs []brew.Formula) {
	sort.SliceStable(fs, func(i, j int) bool {
		oi, oj := 1, 1
		if fs[i].Outdated {
			oi = 0
		}
		if fs[j].Outdated {
			oj = 0
		}
		if oi != oj {
			return oi < oj
		}
		pi, pj := 1, 1
		if fs[i].Pinned {
			pi = 0
		}
		if fs[j].Pinned {
			pj = 0
		}
		if pi != pj {
			return pi < pj
		}
		return strings.ToLower(fs[i].Name) < strings.ToLower(fs[j].Name)
	})
}

// TestSortFormulaeKeepsTheOrder (cr3 AC-07): the order is outdated, then pinned,
// then name case-insensitively, and equal keys keep their input order. Names
// that differ only in case are the ties the fixture is built around.
func TestSortFormulaeKeepsTheOrder(t *testing.T) {
	fixture := []brew.Formula{
		{Name: "Bash"}, {Name: "bash", Outdated: true}, {Name: "BASH", Pinned: true},
		{Name: "curl"}, {Name: "Curl", Outdated: true}, {Name: "cURL"},
		{Name: "zlib", Pinned: true}, {Name: "ZLIB", Outdated: true, Pinned: true},
		{Name: "jq"}, {Name: "Jq"}, {Name: "jQ", Pinned: true}, {Name: "ncurses"},
		{Name: "ripgrep"}, {Name: "RipGrep"}, {Name: "ripgrep", Outdated: true},
	}
	want := append([]brew.Formula(nil), fixture...)
	oldSortFormulae(want)
	got := append([]brew.Formula(nil), fixture...)
	sortFormulae(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sortFormulae order\n%v\nwant\n%v", formulaNames(got), formulaNames(want))
	}
}

func formulaNames(fs []brew.Formula) []string {
	names := make([]string, 0, len(fs))
	for _, f := range fs {
		names = append(names, f.Name)
	}
	return names
}

// TestHelpRowsAreCachedPerWidthAndTheme (cr3 AC-07): the help is laid out once
// for a width and theme, and laid out again when either changes.
func TestHelpRowsAreCachedPerWidthAndTheme(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	style.ApplyTheme(style.DarkTheme())
	first := helpRows(100)
	if second := helpRows(100); &first[0] != &second[0] {
		t.Fatal("the same width and theme should give the cached rows")
	}
	if other := helpRows(60); len(other) == 0 || &other[0] == &first[0] {
		t.Fatal("a different width should lay the help out again")
	}
	style.ApplyTheme(style.LightTheme())
	light := helpRows(100)
	if &light[0] == &first[0] {
		t.Fatal("a different theme should lay the help out again")
	}
	if !reflect.DeepEqual(helpRows(100), light) {
		t.Fatal("the light rows should be stable once laid out")
	}
}

// TestToastNamesFollowTheToast (cr3 AC-07): a toast's motion names are its own,
// the same for every call, and no other toast's.
func TestToastNamesFollowTheToast(t *testing.T) {
	a := modal.NewToast("one", modal.ToastInfo)
	b := modal.NewToast("one", modal.ToastInfo)
	if toastName(a, "slot") != toastName(a, "slot") {
		t.Fatal("a toast's name should not change between calls")
	}
	if toastName(a, "slot") == toastName(b, "slot") {
		t.Fatal("two toasts with the same text should not share motion names")
	}
	if !strings.HasPrefix(toastName(a, "dy"), toastKey(a)+":") {
		t.Fatalf("the name %q should begin with its toast key %q", toastName(a, "dy"), toastKey(a))
	}
}
