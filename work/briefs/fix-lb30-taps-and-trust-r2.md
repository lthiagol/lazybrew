# Fix brief — lb30-taps-and-trust, round r2 (orchestrator-verified)

Findings: `work/inbox/reviewer-lb30-taps-and-trust-r1.md`. Subject ends with the finding id.

- **F-r1-01 (medium):** a test that asserts on the **footer row** (the last row of `View()`)
  with a trusted item selected: `t untrust` leads the panel group, and no `x`, `r`, `T`
  hints; on a tap row the old hints. It must fail when the `selectedTrusted()` branch of
  `tapsHints` is removed (show that run in the receipt). Also make the keymap test cover the
  item-row hints (they answer `t` only on an item row).
- **F-r1-03 (low):** the fix round's receipt section ≤ 8 lines.

Gate: `GOFLAGS=-p=1 make ci`. Never launch the TUI. Append "Fix r2" to
`work/inbox/runner-lb30-taps-and-trust.md`, commit, do not edit afterwards; reply with one done line.
