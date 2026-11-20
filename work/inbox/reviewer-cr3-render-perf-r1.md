verdict: changes-required

# reviewer — cr3-render-perf r1

Scope: dc4df13..795d46e (d8650ab ec0df54 99b737b c043ba2 8da6ed8 1ebf46c 5010219, receipt 795d46e).

## Gate (V1)
- `make ci` exit 0: fmt-check, vet, `go test ./... -race -count=1` (12 ok), all coverage floors met, build.
- `GOOS=darwin go build ./...` ok. One commit per AC, in AC order; no goldens or teatest expectations changed.

## AC table
| AC | verdict | evidence |
|---|---|---|
| AC-01 | ok | `renderSearchInfo` reads `searchInfo`/`searchInfoErr` only, no Unmarshal. Reverted `searchInfoErr(err)` → `err` in a temp worktree: `TestSearchInfoBrewErrorIsShownAsBrewSaysIt` fails (`search_info_test.go:188`). Frames: search info, cask, parse error and empty are byte-identical. |
| AC-02 | ok | `addStepLine` gives the same state as `parseSteps` (table test, checked after every line). The cap drops `Lines` only; progress, bar and `motionProgress` read `op.steps` (no `parseSteps` caller left outside tests). The cap test keeps 1999/2000 after the header is dropped. Bench re-run: 128.6 s / 29.2 GB / 726 M allocs → 2.55 s / 205 MB / 5.48 M. |
| AC-03 | ok | Pre-fix `readChunk` (adapted to `Lines`) on HEAD tests: `manager_test.go:479: iteration 269: 18 of 20 lines arrived before the completion`. HEAD: `-race -count=3` ok. Drain cannot hang: every producer closes `outCh` (`defer close(out)`, `closedCh`, `ExecuteStream` closes stdout after err; wrappers forward `lines` untouched), and ctx cancel ends it. Cancel path: `select` may take `ctx.Done` with lines still buffered and drop them. The brief allows this ("or until ctx is cancelled"). |
| AC-04 | **changes-required** | F-r1-01. Bench re-run (FilesTabPageDown): 125 ms / 48.3 MB / 222 875 allocs → 2.73 ms / 1.0 MB / 2 293. P7 (`indexInstalled` on landing and on leaves toggle, `installedGen` invalidates Used By) ok. |
| AC-05 | ok | `Len`/`Last` tests (empty, partial, wrapped, 0, negative). `logHeight` uses `Len`, `View` uses `Last(h)`. One `screenLayout` per `render`. Frames unchanged. |
| AC-06 | ok | `lipgloss.NewStyle().Foreground(style.*Color)` sites: 28 → 4 (box edge ×2, 2× non-bold Accent), as reported. Bench re-run (FormulaePanelView): 8da6ed8 2.59 ms / 3365 allocs → HEAD 2.49 ms / 2493 allocs (dc4df13 2.90 ms). Both themes byte-identical in my frame diff. |
| AC-07 | ok, one minor | `needsTabFetch` switch (0 allocs). `helpRows` memo keyed by width and theme pointer: help content does not depend on the panel, `ApplyTheme` runs only at startup and every call makes a new `*Theme`, so no stale path. `naturalWidth` ok. `sortFormulae` order locked, including case-only names. Toast id unique (all toasts come from `NewToastAt`). See F-r1-02. |

## Frames (V3), my own harness
92 frames (dark and light, 80x24 and 140x40, 23 states: status, formulae, Files at top / PgDn×3 / G, operation over Files (running, failed, closed), next item, Used By, operation over Info (running, done), casks, leaves, help, help j×2, search info / cask / parse error / empty, outdated, log toggle). Dumped with `View()` at dc4df13 and at HEAD in temp worktrees. Every frame has exactly h rows. **84 are identical; the 8 "operation over Files" frames (running, failed × 2 sizes × 2 themes) differ.** That is F-r1-01, not C11.

## Findings
F-r1-01 (blocker) — `internal/gui/scroll.go:63` `mainRows`, also through `contentLast` (`scroll.go:110`) — The cache stores whatever `renderContent` returns whenever `cachedTab()` is true. But `renderContent` returns `renderOperation` when `m.opState != nil`, and spinner + " Loading..." when the panel is loading. The cache key ignores both.
- (a) Files/Used By tab drawn, then an operation starts: the whole run (and its failed or cancelled state) shows the Files rows instead of the Operation overlay.
- (b) Operation starts while the tab is uncached: its first frame is cached and frozen. After the operation closes, that stale operation frame is shown in place of the Files tab until the tab content or the installed list changes.
- (c) Panel loading over a cached tab shows the old rows instead of Loading (and a cached spinner never moves).
- `G`/End and `contentLast` during an operation measure the wrong rows.
- Evidence: probe tests on HEAD (`View()` after `TaskStartedMsg` + `TaskOutputMsg` over `filesTabModel`; `opState = nil` after; `loading = true`) fail. The same tests pass at dc4df13. Frame diff above.
- The fix must achieve: use the cache only when `renderContent` would actually draw the Files/Used By tab (no `opState`, panel not loading), or key it on that state, so frames match dc4df13. Add regression tests that assert on `View()` for an operation over a cached and an uncached Files tab (start, stream, fail, close), and for loading over a cached tab. The tests must fail on c043ba2's `mainRows`.

F-r1-02 (minor) — `internal/gui/motion.go:126` `toastName` — still builds a string per call (`"toast:" + FormatUint(id) + ":" + part`), several times per toast per frame. P12 asks for no string joins. The fix must achieve: motion keys for a toast built once (e.g. names kept on the toast or memoised by id), or a `Q-cr3-...` question if `anim.Set`'s string keys make that unreasonable.

## Notes (not findings)
- Step titles are not capped (one `opStep` per `==>` line, ≤ 2 output lines each). The receipt says so; the brief caps `Lines` only.
- `updateOutput` (brew update) grows without a bound. This predates the slice and is out of scope.
