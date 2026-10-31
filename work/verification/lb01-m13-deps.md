# Verification — LB-01: the M13 Deps-tab fix

**Slice:** `lb01-m13-deps-verify`
**Backlog item:** `LB-01` (`work/backlog.md`)
**Fix under verification:** `ac5bac6` — *fix(gui): stop Formulae Deps infinite loading (M13)*
**Verified on:** 2026-10-31, branch `wip`
**Result:** all five invariants present, all regression tests pass, all four gate
commands pass. No failure to report.

This is a verification record only. No code, test, Makefile or workflow file was
changed in this slice.

---

## The five invariants of the M13 fix

Each row gives the invariant, the command that proves it is present, the command's
actual output, and the line it was found at.

### I-1 — a tab-content fetch is bounded by a context timeout

Command:

```bash
grep -n "fetchTabContentTimeout" internal/gui/commands.go
```

Actual output:

```
18:// fetchTabContentTimeout bounds the time a tab-content fetch (brew deps --tree,
22:var fetchTabContentTimeout = 15 * time.Second
833:		ctx, cancel := context.WithTimeout(context.Background(), fetchTabContentTimeout)
```

Found at `internal/gui/commands.go:22` (the timeout value) and
`internal/gui/commands.go:833` (`context.WithTimeout` wrapping the fetch). Both
anchors named in the brief are present.

### I-2 — after `DataLoadedMsg`, the tab content load is re-kicked

Command:

```bash
sed -n '455,456p' internal/gui/gui.go
```

Actual output:

```
		if msg.PanelID == m.activePanel && needsTabFetch(m.activePanel, m.activeTab) {
			return m, m.loadTabContent()
```

Found at `internal/gui/gui.go:455-456` — the `DataLoadedMsg` handler re-kicks
`loadTabContent()`.

### I-3 — the Deps tab fast-paths from the cached formula's dependencies instead of calling brew

Command:

```bash
grep -n "renderCachedTabContent" internal/gui/commands.go
```

Actual output:

```
779:		if content, ok := renderCachedTabContent(panelID, tabIdx, panel); ok {
788:// renderCachedTabContent returns a body string for tabs whose content is
794:func renderCachedTabContent(panelID PanelID, tabIdx int, panel *panelData) (string, bool) {
```

Found at the three anchors named in the brief: the call site at
`internal/gui/commands.go:779`, the function at `internal/gui/commands.go:794`,
and the runtime-dependency read inside `renderFormulaDeps` at
`internal/gui/commands.go:810`. The third anchor was confirmed with:

```bash
grep -n "runtime := f.Dependencies" internal/gui/commands.go
```

Actual output:

```
810:	runtime := f.Dependencies
```

### I-4 — `parseFormula` maps build deps from `build_dependencies`, not `dependencies`

Command:

```bash
grep -n "buildDeps := append" internal/brew/formulae.go
```

Actual output:

```
318:	buildDeps := append([]string(nil), f.BuildDependencies...)
```

Found at `internal/brew/formulae.go:318`. The source field is the
`build_dependencies` one, not `dependencies`. Confirmed with:

```bash
grep -n "BuildDependencies   \[\]string" internal/brew/formulae.go
```

Actual output:

```
64:	BuildDependencies   []string        `json:"build_dependencies"`
```

### I-5 — the tab/panel key handlers return early so the fetch command is not dropped

Command:

```bash
sed -n '627,634p' internal/gui/gui.go
```

Actual output:

```
			if needsTabFetch(m.activePanel, m.activeTab) {
				return m, m.loadTabContent()
			}
			return m, nil
		case "k", "up":
			m.panels[m.activePanel].up()
			if needsTabFetch(m.activePanel, m.activeTab) {
				return m, m.loadTabContent()
```

Found at `internal/gui/gui.go:627-634` — both handlers return the fetch command
immediately instead of falling through.

---

## The regression tests

### GUI Deps tests

Command:

```bash
go test ./internal/gui/ -run 'TestDepsTabLoads|TestDepsSelectionChangeUpdatesContent' -count=1 -v
```

Actual output:

```
=== RUN   TestDepsSelectionChangeUpdatesContent
--- PASS: TestDepsSelectionChangeUpdatesContent (0.00s)
=== RUN   TestDepsTabLoads
--- PASS: TestDepsTabLoads (0.00s)
PASS
ok  	github.com/lthiagol/lazybrew/internal/gui	3.108s
```

### Brew formula tests

Command:

```bash
go test ./internal/brew/ -run 'TestParseFormulaBuildDepsFromBuildDependencies|TestFormulaeReaderList' -count=1 -v
```

Actual output:

```
=== RUN   TestFormulaeReaderList
--- PASS: TestFormulaeReaderList (0.00s)
=== RUN   TestParseFormulaBuildDepsFromBuildDependencies
--- PASS: TestParseFormulaBuildDepsFromBuildDependencies (0.00s)
=== RUN   TestFormulaeReaderListCached
--- PASS: TestFormulaeReaderListCached (0.00s)
=== RUN   TestFormulaeReaderListEmpty
--- PASS: TestFormulaeReaderListEmpty (0.00s)
PASS
ok  	github.com/lthiagol/lazybrew/internal/brew	0.064s
```

Note: the `-run` pattern `TestFormulaeReaderList` is an unanchored regex, so it
also selected `TestFormulaeReaderListCached` and `TestFormulaeReaderListEmpty`.
Both are in the tree and both pass. `TestFormulaeReaderList` (the one with the
dependency assertions) passes.

Test declarations in the tree, for reference:

```
internal/gui/gui_test.go:1370:func TestDepsSelectionChangeUpdatesContent(t *testing.T) {
internal/gui/gui_test.go:1446:func TestDepsTabLoads(t *testing.T) {
internal/brew/formulae_test.go:98:func TestFormulaeReaderList(t *testing.T) {
internal/brew/formulae_test.go:173:func TestParseFormulaBuildDepsFromBuildDependencies(t *testing.T) {
```

---

## The gate

All four commands were run on 2026-10-31. Every one passed.

### `make build`

```
go build -o bin/lazybrew ./cmd/lazybrew
```

Exit status 0.

### `make lint`

```
go vet ./...
```

Exit status 0.

### `go test ./... -race -count=1`

```
?   	github.com/lthiagol/lazybrew/cmd/lazybrew	[no test files]
ok  	github.com/lthiagol/lazybrew/internal/app	1.083s
ok  	github.com/lthiagol/lazybrew/internal/brew	3.898s
ok  	github.com/lthiagol/lazybrew/internal/config	1.115s
ok  	github.com/lthiagol/lazybrew/internal/gui	1.235s
ok  	github.com/lthiagol/lazybrew/internal/gui/flows	2.830s
ok  	github.com/lthiagol/lazybrew/internal/gui/modal	1.074s
ok  	github.com/lthiagol/lazybrew/internal/gui/presentation	1.122s
ok  	github.com/lthiagol/lazybrew/internal/gui/style	1.074s
ok  	github.com/lthiagol/lazybrew/internal/gui/task	1.108s
?   	github.com/lthiagol/lazybrew/internal/gui/testutil	[no test files]
```

Exit status 0.

### `make cover-check`

```
OK:   github.com/lthiagol/lazybrew/internal/brew coverage 64.3% >= 62%
OK:   github.com/lthiagol/lazybrew/internal/gui/presentation coverage 95.4% >= 91%
OK:   github.com/lthiagol/lazybrew/internal/gui coverage 55.9% >= 36%
OK:   github.com/lthiagol/lazybrew/internal/gui/modal coverage 41.4% >= 41%
All coverage floors met
```

Exit status 0.

---

## Conclusion

`LB-01` is verified. The M13 fix `ac5bac6` is present in all five of its parts,
the four regression tests that cover it pass, and the project gate is green. The
independent review that was never run for M13 is thereby closed together with
`LB-01` (`LB-06`).

No failures, no contradictions between the brief and the repo, and no open
questions were found in this slice.
