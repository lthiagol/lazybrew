# Fix brief — lb24-copy-and-wording, round r3

Findings: `work/inbox/reviewer-lb24-copy-and-wording-r1.md` (read in full). Tests
first, one commit per finding, subjects end with the finding id.

- **F-r1-01 (medium):** in the real 80x24 frame (`View()` after
  `WindowSizeMsg{80,24}`, not `renderOperation(80, …)`) the failed retry row must
  show `Failed: needs sudo` uncut and the Enter/Esc hint uncut. Put the hint on its
  own line below the status row if it does not fit on one; the command line stays
  below. `y copy` is shown where it fits, and left out at 80x24 if not (say which
  in the receipt). Tests assert on `View()` at 80x24 and 140x40 frames, replacing
  the width-only tests. Name each golden that changes.
- **F-r1-02 (medium):** Logs and Status Config copies hold every line whole: no
  `cutLines`/`…` in the copy path. Test with a log line wider than 400 cells.
- **F-r1-03 (low):** `y` with help open does not copy the hidden panel. Copy the
  help text or do nothing, your pick, and say which; test it.
- **F-r1-04 (low):** README `y` row: on Search `y` types unless an Operation is
  open.

Gate: `make test-fast PKG=./internal/gui/...` per commit, `make ci` before the
receipt. Never launch the TUI. Append "Fix r3" (≤ 10 lines) to
`work/inbox/runner-lb24-copy-and-wording.md`, commit, do not edit it afterwards;
reply with one done line.
