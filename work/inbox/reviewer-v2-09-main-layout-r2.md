verdict: changes-required

# Reviewer — v2-09-main-layout, round r2

Range reviewed: `ee49ea4..182ed03` (ec799b4 fix brief, 801054e, 77f0e2c, d2b31bd, f0f223c, 182ed03).
Gate per ruling Q-v2-builder-speed.

## Verify

- **V1 gate (final tree):** `make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` 11 packages ok, no FAIL · `make cover-check` "All coverage floors met".
- **Per-commit gate (ruling):** 801054e and d2b31bd, checked out alone: `go build ./...` + `CGO_ENABLED=1 go test -race -count=1 ./internal/gui/` exit 0. 77f0e2c is docs only.
- **V2 frames:** harness re-run against real brew, 49.5 s, reads only. All 40 frames exactly W×H. Every `.txt` and `.ansi` byte-identical to the committed `after/`. The only frame changes since r1 are `after/80x24/04-formulae-deps.txt` and `11-deps-bash.txt`: Description kept and cut, `Library for command-line …`. Both are named in `result.md` and in the receipt. `internal/gui/dump` removed; worktree clean.
- **Invariants:** no diff outside `internal/gui/{table,depviews,render}.go`, `table_test.go`, `result.md` and `work/inbox/`. No test weakened: two `pkgRow` call sites take the list argument, and two tests were added.

## Earlier findings

| Finding | Status | Evidence |
|---|---|---|
| F-r1-01 (major) | fixed | `minimal()` drops a column only when every kept column at `min(nat, 12)` does not fit. A temp test at 87 cells with a long description kept the column and cut it: `Generate introspection data for GObject li…`. `TestLongDescriptionIsCutNotDropped` asserts kept, cut and 87 wide. The 12-cell minimum is stated in `result.md`. Drop order is unchanged (`TestTableDropsDescriptionThenVersion` passes). |
| F-r1-02 (minor) | fixed | `result.md` lists the commits in git order (rows 1–7). It names the AC-03 refinement after AC-04 as an out-of-order breach and says why it was not rebased. |
| F-r1-03 (minor) | fixed | `pkgRows` calls `installedList` once per group or table. `pkgRow` takes the list. The Taps Formulae table uses `pkgRows`. `TestPkgRowsMatchPkgRow` passes. |

## AC table (unchanged from r1 unless noted)

| AC | Verdict |
|---|---|
| AC-01 | met |
| AC-02 | met with F-r2-01 (edge case below; real data does not hit it) |
| AC-03 | met (Q-v2-09-a is still open for the orchestrator) |
| AC-04 | met |
| AC-05 | met |

## Findings

**F-r2-01 (minor, visual polish) — internal/gui/table.go:81 vs :108–120 — the minimum-width guarantee is not enforced when the widths are laid out.**
`minimal()` keeps the Description because the kept columns fit at 12 cells each. But the cut loop (`:108`) narrows the other columns only until the last column gets **1** cell, and `:120` gives the last column whatever is left. The Name column is never narrowed to make room for the description's 12. The result is the sliver that `tableMinUseful`'s own comment says cannot happen.
Evidence: a temp test ran `renderTable(pkgTableCols, row, 46)` (the 80x24 content width) with `gobject-introspection-runtime / 1.86.0_1 / Generate introspection data…`. Output: `✓  gobject-introspection-runtime  1.86.0_1  G…`, with the header `D…`.
None of this machine's installed dep tables hit it: checked from `brew info --json=v2 --installed` at 46 and 87 cells.
**The fix must achieve:** a kept column is never narrower than `min(natural, tableMinUseful)`. Either cut the other columns down to that floor, or drop the column by the existing order when the floor cannot be met. Add a unit test with a ~29-character name at 46 cells.
