# Fix receipt — runner, v2-08-preload r3

## Commit

- F-r3-01  39f34d2  docs(gui): result.md names every harness-caused change against before/

## Finding

- **F-r3-01** (docs only): `result.md`'s recapture section said v2-07 showed the
  machine's three taps. It showed the injected one-tap list (`before/` 15 and 17,
  Taps box `homebrew/cask-versions official`). Dropped.
- The section now has a per-frame table of every difference the harness change
  causes against `before/`: Taps box rows in 13 and 15-20; toasts removed; the
  80x24 Formulae cursor row in 15-20. Frames 14 and 16 are not affected by the
  harness; 14's differences are brew data and the preload's counts.
- Two corrections to the reviewer's list, both checked against the files:
  - Toasts are removed in 15, 17 and 18 at **both** sizes (6 frames), not only in
    15/17/18 at 140x40 and 17 at 80x24.
  - At 80x24 the Taps box row also changes in 20 (`▸ h…` to `▸ ⚠…`), and the
    dashboard Taps count is 3 in v2-07 and in these frames (no count change).
- My r2 receipt (ea2f4a8) said the machine's three taps are "as v2-07 did". That is
  wrong, and the committed r2 receipt is not rewritten; this receipt supersedes it.

## Gate (after 39f34d2)

- `make build`, `make lint`, `make cover-check`: pass.
- `CGO_ENABLED=1 go test ./... -race -count=1`: 11 packages `ok`.
- No code or frame change in this round.
