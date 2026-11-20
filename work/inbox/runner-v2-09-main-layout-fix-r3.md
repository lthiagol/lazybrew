# Receipt: runner-v2-09-main-layout-fix-r3

Brief `work/briefs/fix-v2-09-main-layout-r3.md`. Gate per Q-v2-builder-speed.

## Commits, one per item

- Q-v2-09-a b2ac4f6 — Info split at main box inner width 84 (`splitWidth` = 82
  content cells, one named constant). 140x40 now splits. To keep split views
  readable there: the left column is at least 2/5 of the width and leaves the
  right column `infoMinRight` = 26 cells; Unit file (Services) and Remote (Tap
  Info) moved to the right column, since the 140x40 split had cut them.
  Tests: `TestInfoSplitsAt140x40`; Tap Info at 82/81; wrap test finds its column.
- 5d22e50 — follow-up: the 200x60 frames for that move (I had left them unstaged).
- Log padding 5d04885 — the Log box gets `mainPadX` either side, like the main
  box. Test: `TestLogBoxHasInnerPadding`.

## Frames (regenerated after each item)

- Q-v2-09-a: 140x40 changed on 02, 03, 06, 07, 10, 12, 13, 14, 17, 19, 20. 80x24
  unchanged. 200x60 changed on 06, 14, 19, 20 (the moved values).
- Log padding: every frame with a Log box changes in its Log rows (72 files);
  the rest of each frame is unchanged.
- Named in `result.md` per screen.

## Readability at 140x40 (checked by eye in the frames)

- Formulae Info: name and fields left; Description wraps on the right.
- Services Info: fields left; paths on the right wrap at the column edge, mid-word
  where a path has no space. Stated in `result.md`, not a cut.
- Tap Info: fields left; Remote and counts on the right.

## Gate (final tree, after 5d04885)

- `make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`,
  `make cover-check`: all ok.
