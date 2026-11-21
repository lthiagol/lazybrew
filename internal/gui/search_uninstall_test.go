package gui

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// searchUninstallRunner records the write args it streams and answers the
// installed lists from the installed set. A write that streams clears the set
// for its name, as brew does, so a later read sees the change.
type searchUninstallRunner struct {
	mu        sync.Mutex
	installed map[string]bool // installed cask token or formula name
	writes    [][]string
}

func (r *searchUninstallRunner) mock() *brewtest.MockRunner {
	mr := brewtest.NewMockRunner()
	mr.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		r.mu.Lock()
		r.writes = append(r.writes, args)
		if len(args) > 0 && args[0] == "uninstall" {
			delete(r.installed, args[len(args)-1])
		}
		r.mu.Unlock()
		ch := make(chan string)
		close(ch)
		errCh := make(chan error, 1)
		errCh <- nil
		close(errCh)
		return ch, errCh
	}
	mr.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		r.mu.Lock()
		defer r.mu.Unlock()
		var casks, formulae []string
		for name := range r.installed {
			casks = append(casks, `{"token":"`+name+`","name":["`+name+`"]}`)
		}
		if len(args) > 0 && args[len(args)-1] == "--cask" {
			return jsonInto(result, `{"casks": [`+strings.Join(casks, ",")+`]}`)
		}
		for name := range r.installed {
			formulae = append(formulae, `{"name":"`+name+`"}`)
		}
		return jsonInto(result, `{"formulae": [`+strings.Join(formulae, ",")+`]}`)
	}
	return mr
}

func (r *searchUninstallRunner) writesSeen() [][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([][]string(nil), r.writes...)
}

// searchRowsModel is a model on the Search panel with rows already loaded, the
// rows' marks shown as their items show them.
func searchRowsModel(r *searchUninstallRunner, rows []brew.SearchResult) *Model {
	m := newTestModel()
	m.client = brew.NewClient(r.mock())
	m.activePanel = PanelSearch
	m.searchResults = rows
	p := m.panels[PanelSearch]
	p.results = rows
	p.items = make([]string, len(rows))
	for i, row := range rows {
		typ := "formula"
		if row.IsCask {
			typ = "cask"
		}
		mark := ""
		if row.Installed {
			mark = "installed"
		}
		p.items[i] = row.Name + "  " + typ + "  " + mark + "  "
	}
	return m
}

// confirmSearchUninstall presses key on the Search panel, answers the dependency
// check and the confirm, and runs the queued task. It fails the test if the key
// opened no confirm.
func confirmSearchUninstall(t *testing.T, m *Model, key string) *Model {
	t.Helper()
	nm, cmd := pressKeyOn(m, key)
	if cmd == nil {
		t.Fatalf("%s queued nothing", key)
	}
	nm = updateModel(nm, cmd())
	if nm.activeModal == nil {
		t.Fatalf("%s opened no confirm", key)
	}
	done := confirmAndRun(t, nm)
	return &done
}

// TestSearchUninstallsInstalledCask (lb19 AC-04): x on an installed cask row
// asks to uninstall it and runs brew uninstall --cask through the task manager.
func TestSearchUninstallsInstalledCask(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{"claude-code": true}}
	m := searchRowsModel(r, []brew.SearchResult{{Name: "claude-code", IsCask: true, Installed: true}})

	confirmSearchUninstall(t, m, "x")
	if got := r.writesSeen(); len(got) != 1 || strings.Join(got[0], " ") != "uninstall --cask claude-code" {
		t.Fatalf("writes = %v, want [uninstall --cask claude-code]", got)
	}
}

// TestSearchUninstallsInstalledFormula (lb19 AC-04): x on an installed formula
// row runs brew uninstall <name>.
func TestSearchUninstallsInstalledFormula(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{"jq": true}}
	m := searchRowsModel(r, []brew.SearchResult{{Name: "jq", IsFormula: true, Installed: true}})

	confirmSearchUninstall(t, m, "x")
	if got := r.writesSeen(); len(got) != 1 || strings.Join(got[0], " ") != "uninstall jq" {
		t.Fatalf("writes = %v, want [uninstall jq]", got)
	}
}

// TestSearchXOnUninstalledRowToastsAndRunsNothing (lb19 AC-04): x on a row that
// is not installed says so, opens no confirm and runs nothing.
func TestSearchXOnUninstalledRowToastsAndRunsNothing(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{}}
	m := searchRowsModel(r, []brew.SearchResult{{Name: "ripgrep", IsFormula: true}})

	nm, cmd := pressKeyOn(m, "x")
	if nm.activeModal != nil {
		t.Fatal("x on a row that is not installed opened a confirm")
	}
	if !lastToastIs(nm, "ripgrep is not installed") {
		t.Fatalf("toasts = %v, want ripgrep is not installed", toastMessages(nm))
	}
	if cmd != nil {
		if _, ok := cmd().(TaskStartedMsg); ok {
			t.Fatal("x on a row that is not installed started a task")
		}
	}
	if got := r.writesSeen(); len(got) != 0 {
		t.Fatalf("writes = %v, want none", got)
	}
}

// TestSearchIOnInstalledRowToastsAndRunsNothing (lb19 AC-04): i on an installed
// row says it is installed and points at x, and runs nothing.
func TestSearchIOnInstalledRowToastsAndRunsNothing(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{"jq": true}}
	m := searchRowsModel(r, []brew.SearchResult{{Name: "jq", IsFormula: true, Installed: true}})

	nm, _ := pressKeyOn(m, "i")
	if nm.activeModal != nil {
		t.Fatal("i on an installed row opened a confirm")
	}
	if !lastToastIs(nm, "jq is already installed, x to uninstall") {
		t.Fatalf("toasts = %v, want jq is already installed, x to uninstall", toastMessages(nm))
	}
	if got := r.writesSeen(); len(got) != 0 {
		t.Fatalf("writes = %v, want none", got)
	}
}

// TestSearchZapsInstalledCask (lb19 AC-04): X on an installed cask row zaps it,
// as X on the Casks panel does. X on a formula row does nothing.
func TestSearchZapsInstalledCask(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{"claude-code": true}}
	m := searchRowsModel(r, []brew.SearchResult{{Name: "claude-code", IsCask: true, Installed: true}})

	confirmSearchUninstall(t, m, "X")
	if got := r.writesSeen(); len(got) != 1 || strings.Join(got[0], " ") != "uninstall --zap --cask claude-code" {
		t.Fatalf("writes = %v, want [uninstall --zap --cask claude-code]", got)
	}

	f := &searchUninstallRunner{installed: map[string]bool{"jq": true}}
	fm := searchRowsModel(f, []brew.SearchResult{{Name: "jq", IsFormula: true, Installed: true}})
	nm, _ := pressKeyOn(fm, "X")
	if nm.activeModal != nil || len(f.writesSeen()) != 0 {
		t.Fatal("X on an installed formula row did something")
	}
}

// TestSearchMarkClearsAfterUninstall (lb19 AC-04): once the uninstall completes,
// the row loses its installed mark without a new search.
func TestSearchMarkClearsAfterUninstall(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{"claude-code": true}}
	m := searchRowsModel(r, []brew.SearchResult{{Name: "claude-code", IsCask: true, Installed: true}})

	m = confirmSearchUninstall(t, m, "x")
	// The completion is what the task manager reports for the run above. Run
	// the batch its update returns and feed each message back, so the re-mark
	// is whatever the completion itself queued.
	nm, done := m.Update(TaskCompletedMsg{ID: mutationID(mutUninstall, "claude-code"), Title: "Uninstall claude-code"})
	m = asModel(nm)
	for _, msg := range drainCmds(done) {
		m = updateModel(m, msg)
	}

	if got := m.searchResults[0].Installed; got {
		t.Error("the row still says installed after the uninstall")
	}
	if got := m.panels[PanelSearch].items[0]; strings.Contains(got, "installed") {
		t.Errorf("the row's item still shows installed: %q", got)
	}
}

// TestStaleSearchMarksDoNotLandOnNewRows (lb19 AC-04, F-r1-03): marks read for
// an older result set arrive after a newer search. They are matched by name, so
// no row of the new set changes its mark.
func TestStaleSearchMarksDoNotLandOnNewRows(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelSearch
	// The new search's first row is not installed; the old marks say the old
	// search's first row (jq) is. A positional match would mark the new row.
	fresh := []brew.SearchResult{{Name: "wget", IsFormula: true}, {Name: "claude-code", IsCask: true}}
	m = updateModel(m, SearchDoneMsg{Results: searchItems(fresh), Raw: fresh})

	m = updateModel(m, SearchMarksMsg{Results: []brew.SearchResult{{Name: "jq", IsFormula: true, Installed: true}}})

	for _, r := range m.searchResults {
		if r.Installed {
			t.Errorf("row %q became installed from marks for another search", r.Name)
		}
	}
	for i, item := range m.panels[PanelSearch].items {
		if strings.Contains(item, "installed") {
			t.Errorf("item %d shows installed from marks for another search: %q", i, item)
		}
	}
}

// lastToastIs reports whether the newest toast says message.
func lastToastIs(m *Model, message string) bool {
	return len(m.toasts) > 0 && m.toasts[len(m.toasts)-1].Message == message
}

func toastMessages(m *Model) []string {
	var out []string
	for _, t := range m.toasts {
		out = append(out, t.Message)
	}
	return out
}

// TestSearchFooterAndHelpShowUninstall (lb19 AC-04 point 4): the Search footer
// and its help page name x as uninstall.
func TestSearchFooterAndHelpShowUninstall(t *testing.T) {
	m := searchRowsModel(&searchUninstallRunner{}, nil)
	if !slices.ContainsFunc(m.footerHints(), func(h keyHint) bool { return h.key == "x" && h.desc == "uninstall" }) {
		t.Errorf("Search footer = %v, want x uninstall", m.footerHints())
	}
	var help []keyHint
	for _, page := range buildHelp() {
		if page.title == "Search" {
			for _, sec := range page.sections {
				help = append(help, sec.items...)
			}
		}
	}
	if !slices.Contains(help, sUninstall) {
		t.Errorf("Search help lacks the x uninstall line: %v", help)
	}
}
