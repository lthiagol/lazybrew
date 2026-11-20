# Receipt: runner-v2-09-main-layout-fix-r4

Brief `work/briefs/fix-v2-09-main-layout-r4.md`. Gate per Q-v2-builder-speed.

## Commit

- F-r4-01 a3d3d93 — `Info` splits only when `fitsWhole` holds: no non-prose
  right-column value has a space-free word wider than its value column. Otherwise
  the single column. Prose still wraps; Formulae and Cask Info still split at 140x40.

## Tests

- `TestTapInfoKeepsARemoteWhole`, `TestServicesInfoKeepsPathsWhole`: 87 cells,
  paths and remote whole, single column kept.
- `TestFormulaInfoStillSplitsAt87`: prose still splits.
- `TestInfoFallsBackWhenAPathWouldBreak` (presentation).
- `TestTapInfoSplitsWhenWide` is now `TestTapInfoSplitsWhenItFits`: the old test
  measured the column against the whole output, so it passed for any layout.
- `TestServiceInfoShowsDetailAndRunningActions`: at 100 the log path is whole.

## Frames

- Changed, 140x40 only: 06 taps, 19 services info (single column now), and the
  overlays over them, 14 and 20.
- Unchanged: 02, 03, 17 (still split); all 80x24; all 200x60.

## Gate (final tree, after a3d3d93)

- build, lint, `go test -race ./...`, cover-check: all ok.
