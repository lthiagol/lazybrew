# Fix receipt — runner, v2-08-preload r2

## Commits

- F-r2-02  3eed39a  fix(gui): a failed panel read forgets its count on the dashboard
- F-r2-01  f6107b4  test(gui): v2-08 frames 15 and 18 show the machine's taps

## Findings

- **F-r2-02**: a failed landing makes its source's count unknown in `m.landed` and
  in the dashboard, and drops its cache entry, so a later Status load cannot show
  the older count. Test `TestFailedReadForgetsItsCount` asserts on `View()`, before
  and after a Status landing. Mutation check: without the change it fails.
- **F-r2-01**: the injected Taps landing now has the production shape (`Taps`
  consistent with `Items`). After frame 13 a `RefreshMsg` lands the real Taps read,
  so frames 15 on show the machine's three taps (as v2-07 did). Frames regenerated
  with the same harness: 18 files changed against e3c4e08. result.md ("Recapture
  after F-r2-01") names each change:
  - 13 (both sizes): the injected tap's trust and detail rows (harness setup only).
  - 15, 18 (140x40): dashboard Taps 3 (was 0) and the Taps box.
  - 17, 19, 20 (140x40), 15-20 (80x24): the Taps box lists the real taps; the
    80x24 Formulae cursor row shows differently (overlay or scroll), selection kept.

## Gate (final tree, after f6107b4)

- `make build`, `make lint`, `make cover-check`: pass.
- `CGO_ENABLED=1 go test ./... -race -count=1`: 11 packages `ok`.
- No brew write ran; no test deleted. STATE.md is not in this round's brief.
