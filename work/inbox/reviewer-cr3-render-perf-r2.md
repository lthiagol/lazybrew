verdict: ok

# reviewer — cr3-render-perf r2

Scope: dc4df13..4cc3e6f. The r1 slice plus fix commits b014689 (F-r1-01) and d19018e (F-r1-02), and receipt 4cc3e6f.

## Gate (V1)
- `make ci` exit 0: fmt-check, vet, `go test ./... -race -count=1` (12 ok), all coverage floors met (gui 87.6% ≥ 83%), build.
- `GOOS=darwin go build ./...` ok.
- The fix commits touch only `scroll.go`, `motion.go` and `gui.go` (one field), plus two test files. No golden or teatest expectation changed. Each finding has its own commit with its id in the subject.

## Earlier findings
| id | status | evidence |
|---|---|---|
| F-r1-01 (blocker) | fixed | `cachedTab` (`scroll.go:57`) returns no rows when `opState != nil`, the panel is loading, or no item is selected. Those are the three early returns in `renderContent` that draw something other than the tab under the same key. Revert check: b014689^'s `scroll.go` in a temp worktree at HEAD. All 4 new tests fail there: `operation_main_cache_test.go:28` (overlay hidden), `:52` (frozen frame after close), `:63` (loading hidden), `:81` (`mainRows` during the operation). All green at HEAD. |
| F-r1-02 (minor) | fixed | `toastMotionNames` is built in `motionToasts` on each stack layout. `toastOffset` builds no string. Revert check: d19018e^'s `motion.go` and `gui.go` → `hot_path_test.go:154: reading a toast's offsets allocated 2 times, want none`. A toast missing from the map reads `Get("")` = 0, which is what an unseen name read before. The existing toast motion tests are unchanged and green. |

## Frames (V3): my harness from r1, extended
164 frames: dark and light, 80x24 and 140x40, 41 states. That is the 23 r1 states (operation-over-Info harness bug fixed) plus 18 new ones:
- Files: cached → loading → loaded; resize bigger and back; theme switched to the other one mid-run and back; selection j then k back to a cached item.
- Used By: plain, operation running over it, after the operation, after a reload that removes installed formulae, leaves on and off.
- Two toasts after a resize.
- C11: search with a brew error, then info.

Each state was dumped with `View()` at dc4df13 and at HEAD in temp worktrees. **160 are byte-identical. Only the 4 `x-search-error-C11` frames differ**, and only in the text: `Parse error: invalid character 'E' looking for beginning of value` → `Error: No available formula with the name "zz".`

Every frame is exactly h rows, and every row is exactly w cells after ANSI strip (0 bad rows).

## What renderContent reads that the cache key might ignore
I checked each input on the Files and Used By paths of `renderContent` against the cache.
- **Covered by the gate in `cachedTab`:**
  - panel and tab;
  - `opState`, loading, and a missing selection;
  - content not loaded yet, so the spinner is never cached.
- **Covered by the key or the entry:**
  - selection (`contentKey` holds the item name);
  - width (key);
  - theme (`*Theme` pointer);
  - tab content (`src` compared);
  - installed list (`installedGen`, bumped by `indexInstalled` on landing and on both leaves branches).
- **Fixed for the whole run:** `m.icons` and `iconsOff`, which come from config.
- **Not on this path:** `renderUses` reads only `depFormula`, so the casks list is not involved. Search results belong to another panel, and the Deps, Info, Caveats, Casks and Status tabs are not cached.

Resize also clears the map. I found no stale path.

## Benchmarks (V6), re-run at HEAD after the fix
- FilesTabPageDown: 2.70–2.75 ms, 1.0 MB, about 2 295 allocs. The cache still holds: `TestScrollingFilesTabRendersOnce` is green.
- FormulaePanelView: 2.41 ms, 2 493 allocs.
- OperationRenderPerLine was unchanged by the fixes. r1 measured 2.55 s against 128.6 s at dc4df13.

## New findings
None.

## Notes (not findings)
- On the cancel path, `readChunk` may drop lines that are still buffered (the brief allows this). Step titles are not capped. `updateOutput` has no bound (predates the slice).
