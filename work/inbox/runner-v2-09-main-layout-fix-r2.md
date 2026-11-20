# Receipt: runner-v2-09-main-layout-fix-r2

Brief `work/briefs/fix-v2-09-main-layout-r2.md`. Gate per Q-v2-builder-speed.

## Commit, one per finding

- F-r2-01 e8bde9a — `tableWidths`: every kept column keeps at least
  `min(natural, tableMinUseful)`. Phase 1 cuts the widest other columns (the name)
  down to their floors before the last column gets less than its own. Phase 2, only
  when the floors still do not fit, cuts them to `tableMinW` (4), as before. A
  column is dropped only by the existing order, and only when the floors cannot be
  met.
  Test: `TestDescriptionKeepsItsFloorBesideALongName`: a 31-character name at 46
  cells; the description starts under its header, keeps at least 12 cells, ends
  with `…`, every line 46 wide. It fails on the r1 `table.go` and passes on this one.

## Frames

- Regenerated all 40 `after/` screens and the 200x60 set. Byte-identical to the
  committed `after/`: no frame changes from F-r2-01. `result.md` gains the floor rule.

## Gate (final tree, after e8bde9a)

- `make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`: all ok

Open, not changed here: Q-v2-09-a (140x40 threshold); the Log box is not padded.
