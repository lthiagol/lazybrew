# Fix brief — lb24-copy-and-wording, round r2 (before review)

Ruling: `Q-lb24-y-search` in `work/questions.md`. Tests first, one commit, subject
ends `[Q-lb24-y-search]`.

1. With an Operation overlay open (any status), `y` copies the overlay on every
   panel, Search included, even when the hidden search box exists (not focused).
   With a focused text input anywhere, `y` still types. With no overlay, `y` on
   Search still types, as the builder did.
2. The overlay copy includes the status row (e.g. `Failed: needs sudo · …`,
   elapsed time) after the title, ANSI-free, untruncated, plus the command line
   AC-02 shows under it, then every output line.
3. Overlay hint/footer: when an overlay is open, show `y copy` in its hint line
   where there is room, as the close hint is shown; frame tests at 80x24 and
   140x40 must not clip the existing `Enter … · Esc …` hint. If it does not fit
   at 80 columns, leave it out there and say so.
4. Tests: failed overlay on Search → `y` → payload has the status row and the
   output lines; the same on Formulae; no overlay on Search → `y` types; focused
   input with overlay open does not copy. Revert the overlay-first check: the
   Search test must fail.

Gate: `make test-fast PKG=./internal/gui/...`, then `make ci`. Never launch the
TUI. Append "Fix r2" (≤ 8 lines) to `work/inbox/runner-lb24-copy-and-wording.md`,
commit, do not edit it afterwards; reply with one done line.
