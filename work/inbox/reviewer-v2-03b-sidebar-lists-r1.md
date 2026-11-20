verdict: changes-required

# reviewer r1 — v2-03b-sidebar-lists (reviewed `fe12a58..8a9f7ef`)

F-orch-01..03 (Q-v2-03b-a/c/d) are the orchestrator's and not repeated here; Q-v2-03b-b accepted.

## Verify

- **V1 gate** (HEAD): `make build` ok, `make lint` ok, `CGO_ENABLED=1 go test ./... -race -count=1` every package ok, `make cover-check` brew 72.9 / presentation 96.2 / gui 77.5 / modal 50.5, all floors met. Per commit (`0fa39ff`, `4b16b65`, `c30158e`, `33f6348`, each in a detached worktree): `go build ./...` + `go test ./internal/gui/...` pass, `TestEveryFrameIsTerminalSize` included.
- **V2 frames**: committed harness run against real Homebrew 7.0.15-36. `diff -r` against `after/` gives no differences (the only extra files are the 80×24 `.ansi`, which are not committed by design). Every `.txt` frame is exactly W×H. `internal/gui/dump` removed, worktree clean.
- **V3**: `grep renderList|listWindow internal/gui/*.go` finds no matches; both functions are deleted. Frames 01/02/06: the main area shows the dashboard, Info, and the tap summary plus details. No list is drawn twice. The Status sidebar holds only `statusRows`: version, prefix, and the known counts.
- **V4**: the `List` tab is gone. The `formula*Tab` constants are used in `needsTabFetch`, `loadTabContent`, `renderCachedTabContent`, `fetchTabContentCmd`, and `render.go`. Grepping for bare `activeTab ==`/`tab ==` numbers finds none (only the generic `activeTab = 0` in `switchPanel`). Deps/Uses fetch tests now use the constants. `sortFormulae` is unchanged, and the leaves filter narrows `formulae` with `items` (`TestLeavesFilterKeepsFormulaeAligned`).
- **V5**: batch handlers are unchanged. `sidebarRow` marks `batch[p.offset+i]`, so marks stay on the right rows after scrolling. `renderOutdatedVersions` shows the item plus `N selected — u to upgrade, U all`.
- **V6**: default 35, clamp 15–50 %, floor 20, cap 60 (`gui.go:1199`). README is updated. Rows come from typed data through `typedRow`. Versions are right-aligned and long names end in `…`. No `...` is left in sidebar rows. In the ANSI frames, the focused row (06-taps line 22) has background `49;50;68` on every cell; the unfocused Formulae row has the marker only.
- **V7**: in the render and layout files, panel fields are written only in Update-loop methods (`normalize`, `selectIndex`, `syncVisibleRows`). `contentScroll` is written only in `scrollContent`, which is called from the wheel handler.
- **V8**: `sidebarBoxAt` is shared by click and wheel. The wheel over main calls `scrollContent(±3)`, and tab clicks are unchanged. No `List` mention remains in help or hints.
- **V9**: `result.md` covers every screen at both sizes. The `not loaded`/`–` label changes at 80×24 follow from the wider sidebar (v2-03 label rule). No restyle of v2-03 helpers, no modal/toast restyle, no new fetch.
- **V10**: `internal/brew/**` untouched, and no keymap file changed. Top-level `func Test` count: 469 → 479. The 11 removed tests each have a named replacement (receipt). The removed `renderList` assertion in `TestBatchSelectionShowsIndicator` is still covered by the sidebar half of the same test. One commit per AC, in order.

## AC table

| AC | verdict | evidence |
|---|---|---|
| AC-01 | met (F-r1-01) | `0fa39ff`; `TestSidebarWidthRule`, `TestSelectedRowHighlightOnlyWhenFocused`; ANSI 06-taps |
| AC-02 | met | `4b16b65`; `table.go` deleted, constants, `TestFormulaeDefaultTabIsInfo`, `TestFormulaInfoShowsTap` |
| AC-03 | met (F-r1-02) | `c30158e`; `TestMainAreaShowsDetailsNotTheList`, `TestHeaderShowsCountsOnFirstLaunch`; frame 01 header `125 formulae  2 casks` |
| AC-04 | met | `33f6348`; `TestWheelOverAnotherBoxFocusesIt`, `TestWheelOverMainScrollsItsContent` |
| AC-05 | met | `020011f`; frames reproduce byte-for-byte; STATE row present |

## Findings

- **F-r1-01 (minor) — commit `0fa39ff` subject.** The subject lacks the `(v2-03b AC-01)` tag that the other four AC commits carry. Evidence: `git log --format=%s fe12a58..HEAD`. History must not be rewritten (no amend/rebase), so the fix is only to note in the fix receipt that `0fa39ff` is AC-01. No code change.
- **F-r1-02 (minor) — `internal/gui/status.go:35`.** The first Status sidebar row reads `Homebrew 7.0.15-36-g1f335ae`, while every other row is a 10-column label plus a value (`Prefix    /home/…`, `Formulae  125`). In frame `after/140x40/01-status.ansi` (rows 3–8) the version breaks the column. Fix: render it as label `Homebrew` + value in the same `%-10s` column, and update `TestStatusRowsAreCompactSummary`.
- **F-r1-03 (minor) — stale comments.** `internal/gui/gui_test.go:1336` still says "List tab (tab 0 is not a fetch tab)", but tab 0 is now Info. `internal/gui/panel.go:74` describes `formulae`/`casks` but sits above `unfilteredFormulae`. Fix: reword the test message to name Info, and move or rewrite the field comment so it describes the field it sits on.
