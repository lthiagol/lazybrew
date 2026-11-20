verdict: changes-required

# Review r1 — v2-02-layout-bugs (wip a5ac841..05aad2f)

Scope: runner commits 2611e00 a9ccbbe a56fd8e 2498d26 23a031c 5525dfd 9046165 8d9af6b 350f24e.
Orchestrator commits 5107598 d155b51 24c5f8b 05aad2f (questions/backlog only) are not reviewed as findings.

## Checks run
- V1 gate: `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check` → exit 0; all packages ok; floors met (brew 72.9, presentation 96.2, gui 74.3, modal 50.5).
- V2 harness copied to `internal/gui/dump/`, run on real Homebrew (`ok 8.2s`), then removed. `diff -r` vs `after/140x40` and `after/80x24`: identical. Cell-width script over all 22 frames: every frame exactly 40×140 / 24×80.
- `TestEveryFrameIsTerminalSize`: 18 states (7 panels, loading, log collapsed, help, 3 modals, toast, op running/done/failed) × 140x40, 80x24, 200x60 + too-small 70x20; exact rows and `lipgloss.Width` per row.
- V3: `render()` builds a fitted base and composites toast/modal via `lipgloss.NewLayer`/`NewCompositor`/`NewCanvas` (`render.go` `compose`). Tests compare every cell outside the overlay; modal centring checked at all three sizes; 80x24 frame 07 starts the box at x=13, y=6 (= (80-54)/2, (24-12)/2).
- V4: `ready`/`Loading...` gate removed from confirm and menu; input never had it; `grep ready|Loading internal/gui/modal` → nothing. Resize while open re-centres (compositor) and keeps 100x30.
- V5: `loaded` set on `DataLoadedMsg` (error or not, gui.go:510) and `SearchDoneMsg` (gui.go:258) — the only two places items land. R-refresh rule at gui.go:573 untouched.
- V6: frames show one border around main, one around the Log; `TestMouseHitsWhatIsDrawn` clicks a drawn Formulae row, tab geometry from `tabBarRows()`; C collapse gives `mainH == bodyH`.
- V7: `brew info --json=v2 bash` → runtime_dependencies objects with `full_name`/`version`/`declared_directly`; fixture matches; `after/140x40/11-deps-bash.txt` shows `Runtime: ncurses, readline`. Fish (uninstalled) falls back to `[pcre2]`. Cask object `depends_on` parses as before (`[]`).
- V8: result.md classifies every changed region; spot-checked 01, 07, 08 at both sizes — no colour/border/glyph/header change found. Q-v2-hints/help/small/cask-deps ruled by the orchestrator (05aad2f).
- V9: tests 410 → 425; removed assertion lines are only the renamed `logLayoutHeights` checks (same intent on `screenLayout`) and the fish deps expectation (AC-07.2); all three named in the receipt. `internal/brew` diff limited to `flexStrings`/`parseFormula`/tests/fixture. No new deps, no keybinding change, no `program.Send`. Worktree clean, `internal/gui/dump` absent.

## AC table
| AC | result | evidence |
|---|---|---|
| AC-01 | pass | size test + 22 real frames exact; see F-r1-01 for the list-fill fix that rides on it |
| AC-02 | pass | compositor layers; `TestModalFloatsOverFrame`; `TestMouseIgnoredWhileModalOpen` |
| AC-03 | pass with F-r1-02 | behaviour correct (frame 07 shows title, question, buttons); test weaker than AC-03.3 asks |
| AC-04 | pass | `TestToastFloatsAboveBottomBar`; frame 10 toast on row h-3, nothing else moved |
| AC-05 | pass | explicit `loaded`; both tests; frames show `not loaded` vs `(empty)` |
| AC-06 | pass | behaviour correct; note: the sibling move landed in the AC-01 commit (disclosed in the receipt); not actionable without a history rewrite |
| AC-07 | pass | deps_test.go; real shape; frame 11 |
| AC-08 | pass | frames reproduce byte-identically; result.md; STATE row present |

## Findings

F-r1-01 (major) — internal/gui/panel.go:252 (`renderList`, `visible := max(1, height)`) — the "main list fills the panel" fix draws `contentH` rows from the **sidebar's** shared `p.offset`, and that offset is still clamped to the sidebar window (`visibleCount`). At 80x24 the fitted sidebar gives the active panel vc=1 (it was 17 at a5ac841 because the old frame overflowed), so every `j` scrolls the main list and the selected row is pinned to its top row with all rows above it hidden although the panel has room. — evidence: throwaway probe (removed), Status with 30 rows, 6× `j`: 80x24 now `vc=1 sel=6 off=6`, main list starts `▸ row-06`; at a5ac841 `vc=17 off=0`, rows 00–12 shown with the cursor on row-06. Same at 140x40 once the cursor passes vc=9 (offset moves while 26 rows are available). Affects every `renderList` panel (Status, Services, Outdated, Search); the Formulae table has its own offset and is fine. `TestMainListFillsMainPanel` only checks rows at selected=0. — fix must achieve: the main list scrolls within its own height (a render-time window derived from `selected` and the main content height, as the Formulae table does), so at 80x24 rows above the cursor stay visible when they fit; a test at 80x24 that moves the cursor down and asserts the earlier rows and the cursor row are both drawn.

F-r1-02 (minor) — internal/gui/modal_first_frame_test.go:15-40 — AC-03.3 asks the first `View()` to contain the confirm's title **and question**; the test asserts only `Uninstall bash`. The trust-menu case asserts `Trust`, which is always on screen on the Taps panel (tab name, panel.go:52), so that case passes without the menu being drawn. — fix must achieve: assert the confirm's question text (`This cannot be undone.`) and, for each menu/input case, a string only the modal draws (e.g. `Trust entire tap:` / a menu item, `Tap repository (user/repo):`).
