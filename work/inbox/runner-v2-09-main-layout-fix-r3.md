# Receipt: runner-v2-09-main-layout-fix-r3

Brief `work/briefs/fix-v2-09-main-layout-r3.md`. Gate per Q-v2-builder-speed.

## Commits, one per item

- Q-v2-09-a b2ac4f6 — split at main box inner 84 (`splitWidth` = 82 content
  cells). Left column ≥ 2/5 of width, leaves right ≥ `infoMinRight` = 26. Unit
  file (Services) and Remote (Tap Info) moved right; the 140x40 split had cut them.
  Tests: `TestInfoSplitsAt140x40`, Tap Info at 82/81, wrap test.
- 5d22e50 — follow-up: 200x60 frames for that move (left unstaged in b2ac4f6).
- Log padding 5d04885 — Log box gets `mainPadX` either side. Test:
  `TestLogBoxHasInnerPadding`.

## Frames

- Q-v2-09-a: 140x40 changed on 02, 03, 06, 07, 10, 12, 13, 14, 17, 19, 20; 80x24
  unchanged; 200x60 changed on 06, 14, 19, 20.
- Log padding: every frame with a Log box changes in its Log rows (72 files).
- `result.md` names the screens.

## Readability at 140x40 (by eye)

- Formulae and Tap Info read well. Services Info paths wrap mid-word on the right
  (no space to break at); stated in `result.md`, not a cut.

## Gate (final tree, after 5d04885)

- build, lint, `go test -race ./...`, cover-check: all ok.
