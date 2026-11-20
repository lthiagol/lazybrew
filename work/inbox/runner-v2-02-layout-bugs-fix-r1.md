# Fix receipt — runner, v2-02-layout-bugs, round r1

Commits, one per finding, in order:
- F-r1-01 690e5dd — `fix(gui): F-r1-01 scroll the main list on its own cursor window`
- F-r1-02 4a649cf — `test(gui): F-r1-02 assert text only the modal draws`

## F-r1-01 (major) — main list windowed by the sidebar offset
- Cause: `renderList` drew from `p.offset`, which the sidebar clamps to its own
  window (`visibleCount`). At 80x24 that window is 1 row, so the cursor sat on
  the top row and rows above it were hidden.
- Fix: `listWindow(selected, count, height)` in `panel.go`, a pure function of the
  cursor and the main content height, as the Formulae table does. `renderList`
  no longer reads `p.offset`; the sidebar keeps its own scrolling. Nothing is
  written at render time.
- Test: `TestMainListScrollsWithCursor` (80x24 and 140x40). Cursor moves down 6
  rows: rows 00..06 and `▸ row-06` drawn. Cursor moves past `contentH+2`: that
  row is drawn in the main panel and the sidebar scrolls to it.
- Evidence it catches the bug: on the committed `renderList` the test fails with
  `cursor on row-06: main panel should show "row-00" above the cursor`.

## F-r1-02 (minor) — modal test asserted text the frame already had
- `TestModalsShowContentOnFirstFrame` now asserts a string only the modal draws
  per case: `Uninstall bash? This cannot be undone.`, `Tap repository (user/repo):`,
  `Export to Brewfile (brew bundle dump)`, `Trust entire tap: homebrew/foo`.
  Each is checked in the modal's own view and in the frame.
- Probe (removed): with Taps open and no menu, `Trust entire tap` is absent.

## Frames
- Harness re-run against real Homebrew: 22 frames compared byte-for-byte with
  `after/140x40` and `after/80x24`. None changed, so `after/` and `result.md` are
  untouched.

## Gate (tree at 4a649cf)
- `make build`: ok; `make lint`: clean
- `CGO_ENABLED=1 go test ./... -race -count=1`: all nine packages ok
- `make cover-check`: All coverage floors met (brew 72.9, presentation 96.2, gui 74.4, modal 50.5)
- Tests: 425 → 426 (`TestMainListScrollsWithCursor`). No test changed expectation or
  was removed in this round.

State: worktree clean; `internal/gui/dump` absent; nothing pushed.
