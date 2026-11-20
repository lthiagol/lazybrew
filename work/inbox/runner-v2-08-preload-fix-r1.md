# Fix receipt — runner, v2-08-preload r1

## Commits

- F-r1-01  253c3ba  fix(gui): every panel landing fills the dashboard count (preload on)
- F-r1-01  e3c4e08  test(gui): v2-08 frames 15 and 18 show the harness's Taps landing
- F-r1-02  b23a9d8  docs(plan): the v2-08 STATE paragraph matches the recapture

## Findings

- **F-r1-01**: `absorbLanded` runs for every non-Status landing with the preload on.
  `m.landed` keeps those counts, and a Status landing fills unknown counts from them
  (`fillUnknown`). The race is real: Status reads outdated from the cache at the end
  of its fan-out, so a lazy Outdated landing can come first and be overwritten.
  `TestEarlySwitchCountReachesTheDashboard` covers both orders on `View()`; it
  fails without the fix.
- **F-r1-02**: the STATE paragraph describes the committed frames (a8a1c38) and marks
  the slow capture superseded. `work/inbox/runner-v2-08-preload.md` is not rewritten;
  this receipt and the STATE paragraph supersede it. Q-v2-08-b stays open.

## Side effect

Frames 15 and 18 at 140x40 changed: harness step 13 injects a Taps landing with an
empty `Taps` list, which now reaches the dashboard (Taps 0). Other frames are
byte-identical. Regenerated with the same harness in e3c4e08.

## Gate (after 253c3ba)

- `make build`, `make lint`, `make cover-check`: pass.
- `CGO_ENABLED=1 go test ./... -race -count=1`: every package `ok`.
