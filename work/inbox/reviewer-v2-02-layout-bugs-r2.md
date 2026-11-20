verdict: ok

# Review r2 — v2-02-layout-bugs (fix commits 690e5dd, 4a649cf; receipt 3d7cc96)

## r1 findings
| id | status | evidence |
|---|---|---|
| F-r1-01 (major) | fixed | `renderList` now windows on `listWindow(p.selected, len, height)` (panel.go:256, 290) and no longer reads `p.offset`. My r1 probe re-run (throwaway test, removed): Status with 30 rows, 6× `j` at 80x24 → main list draws `row-00`…`row-05`, `▸ row-06`, …`row-12` (r1: `▸ row-06` pinned on the top row). 26× `j`, 3× `k`, then `C` at 80x24 → cursor `▸ row-23` drawn on the last list row, sidebar shows `▸ row-23`, frame 24×80. 140x40 likewise shows rows 00–05 above the cursor. `TestMainListScrollsWithCursor` covers both sizes and the past-the-height case. |
| F-r1-02 (minor) | fixed | `TestModalsShowContentOnFirstFrame` now wants `Uninstall bash? This cannot be undone.`, `Tap repository (user/repo):`, `Export to Brewfile (brew bundle dump)`, `Trust entire tap: homebrew/foo`, and checks each in the modal's own view as well as the frame — none of these is drawn by the base screen (the old `Trust` was a tab name). |

## New problems introduced by the fixes
None found.
- `listWindow` is pure and clamps: `count==0 → 0`; `count < height → 0`; cursor always within `[start, start+height)`. Nothing is written at render time; the sidebar keeps `p.offset`.
- No main-panel code path other than `renderList` read `p.offset` for these panels: `handleContentClick` is a no-op outside the Formulae table, which has its own window. Batch marks use `idx := start + i`, so `●` still lands on the right row.
- 4a649cf is test-only; no expectation weakened (each `want` is a superset of the old string).

## Re-run checks
- Gate: `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check` → exit 0; floors met (brew 72.9, presentation 96.2, gui 74.4, modal 50.5).
- Frame harness on real Homebrew (`ok 8.1s`, `internal/gui/dump` removed after): 22 frames identical to `after/140x40` and `after/80x24` (`diff -r`), all exact size by the cell-width check. `after/`/`result.md` correctly unchanged.
- Test count 426 (≥ 410). `program.Send` only in a doc comment (task/doc.go:5), none added by the slice. One commit per finding, subject names the id.

## AC table (unchanged from r1, findings closed)
| AC | result |
|---|---|
| AC-01 | pass |
| AC-02 | pass |
| AC-03 | pass |
| AC-04 | pass |
| AC-05 | pass |
| AC-06 | pass (sibling move landed in the AC-01 commit, disclosed; not actionable) |
| AC-07 | pass |
| AC-08 | pass |

## Note
`work/briefs/v2-03-look-foundation.md` and `work/briefs/review-v2-03-look-foundation.md` are untracked in the tree — orchestrator files for the next slice, not the runner's; left untouched.
