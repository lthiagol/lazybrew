# Verification — LB-01: the M13 Deps-tab fix

**Slice:** `lb01-m13-deps-verify`
**Backlog item:** `LB-01` (`work/backlog.md`)
**Fix under verification:** `ac5bac6` — *fix(gui): stop Formulae Deps infinite loading (M13)*
**Verified on:** 2026-10-31, branch `wip`
**Result:** all five invariants present; each of the five invariants is covered by
at least one test that **fails when the invariant is broken** (mutation evidence
below); all four gate commands pass. No failure to report.

This is a verification record only. No code, test, Makefile or workflow file was
changed in this slice. The mutations below were temporary experiments on
implementation files only: each was reverted with `git checkout --` immediately
after its test run, and no test file was edited at any point.

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

## Mutation evidence — do the tests actually bite?

The presence checks above prove the five invariants are *in the tree*. They prove
nothing about whether the tests would *catch* a regression. This section closes
that gap: each invariant was broken on purpose, its covering test was run, and
the **actual failing output** is recorded verbatim below.

Every mutation was applied to an **implementation file only**, was reverted with
`git checkout --` right after its test run, and left the worktree clean. No test
file was edited. The gate was re-run after all five mutations and was green.

### M1 — breaks I-1 (no context timeout on the tab-content fetch)

Mutation — `internal/gui/commands.go:833`, the deadline is removed:

```diff
 func fetchTabContentCmd(client *brew.Client, panel PanelID, tab int, name string) tea.Cmd {
 	return func() tea.Msg {
-		ctx, cancel := context.WithTimeout(context.Background(), fetchTabContentTimeout)
+		ctx, cancel := context.WithCancel(context.Background())
 		defer cancel()
```

Note on the exact form: the brief describes this as "replace with
`context.Background()`". A literal `ctx, cancel := context.Background()` does not
compile (`context.Background()` returns a single value), so the deadline was
removed with `context.WithCancel(context.Background())` instead. The effect on
the invariant is identical — the fetch is no longer bounded by
`fetchTabContentTimeout`.

Command:

```bash
go test ./internal/gui/ -run TestFetchTabContentCmdTimesOut -count=1
```

Actual output:

```
--- FAIL: TestFetchTabContentCmdTimesOut (2.00s)
    gui_test.go:1310: fetchTabContentCmd hung past fetchTabContentTimeout (infinite-Loading reproducer)
FAIL
FAIL	github.com/lthiagol/lazybrew/internal/gui	2.177s
FAIL
```

Exit status 1. **The test bites.** I-1 is covered by
`TestFetchTabContentCmdTimesOut` (`internal/gui/gui_test.go:1276`).

Reverted with `git checkout -- internal/gui/commands.go`.

### M2 — breaks I-2 (the `DataLoadedMsg` re-kick is removed)

Mutation — `internal/gui/gui.go:455-456`, the re-kick of `loadTabContent()` is
deleted:

```diff
 		if msg.PanelID == m.activePanel && needsTabFetch(m.activePanel, m.activeTab) {
-			return m, m.loadTabContent()
+			// MUTATION-M2: re-kick removed
 		}
 		return m, nil
```

Command:

```bash
go test ./internal/gui/ -run 'TestDepsTabPopulatedAfterDataLoaded|TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply' -count=1
```

Actual output:

```
--- FAIL: TestDepsTabPopulatedAfterDataLoaded (0.00s)
    gui_test.go:1187: expected Deps tabContent populated after DataLoadedMsg, got cmd=nil content="" ok=false
--- FAIL: TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply (0.00s)
    gui_test.go:1220: expected loadTabContent cmd re-issued for Used By (no fast path) after DataLoadedMsg
FAIL
FAIL	github.com/lthiagol/lazybrew/internal/gui	0.182s
FAIL
```

Exit status 1. **Both tests bite.** I-2 is covered by
`TestDepsTabPopulatedAfterDataLoaded` (`internal/gui/gui_test.go:1166`) and
`TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply` (`:1202`).

Reverted with `git checkout -- internal/gui/gui.go`.

### M3 — breaks I-3 (the Deps fast path never applies)

Mutation — `internal/gui/commands.go:795`, the fast-path condition is made
permanently false:

```diff
 func renderCachedTabContent(panelID PanelID, tabIdx int, panel *panelData) (string, bool) {
-	if panelID == PanelFormulae && tabIdx == 2 {
+	if panelID == PanelFormulae && tabIdx == 2 && false {
 		f := panel.selectedFormula()
```

Command:

```bash
go test ./internal/gui/ -run 'TestDepsFastPathRendersCachedDependencies|TestDepsTabViewLeavesLoadingAfterFastPath|TestDepsTabLoads' -count=1
```

Actual output:

```
--- FAIL: TestDepsFastPathRendersCachedDependencies (0.00s)
    gui_test.go:1329: expected nil cmd from fast path (Dependencies already on model)
--- FAIL: TestDepsTabViewLeavesLoadingAfterFastPath (0.00s)
    gui_test.go:1433: expected fast-path nil cmd
--- FAIL: TestDepsTabLoads (0.00s)
    gui_test.go:1489: Deps body missing pcre2:
        ╭──────────────────────────────────╮╭────────────────────────────────────────────────────────────────────────────────╮
        │1 Status                          ││Formulae › Deps                                                                 │
        │(empty)                           ││  List    Info    Deps    Used By    Caveats    Files
FAIL
FAIL	github.com/lthiagol/lazybrew/internal/gui	0.172s
FAIL
```

Exit status 1. **All three tests bite.** I-3 is covered by
`TestDepsFastPathRendersCachedDependencies` (`internal/gui/gui_test.go:1314`),
`TestDepsTabViewLeavesLoadingAfterFastPath` (`:1418`) and `TestDepsTabLoads`
(`:1446`).

Reverted with `git checkout -- internal/gui/commands.go`.

### M4 — breaks I-4 (`parseFormula` reads build deps from the wrong field)

Mutation — `internal/brew/formulae.go:318`:

```diff
-	buildDeps := append([]string(nil), f.BuildDependencies...)
+	buildDeps := []string(f.Dependencies)
```

Command:

```bash
go test ./internal/brew/ -run 'TestParseFormulaBuildDepsFromBuildDependencies|TestFormulaeReaderList' -count=1
```

Actual output:

```
--- FAIL: TestFormulaeReaderList (0.00s)
    formulae_test.go:155: python BuildDeps = [openssl xz zlib], want [pkg-config] (from build_dependencies, not dependencies)
    formulae_test.go:169: fish BuildDeps = [pcre2], want [cmake]
--- FAIL: TestParseFormulaBuildDepsFromBuildDependencies (0.00s)
    formulae_test.go:195: BuildDeps = [runtime-a], want [build-only]
    formulae_test.go:199: BuildDeps must not contain runtime dependency runtime-a
FAIL
FAIL	github.com/lthiagol/lazybrew/internal/brew	0.173s
FAIL
```

Exit status 1. **Both tests bite.** I-4 is covered by
`TestParseFormulaBuildDepsFromBuildDependencies` and `TestFormulaeReaderList`
(`internal/brew/formulae_test.go`). As noted below, the unanchored `-run` pattern
also selected `TestFormulaeReaderListCached` and `TestFormulaeReaderListEmpty`;
those two still passed under M4 because they make no dependency assertion. The
two tests that assert on build deps are the ones that bite.

Reverted with `git checkout -- internal/brew/formulae.go`.

### M5 — breaks I-5 (the key handler drops the fetch command)

Mutation — `internal/gui/gui.go:625-630`, the `"j"/"down"` handler no longer
returns the fetch command:

```diff
 		case "j", "down":
 			m.panels[m.activePanel].down()
-			if needsTabFetch(m.activePanel, m.activeTab) {
-				return m, m.loadTabContent()
-			}
+			// MUTATION-M5: fetch command dropped
 			return m, nil
```

Command:

```bash
go test ./internal/gui/ -run TestDepsSelectionChangeUpdatesContent -count=1
```

Actual output:

```
--- FAIL: TestDepsSelectionChangeUpdatesContent (0.00s)
    gui_test.go:1403: expected zlib deps after j, ok=false content=""
FAIL
FAIL	github.com/lthiagol/lazybrew/internal/gui	0.181s
FAIL
```

Exit status 1. **The test bites.** I-5 is covered by
`TestDepsSelectionChangeUpdatesContent` (`internal/gui/gui_test.go:1370`).

Reverted with `git checkout -- internal/gui/gui.go`.

### Mutation summary

| mut | invariant | mutation | test(s) that failed | exit |
|---|---|---|---|---|
| M1 | I-1 | `commands.go:833` — deadline removed | `TestFetchTabContentCmdTimesOut` | 1 |
| M2 | I-2 | `gui.go:455-456` — re-kick deleted | `TestDepsTabPopulatedAfterDataLoaded`, `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply` | 1 |
| M3 | I-3 | `commands.go:795` — fast path never applies | `TestDepsFastPathRendersCachedDependencies`, `TestDepsTabViewLeavesLoadingAfterFastPath`, `TestDepsTabLoads` | 1 |
| M4 | I-4 | `formulae.go:318` — build deps from `dependencies` | `TestParseFormulaBuildDepsFromBuildDependencies`, `TestFormulaeReaderList` | 1 |
| M5 | I-5 | `gui.go:625-630` — fetch command dropped | `TestDepsSelectionChangeUpdatesContent` | 1 |

**No blocker.** Every one of the five mutations made its covering test fail, so
every one of the five invariants has a test that bites. After the last revert,
`git status` reported a clean working tree, and the gate below was run on that
clean tree.

---

## The regression tests

These four are the tests the slice contract named. They are green — but on their
own they only cover I-3, I-4 and I-5; they stay green when I-1 or I-2 is
broken. The **coverage** claim rests on the mutation evidence above, not on
these four.

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
ok  	github.com/lthiagol/lazybrew/internal/gui	3.082s
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
ok  	github.com/lthiagol/lazybrew/internal/brew	0.068s
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

All four commands were run on 2026-10-31, on the clean tree after every mutation
had been reverted. Every one passed.

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
ok  	github.com/lthiagol/lazybrew/internal/app	1.310s
ok  	github.com/lthiagol/lazybrew/internal/brew	4.123s
ok  	github.com/lthiagol/lazybrew/internal/config	1.283s
ok  	github.com/lthiagol/lazybrew/internal/gui	1.473s
ok  	github.com/lthiagol/lazybrew/internal/gui/flows	3.049s
ok  	github.com/lthiagol/lazybrew/internal/gui/modal	1.240s
ok  	github.com/lthiagol/lazybrew/internal/gui/presentation	1.328s
ok  	github.com/lthiagol/lazybrew/internal/gui/style	1.355s
ok  	github.com/lthiagol/lazybrew/internal/gui/task	1.328s
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

`LB-01` is verified. What is now established, and only what is now established:

- **The fix is present in all five parts.** The presence checks above
  reproduce every anchor named in the brief: I-1 `internal/gui/commands.go:22`
  and `:833`; I-2 `internal/gui/gui.go:455-456`; I-3 `internal/gui/commands.go:779`,
  `:794` and `:810`; I-4 `internal/brew/formulae.go:318` (field `:64`); I-5
  `internal/gui/gui.go:627-634`.
- **Each of the five invariants is covered by at least one test that fails when
  the invariant is broken.** The five mutations in this record were performed
  and each made its covering test fail:

  | invariant | test that fails when the invariant is broken |
  |---|---|
  | I-1 | `TestFetchTabContentCmdTimesOut` |
  | I-2 | `TestDepsTabPopulatedAfterDataLoaded`, `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply` |
  | I-3 | `TestDepsFastPathRendersCachedDependencies`, `TestDepsTabViewLeavesLoadingAfterFastPath`, `TestDepsTabLoads` |
  | I-4 | `TestParseFormulaBuildDepsFromBuildDependencies`, `TestFormulaeReaderList` |
  | I-5 | `TestDepsSelectionChangeUpdatesContent` |

  The four contract-named tests (`TestDepsTabLoads`,
  `TestDepsSelectionChangeUpdatesContent`,
  `TestParseFormulaBuildDepsFromBuildDependencies`, `TestFormulaeReaderList`)
  cover only I-3, I-4 and I-5; they stay green when I-1 or I-2 is broken. The
  I-1 and I-2 coverage comes from the three additional tests named in the table
  above, and those are exactly the tests that bite under M1 and M2.
- **The gate is green.** `make build`, `make lint`,
  `go test ./... -race -count=1` and `make cover-check` all exited 0 on the
  clean tree.

Every mutation was reverted; `git status` was clean at the end of the run, and
no code or test file is changed by this record.

No failures and no open questions were found in this slice.

