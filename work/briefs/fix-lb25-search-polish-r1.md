# Fix brief — lb25-search-polish, round r1 (before review)

Rulings: `work/questions.md` (lb25). One commit, tests first, subject ends
`[Q-lb25-step-tick]`.

The `Completed in …` row uses `Icons.OK` (`✓`); finished steps use `✔`. Make the
Completed row show `✔` too. Check every use of `Icons.OK`: if all are success
ticks of the same meaning, set `OK` to `✔` in both icon sets; if some are other
things (trusted tap, toasts), leave those and give the Completed row the
`Installed` glyph or a new field, whichever is the smaller change. Say which and
list the uses in the receipt. Tests: the Completed row and the steps above it
show the same glyph (`View()` at 80x24 and 140x40); name each golden that
changes. Gate: `make test-fast PKG=./internal/gui/...`, then `make ci`. Append
"Fix r1" (≤ 6 lines) to `work/inbox/runner-lb25-search-polish.md`, commit, do
not edit afterwards; reply with one done line. Never launch the TUI.
