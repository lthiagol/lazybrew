package gui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"slices"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// isMsgString reports whether e is msg.String(), the expression the key
// dispatch switches on.
func isMsgString(e ast.Expr) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "String" {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "msg"
}

// dispatchKeysInSource is every key string the key dispatch answers to: the
// string cases of each switch on msg.String() in the package, read from the
// source. Reading the source is what lets the test see a case the tables lack.
func dispatchKeysInSource(t *testing.T) map[string]bool {
	t.Helper()
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{}
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					sw, ok := n.(*ast.SwitchStmt)
					if !ok || !isMsgString(sw.Tag) {
						return true
					}
					for _, stmt := range sw.Body.List {
						for _, e := range stmt.(*ast.CaseClause).List {
							lit, ok := e.(*ast.BasicLit)
							if !ok || lit.Kind != token.STRING {
								continue
							}
							key, err := strconv.Unquote(lit.Value)
							if err != nil {
								t.Fatal(err)
							}
							keys[key] = true
						}
					}
					return true
				})
			}
		}
	}
	return keys
}

// TestKeymapMatchesDispatch holds the key dispatch to the key tables: every key
// the dispatch answers to is in the tables, and every key in the tables is
// answered. An entry with no footer word and no help line is in neither place,
// so it is an error too.
func TestKeymapMatchesDispatch(t *testing.T) {
	dispatched := dispatchKeysInSource(t)
	if len(dispatched) == 0 {
		t.Fatal("found no key cases in Update: the source scan is broken")
	}
	listed := map[string]bool{}
	for _, h := range keyTable {
		if h.desc == "" && h.help == "" {
			t.Errorf("%q is in the key table but on neither the footer nor the help", h.key)
		}
		for _, k := range h.dispatchKeys() {
			listed[k] = true
		}
	}
	for k := range dispatched {
		if !listed[k] {
			t.Errorf("the dispatch answers %q, which no key table lists", k)
		}
	}
	for k := range listed {
		if !dispatched[k] {
			t.Errorf("the key table lists %q, which the dispatch does not answer", k)
		}
	}
}

// TestHelpIsBuiltFromKeyTable holds the help screen to the table: every help
// line is an entry of the table, and every entry with a help line is on the
// help screen.
func TestHelpIsBuiltFromKeyTable(t *testing.T) {
	onHelp := map[keyHint]bool{}
	for _, page := range buildHelp() {
		for _, sec := range page.sections {
			for _, item := range sec.items {
				onHelp[item] = true
				if !inKeyTable(item) {
					t.Errorf("%s: help line %q %q is not in the key table", page.title, item.key, item.help)
				}
			}
		}
	}
	for _, h := range keyTable {
		if h.help != "" && !onHelp[h] {
			t.Errorf("%q %q is in the key table with a help line, but the help screen lacks it", h.key, h.help)
		}
	}
}

// TestFootersAreBuiltFromKeyTable holds every footer to the table, over every
// panel and every state the footers change with.
func TestFootersAreBuiltFromKeyTable(t *testing.T) {
	var footers [][]keyHint
	footers = append(footers, globalHints)
	for id := PanelStatus; id <= PanelSearch; id++ {
		footers = append(footers, panelHints(id))
	}
	for _, status := range []brew.ServiceStatus{brew.ServiceStarted, brew.ServiceStopped, brew.ServiceError, brew.ServiceScheduled} {
		footers = append(footers, serviceActionHints(brew.Service{Status: status}))
	}
	for _, tap := range []brew.Tap{
		{Name: "a/b", Installed: true},
		{Name: "a/b", Installed: true, Trusted: true},
		{Name: "homebrew/core", Installed: true, IsOfficial: true},
	} {
		footers = append(footers, tapActionHints(tap, true), tapActionHints(tap, false))
	}
	footers = append(footers, []keyHint{svLogStream}, []keyHint{keyClose}, []keyHint{keyHelpOff})

	for _, f := range footers {
		for _, h := range f {
			if h.desc == "" {
				t.Errorf("footer lists %q with no footer word", h.key)
			}
			if !inKeyTable(h) {
				t.Errorf("footer lists %q %q, which is not in the key table", h.key, h.desc)
			}
		}
	}
}

// inKeyTable reports whether h is an entry of the key table.
func inKeyTable(h keyHint) bool {
	for _, e := range keyTable {
		if e == h {
			return true
		}
	}
	return false
}

// dispatchKeys are the msg.String() spellings the key dispatch answers to h.
func (h keyHint) dispatchKeys() []string {
	if h.keys == "" {
		return []string{h.key}
	}
	return strings.Fields(h.keys)
}

// dispatchByPanel is the panels each key of the main key dispatch (handleKey in
// keys.go) answers on. A case that compares m.activePanel to a panel answers on
// that panel; one that compares it with != answers on every other panel; a case
// that names no panel answers on all of them. Cases that are not at the top of
// handleKey, such as the search box's, are not panel keys and are left out.
func dispatchByPanel(t *testing.T) map[PanelID]map[string]bool {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "keys.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]PanelID{}
	dispatch := map[PanelID]map[string]bool{}
	for id := PanelStatus; id <= PanelSearch; id++ {
		byName["Panel"+id.String()] = id
		dispatch[id] = map[string]bool{}
	}
	found := false
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "handleKey" {
			continue
		}
		for _, stmt := range fn.Body.List {
			sw, ok := stmt.(*ast.SwitchStmt)
			if !ok || !isMsgString(sw.Tag) {
				continue
			}
			found = true
			for _, c := range sw.Body.List {
				clause := c.(*ast.CaseClause)
				panels := answeredOn(clause.Body, byName)
				for _, e := range clause.List {
					lit, ok := e.(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						continue
					}
					key, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatal(err)
					}
					for _, id := range panels {
						dispatch[id][key] = true
					}
				}
			}
		}
	}
	if !found {
		t.Fatal("found no key switch in handleKey: the source scan is broken")
	}
	return dispatch
}

// answeredOn is the panels a case body answers on, from the comparisons with
// m.activePanel in it. See dispatchByPanel.
func answeredOn(body []ast.Stmt, byName map[string]PanelID) []PanelID {
	eq, neq := map[PanelID]bool{}, map[PanelID]bool{}
	for _, stmt := range body {
		ast.Inspect(stmt, func(n ast.Node) bool {
			b, ok := n.(*ast.BinaryExpr)
			if !ok || (b.Op != token.EQL && b.Op != token.NEQ) {
				return true
			}
			id, ok := panelComparison(b, byName)
			if !ok {
				return true
			}
			if b.Op == token.EQL {
				eq[id] = true
			} else {
				neq[id] = true
			}
			return true
		})
	}
	var panels []PanelID
	for id := PanelStatus; id <= PanelSearch; id++ {
		if len(eq) > 0 && !eq[id] || len(eq) == 0 && neq[id] {
			continue
		}
		panels = append(panels, id)
	}
	return panels
}

// panelComparison is the panel a comparison of m.activePanel with a panel
// constant names.
func panelComparison(b *ast.BinaryExpr, byName map[string]PanelID) (PanelID, bool) {
	for _, pair := range [][2]ast.Expr{{b.X, b.Y}, {b.Y, b.X}} {
		sel, ok := pair[0].(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "activePanel" {
			continue
		}
		id, ok := pair[1].(*ast.Ident)
		if !ok {
			continue
		}
		if panel, ok := byName[id.Name]; ok {
			return panel, true
		}
	}
	return 0, false
}

// footerOn is every footer line a panel shows, over the states its footer
// changes with: the selected Services or Taps item in each state, the Services
// logs tab, a finished operation (its Esc line) and the help screen.
func footerOn(id PanelID) []keyHint {
	var out []keyHint
	show := func(setup func(m *Model)) {
		m := newTestModel()
		m.activePanel = id
		setup(m)
		out = append(out, m.footerHints()...)
	}
	show(func(m *Model) {})
	show(func(m *Model) { m.activeTab = servicesLogsTab })
	show(func(m *Model) { m.opState = &Operation{Status: opSuccess} })
	show(func(m *Model) { m.showHelp = true })
	if id == PanelServices {
		for _, status := range []brew.ServiceStatus{brew.ServiceStarted, brew.ServiceStopped, brew.ServiceError, brew.ServiceScheduled} {
			show(func(m *Model) {
				m.panels[PanelServices].services = []brew.Service{{Status: status}}
			})
		}
	}
	if id == PanelTaps {
		for _, tap := range []brew.Tap{
			{Name: "a/b", Installed: true},
			{Name: "a/b", Installed: true, Trusted: true},
			{Name: "homebrew/core", Installed: true, IsOfficial: true},
		} {
			for _, unknown := range []bool{false, true} {
				show(func(m *Model) {
					m.panels[PanelTaps].taps = []brew.Tap{tap}
					m.panels[PanelTaps].trustUnknown = unknown
				})
			}
		}
	}
	return out
}

// helpOn is the help lines a panel's keys are on: the Global page and the page
// named for the panel.
func helpOn(id PanelID) []keyHint {
	var out []keyHint
	for _, page := range buildHelp() {
		if page.title != "Global" && page.title != id.String() {
			continue
		}
		for _, sec := range page.sections {
			out = append(out, sec.items...)
		}
	}
	return out
}

// TestPanelKeysAreOnFooterAndHelp holds each panel's dispatch to its footer and
// its help screen, both ways. A key the panel answers must be on its footer, when
// the table gives it a footer word, and on its help, when the table gives it a
// help line (see onFooterAndHelp). Every
// footer and help line must name a key the panel answers. Both lists come from
// the real footer and help, and the dispatch from the source of handleKey.
func TestPanelKeysAreOnFooterAndHelp(t *testing.T) {
	dispatch := dispatchByPanel(t)
	for id := PanelStatus; id <= PanelSearch; id++ {
		answers := dispatch[id]
		if len(answers) == 0 {
			t.Fatalf("%s: the dispatch answers no key", id)
		}
		footer, help := footerOn(id), helpOn(id)
		exempt := footerExempt[id]
		for k := range answers {
			if !onFooterAndHelp(k, footer, help, exempt[k]) {
				t.Errorf("%s: the dispatch answers %q, which its footer and help do not both show", id, k)
			}
		}
		for k := range exempt {
			if !answers[k] {
				t.Errorf("%s: footerExempt lists %q, which the panel does not answer; drop it", id, k)
			}
			if slices.ContainsFunc(footer, func(h keyHint) bool { return slices.Contains(h.dispatchKeys(), k) }) {
				t.Errorf("%s: footerExempt lists %q, but the footer shows it; drop it from footerExempt", id, k)
			}
		}
		for _, h := range footer {
			if !answersAny(h, answers) {
				t.Errorf("%s: the footer lists %q %q, which the panel does not answer", id, h.key, h.desc)
			}
		}
		for _, h := range help {
			if !answersAny(h, answers) {
				t.Errorf("%s: the help lists %q %q, which the panel does not answer", id, h.key, h.help)
			}
		}
	}
}

// footerExempt is the keys a panel answers that its footer leaves out on purpose.
// The footers are pinned by goldens, so each omission is recorded here instead of
// being changed. A listed key is checked on the help only. Each entry says why.
var footerExempt = map[PanelID]map[string]bool{
	// Search has no tabs: h and l type there, and the arrows do nothing. X zaps
	// an installed cask row; its footer word is left out, and its help line shows it.
	PanelSearch: {"left": true, "right": true, "X": true},
}

// onFooterAndHelp reports whether k is shown on the footer and on the help. The
// two are checked apart. When some table entry for k has a footer word, one such
// entry must be on the footer, and when some entry for k has a help line, one such
// entry must be on the help. An entry with a footer word and no help line shows k
// on the footer only, and the reverse on the help only, so one check cannot cover
// for the other. An exempt key (see footerExempt) skips the footer. k must have a
// footer word or a help line.
func onFooterAndHelp(k string, footer, help []keyHint, exempt bool) bool {
	var footerWord, helpLine, onFooter, onHelp bool
	for _, e := range keyTable {
		if !slices.Contains(e.dispatchKeys(), k) {
			continue
		}
		if e.desc != "" {
			footerWord = true
			onFooter = onFooter || slices.Contains(footer, e)
		}
		if e.help != "" {
			helpLine = true
			onHelp = onHelp || slices.Contains(help, e)
		}
	}
	return (footerWord || helpLine) && (exempt || !footerWord || onFooter) && (!helpLine || onHelp)
}

// answersAny reports whether the panel answers one of h's keys.
func answersAny(h keyHint, answers map[string]bool) bool {
	for _, k := range h.dispatchKeys() {
		if answers[k] {
			return true
		}
	}
	return false
}

// modalKeys is the msg.String() cases of each modal's Update, read from the
// source of the modal package.
func modalKeys(t *testing.T) map[string]map[string]bool {
	t.Helper()
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "modal", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]map[string]bool{}
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name.Name != "Update" || fn.Recv == nil {
					continue
				}
				star, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
				if !ok {
					continue
				}
				recv := star.X.(*ast.Ident).Name
				keys[recv] = map[string]bool{}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					sw, ok := n.(*ast.SwitchStmt)
					if !ok || !isMsgString(sw.Tag) {
						return true
					}
					for _, c := range sw.Body.List {
						for _, e := range c.(*ast.CaseClause).List {
							lit, ok := e.(*ast.BasicLit)
							if !ok || lit.Kind != token.STRING {
								continue
							}
							key, err := strconv.Unquote(lit.Value)
							if err != nil {
								t.Fatal(err)
							}
							keys[recv][key] = true
						}
					}
					return true
				})
			}
		}
	}
	return keys
}

// TestModalHintsMatchUpdate holds each modal's Hints to the keys its Update
// answers. A key the Update answers must be listed by Hints, or be on the short
// list of keys shown another way, and every key Hints lists must be answered.
func TestModalHintsMatchUpdate(t *testing.T) {
	modals := map[string]modal.Hinter{
		"ConfirmModal": &modal.ConfirmModal{},
		"InputModal":   &modal.InputModal{},
		"MenuModal":    &modal.MenuModal{},
	}
	// The keys a modal answers that its Hints leave out, each shown another
	// way: the confirm's choice keys show as the Yes and No buttons, and the
	// menu's digits are drawn on each item's row.
	digits := map[string]bool{}
	for d := '0'; d <= '9'; d++ {
		digits[string(d)] = true
	}
	unlisted := map[string]map[string]bool{
		"ConfirmModal": {"enter": true, "h": true, "left": true, "l": true, "right": true},
		"MenuModal":    digits,
	}

	updates := modalKeys(t)
	for name, m := range modals {
		answers, ok := updates[name]
		if !ok {
			t.Errorf("%s: no Update in the modal package", name)
			continue
		}
		listed := map[string]bool{}
		for _, h := range m.Hints() {
			if h.Keys == "" {
				t.Errorf("%s: hint %q has no Keys", name, h.Key)
			}
			for _, k := range strings.Fields(h.Keys) {
				listed[k] = true
			}
		}
		for k := range listed {
			if !answers[k] {
				t.Errorf("%s: Hints lists %q, which Update does not answer", name, k)
			}
		}
		for k := range answers {
			if !listed[k] && !unlisted[name][k] {
				t.Errorf("%s: Update answers %q, which Hints does not list", name, k)
			}
		}
		for k := range unlisted[name] {
			if !answers[k] {
				t.Errorf("%s: the unlisted key %q is no longer answered; drop it from the list", name, k)
			}
		}
	}
}

// keyTable lists every entry of the key tables in keybindings.go. The tests
// walk it: a new entry goes here too, or the tests cannot see it.
var keyTable = []keyHint{
	keyNext, keyPrev, keyJump, keyScroll, keyEnds, keyTabs, keySearch,
	keyHelp, keyRefresh, keyLog, keyQuit, keyClose, keyHelpOff,
	keyNavigate, keyPanelTabs, keyUpgradeSelected, keyUpgradeAll, keyCopy,
	fInstall, fRemove, fReinstall, fPin, fLeaves, fFetch,
	cInstall, cRemove, cZap, cReinstall, cUpgrade, cFetch,
	oSelect, oSelectAll,
	tAdd, tUntap, tTrust, tUntrust, tTrustItem, tRepair,
	svStart, svStop, svRestart, svRun, svCleanup, svLogStream,
	stCleanup, stDoctor, stAutoremove, stBrewfile, stVulns, stMissing, stTabs,
	sExecute, sType, sInstall, sInstallCap, sUninstall, sZap, sFetch,
}

// TestRetiredTabAndPageKeysDoNothing (lb19 AC-03): [ and ] switch no tab, and
// PgUp and PgDn scroll nothing, on a panel that has tabs and content to scroll.
func TestRetiredTabAndPageKeysDoNothing(t *testing.T) {
	m := scrollFixture(t)
	tab, panel := m.activeTab, m.activePanel
	scroll := m.contentScroll[m.contentKey()]

	for _, key := range []string{"[", "]"} {
		m = sendKey(m, key)
	}
	m = sendSpecial(m, tea.KeyPgDown)
	m = sendSpecial(m, tea.KeyPgUp)

	if m.activeTab != tab || m.activePanel != panel {
		t.Errorf("[ ] switched the tab or panel: tab %d->%d panel %v->%v", tab, m.activeTab, panel, m.activePanel)
	}
	if got := m.contentScroll[m.contentKey()]; got != scroll {
		t.Errorf("PgUp/PgDn scrolled the content: offset %d->%d", scroll, got)
	}

	// The open help screen has no page keys either.
	m = sendKey(m, "?")
	m = sendKey(m, "j")
	m = sendKey(m, "j")
	help := m.helpScroll
	if !m.showHelp || help != 2 {
		t.Fatalf("help not open at offset 2: showHelp=%v helpScroll=%d", m.showHelp, help)
	}
	m = sendSpecial(m, tea.KeyPgDown)
	m = sendSpecial(m, tea.KeyPgUp)
	if got := m.helpScroll; got != help {
		t.Errorf("PgUp/PgDn scrolled the help screen: offset %d->%d", help, got)
	}
}
