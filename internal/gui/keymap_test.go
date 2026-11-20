package gui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strconv"
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
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

// keyTable lists every entry of the key tables in keybindings.go. The tests
// walk it: a new entry goes here too, or the tests cannot see it.
var keyTable = []keyHint{
	keyNext, keyPrev, keyJump, keyScroll, keyPage, keyEnds, keyTabs, keySearch,
	keyHelp, keyRefresh, keyLog, keyQuit, keyClose, keyHelpOff,
	keyNavigate, keyPanelTabs, keyUpgradeSelected, keyUpgradeAll,
	fInstall, fRemove, fReinstall, fPin, fLeaves, fFetch,
	cInstall, cRemove, cZap, cReinstall, cUpgrade, cFetch,
	oSelect, oSelectAll,
	tAdd, tUntap, tTrust, tUntrust, tTrustItem, tRepair,
	svStart, svStop, svRestart, svRun, svCleanup, svLogStream,
	stCleanup, stDoctor, stAutoremove, stBrewfile, stVulns, stMissing, stTabs,
	sExecute, sType, sInstall, sInstallCap, sFetch,
}
