verdict: changes-required

# Reviewer — v2-09-main-layout, round r1

Range reviewed: `2a7e2a5..228b145` (fd1bde3, 12a3b1b, 95766a6, 539a515, 7f6e93b, d61ea76, 228b145).

## Verify

- **V1 gate:** `make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` every package ok · `make cover-check` "All coverage floors met". Each of the five code commits, checked out alone: `go vet ./...` + `go test ./... -count=1` exit 0.
- **V2 frames:** harness (`frame_dump_test.go.txt`, identical to v2-08's) re-run against real brew, 48.7 s, reads only. All 40 frames exactly W×H (140x40, 80x24). Every `.txt` and `.ansi` byte-identical to the committed `after/`. Before→after diffs: padding shift, tab row and underline shifted together, Log 3 rows, the Deps table. All named in `result.md`. Sidebar, header, footer identical. `internal/gui/dump` removed; worktree clean.
- **Invariants:** no diff in `style/`, `modal/`, `anim/`, `header.go`, sidebar, `go.mod`, `go.sum`, `Makefile`, `scripts/`. No new brew call: the table cells come from the loaded lists. Changed test expectations check the same facts (`hasRow`, `renderContent`, the wrapped log path, 15/15 rows). None removed or skipped.

## Acceptance criteria

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 padding | met | `TestMainAreaHasInnerPadding` covers every non-overlay state, the Operation running/completed/failed states included. Frames: content, tab row and `━━` underline all start at column 1. The tab click subtracts `mainPadX`. `contentLast` uses `mainContentWidth`. |
| AC-02 tables | **not met** | Helper, Muted header (ANSI `38;2;108;112;134`), groups and drop order are OK. But the last column is never cut with `…` when it has a droppable column. See F-r1-01. |
| AC-03 two-column Info | met | `presentation.Info` splits at content ≥ 98 (box inner ≥ 100). The 200x60 frames 03, 06, 17 and 19 split, and the right column wraps with the hanging indent. Below the threshold it falls through to `Detail` unchanged (140x40 and 80x24 frames are identical in content). Q-v2-09-a is correctly raised. |
| AC-04 Log box | met | `logHeight`: 3 idle and empty, `min(bodyH/5,7)` idle with lines (5 body rows), `bodyH/5` running, 0 collapsed. `TestLogBoxHeightByState` asserts the height and main + Log = body at 80x24, 140x40 and 200x60. No animation. Min terminal 80x24 keeps `bodyH/5 ≥ 4`, so no hidden-while-running edge. |
| AC-05 evidence | met with F-r1-02 | Harness copied. `before/` = v2-08 `after/`. `after/` has `.ansi` at 140x40. STATE row is exact. `result.md` is per screen. |

## Findings

**F-r1-01 (major) — internal/gui/table.go:75 — a long description drops the whole Description column instead of being cut with `…`.**
`tableWidths` drops columns while the *natural* width (the widest cell) exceeds `w`. The last kept column is cut only once nothing droppable is left. So a single description longer than the slack removes Description from every row, even with ~60 cells free.
Evidence: a throwaway test ran `renderTable(pkg cols, rows, 87)`, which is the 140x40 content width. Rows were `ncurses / Text-based UI library` and `openssl@3 / <100-char desc>`. Output: `Name  Version` only, with no Description column. On this machine, by an estimate from `brew info --json=v2 --installed` with the Version column at 7 cells, 19 of the 55 installed formulae with installed deps exceed 87 natural cells (e.g. `zstd`, 91). Their Deps tab shows no descriptions at 140x40.
AC-02.1 asks that the last text column take the slack and be cut with `…`.
**The fix must achieve:**
- A column is dropped only when it cannot get a minimum useful width (state it in `result.md`).
- Otherwise the last kept column is cut with `…` to the slack.
- The drop order stays Description, then Version.
- A unit test covers a long description at 87 cells: it is kept and cut with `…`.

**F-r1-02 (minor) — work/verification/v2-09-main-layout/result.md:7 — the commit table claims "one per AC, in order", but git order is AC-03 (95766a6), AC-04 (539a515), AC-03 (7f6e93b).**
AC-03 has two commits, and the second one lands after AC-04. The table lists 7f6e93b before 539a515. This breaches the "one commit per AC, in order" invariant. Rebasing is not allowed, so the record must say what happened.
**The fix must achieve:** `result.md` (and the STATE row note if any) lists the commits in true git order and names the breach.

**F-r1-03 (minor, efficiency) — internal/gui/depviews.go:120 — `pkgRow` rebuilds the whole installed map (`installedList`) for every row.**
A Taps Formulae table of N names with M installed formulae costs N×M map inserts per render. `renderContent` also runs twice per frame (`View` and `contentLast`). Today's taps are small, so this is not visible yet.
**The fix must achieve:** the installed map is built once per table (per group check), not once per row.
