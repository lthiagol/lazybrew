# Fix brief — v2-09-main-layout, round r2

Source: `work/inbox/reviewer-v2-09-main-layout-r2.md`. One commit (subject
names the id), receipt `work/inbox/runner-v2-09-main-layout-fix-r2.md`
(≤ 30 lines, committed once). Gate per ruling `Q-v2-builder-speed`.

- **F-r2-01 (minor)** — `internal/gui/table.go`: a kept column can end up
  narrower than `tableMinUseful` (e.g. `G…` for Description at 46 cells with a
  29-char name). A kept column is never narrower than
  `min(natural, tableMinUseful)`: cut the other columns down to that floor, or
  drop the column by the existing order when the floor cannot be met. Unit
  test with a ~29-char name at 46 cells.

Reply with exactly one line: `done: work/inbox/runner-v2-09-main-layout-fix-r2.md`.
