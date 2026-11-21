# Fix brief — lb31-cleanup-tabs, round r2 (orchestrator-verified)

Findings: `work/inbox/reviewer-lb31-cleanup-tabs-r1.md` (read in full). Tests first, one commit
per finding, subjects end with the finding id.

- **F-r1-01 (medium):** `cutMiddle` uses the room it has: snap the cut to a `/` only when one is
  within a few cells (say ≤ 4) of the exact cut point, otherwise cut exactly at the width, so
  distinct paths stay distinct and the last component (formula / file name) survives where it
  fits. Tests at the 80x24 content width with the reviewer's paths (`<prefix>/Cellar/python@3.12/
  3.12.4_1`, two different `~/Library/…bottle.tar.gz` names, `<prefix>/Cellar/llvm/17.0.6` with
  the 15-wide Files column): the rows differ and keep the identifying part. They must fail on HEAD.
- **F-r1-02 (low):** a fixture line `Would remove: … (1,700 files, 64.4MB)` asserting `Files: 1700`;
  mutation (`atoiComma` without `ReplaceAll`) must fail it.
- Rename `TestTabWindowAt60Columns` to say it is a 60-cell content width (60x24 terminals show the
  too-small screen).

Gate: `GOFLAGS=-p=1 make ci`. Never launch the TUI. Append "Fix r2" (≤ 8 lines) to
`work/inbox/runner-lb31-cleanup-tabs.md`, commit, do not edit afterwards; reply with one done line.
