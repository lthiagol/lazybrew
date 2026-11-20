# Receipt: runner-v2-09-main-layout-fix-r1

Brief `work/briefs/fix-v2-09-main-layout-r1.md`; reviewer `work/inbox/reviewer-v2-09-main-layout-r1.md`.
Gate per Q-v2-builder-speed: touched packages per commit, full gate once here.

## Commits, one per finding

- F-r1-01 801054e — a column drops only when the kept columns, each at its minimum
  useful width (12 cells, `tableMinUseful`), do not fit. A long description is cut
  with `…`. Drop order unchanged: Description, then Version.
  Test: `TestLongDescriptionIsCutNotDropped` (87 cells, kept, cut, lines 87 wide).
- F-r1-02 77f0e2c — result.md lists commits in git order; names the AC-03/AC-04
  breach (7f6e93b after 539a515). No rebase, no amend.
- F-r1-03 d2b31bd — `pkgRows` reads the installed list once per table; `pkgRow`
  takes the list. Test: `TestPkgRowsMatchPkgRow`.

## Frames

- F-r1-01 changed `after/80x24/04-formulae-deps.txt` and `11-deps-bash.txt`
  (Description kept, cut: `Library for command-line …`). All other frames unchanged,
  including every `.ansi`, the 140x40 set and the 200x60 set.
- F-r1-03 changes no frame: a regenerated run matches the committed `after/` byte for byte.
- `result.md` drop rule and rows 04/11 updated.

## Gate (final tree, after d2b31bd)

- `make build` ok; `make lint` (go vet) ok
- `CGO_ENABLED=1 go test ./... -race -count=1`: every package ok
- `make cover-check`: All coverage floors met

## Open

- Q-v2-09-a (140x40 threshold) still open for the orchestrator.
- The Log box is still not padded; this round did not touch it.
