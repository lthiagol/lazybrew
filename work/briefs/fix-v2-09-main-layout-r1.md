# Fix brief — v2-09-main-layout, round r1

Source: `work/inbox/reviewer-v2-09-main-layout-r1.md` (verdict changes-required).
Fix every finding, one commit each (subject names the id), receipt
`work/inbox/runner-v2-09-main-layout-fix-r1.md` (≤ 30 lines, committed once).
The slice brief `work/briefs/v2-09-main-layout.md` still applies (Safety included).

Gate (ruling `Q-v2-builder-speed`): each commit builds and passes
`CGO_ENABLED=1 go test -race` for the packages it touches; the full gate runs
once, before the receipt.

- **F-r1-01 (major)** — `internal/gui/table.go` drops the whole Description
  column when one description is longer than the slack. Drop a column only
  when it cannot get a minimum useful width (state that width in
  `result.md`); otherwise keep it and cut the last column with `…`. Drop order
  stays Description, then Version. Unit test: a long description at 87 cells
  is kept and cut with `…`. Regenerate frames if they change; name changes.
- **F-r1-02 (minor)** — `result.md` lists commits in true git order and
  names the AC-03 / AC-04 order breach (no rebase).
- **F-r1-03 (minor)** — `pkgRow` (`internal/gui/depviews.go`) builds the
  installed map once per table, not once per row.

Reply with exactly one line: `done: work/inbox/runner-v2-09-main-layout-fix-r1.md`.
