verdict: ok

# Reviewer — v2-09-main-layout, round r5 (last)

Range reviewed: `92178cb..ef7b519` (5638a03 fix brief r4, a3d3d93, ef7b519).
Gate per Q-v2-builder-speed and its amendment.

## Verify

- **V1 gate (final tree, = a3d3d93 code):** `make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` every package ok, 0 FAIL · `make cover-check` "All coverage floors met". a3d3d93 is the only code commit, so the final-tree run is also its per-commit gate.
- **V2 frames:** harness re-run against real brew, 49.1 s, reads only. All 60 files are byte-identical to the committed `after/`. All 40 `.txt` frames are exactly W×H.
- The r4 change touches only 140x40 frames 06, 14, 19 and 20, all named in `result.md`. The frames, `result.md` and the tests are in the one fix commit, as the brief asked. The receipt is 29 lines, committed once. `internal/gui/dump` removed; worktree clean.
- **Scope:** the diff is `presentation/detail.go`, `info_test.go`, `info_split_test.go`, `services_view_test.go`, frames, `result.md` and the receipt. Nothing else is touched.

## F-r4-01 — fixed

- `Info` now splits only when `fitsWhole(right, rightW)` holds. That function uses the same label-column maths as `Detail`: the widest label plus `:` plus 2. It checks that no space-free word of a non-prose right value is wider than the value column. `Prose` is set on the right fields only after the check, so every path, URL and command is checked.
- Tests: `TestTapInfoKeepsARemoteWhole`, `TestServicesInfoKeepsPathsWhole` (87 cells, real-length paths), `TestFormulaInfoStillSplitsAt87` and `TestInfoFallsBackWhenAPathWouldBreak` all pass.
- `TestTapInfoSplitsWhenItFits` now measures the column within the line. The receipt is right that the old check passed for any layout.
- The `services_view_test` expectation moved from the wrapped half-path to the whole path. It is stronger, not weaker.

## Info views at 140x40, read from the `.ansi` frames

- **03 Formulae Info, 17 Cask Info:** split. Fields on the left; the Description wraps at spaces with the hanging indent. Reads well.
- **06 Tap Info:** single column. `Remote:    https://github.com/danielfoehrkn/homebrew-switch` is whole, and the counts follow it. Reads well.
- **19 Services Info:** single column. Unit file, Command, Working dir and Log are each whole on one row; the actions line is below them. Reads well.
- **200x60** (wider): 06 and 19 still split, with every path and URL whole in the right column.

## Earlier findings

| Finding | Status |
|---|---|
| F-r1-01 | holds — `TestLongDescriptionIsCutNotDropped` passes; `table.go` unchanged since r3 |
| F-r1-02 | holds — `result.md` commit list in git order, breach named |
| F-r1-03 | holds — `TestPkgRowsMatchPkgRow` passes |
| F-r2-01 | holds — `TestDescriptionKeepsItsFloorBesideALongName` passes; 80x24 Deps frames unchanged |
| F-r4-01 | fixed (above) |
| r3 Log padding | holds — `TestLogBoxHasInnerPadding` passes; Log rows padded in every frame |

## Acceptance criteria

| AC | Verdict |
|---|---|
| AC-01 padding (main box and Log box) | met |
| AC-02 tables | met |
| AC-03 two-column Info (inner ≥ 84 per Q-v2-09-a, whole-token rule per F-r4-01) | met |
| AC-04 Log box heights | met |
| AC-05 evidence and bookkeeping | met |

## Findings

None.
