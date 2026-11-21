# Fix brief — lb29-footer-redesign, round r2

Findings: `work/inbox/reviewer-lb29-footer-redesign-r1.md` (read in full); rulings at the
end of `work/questions.md`. Tests first, one commit per item, subjects end `[LB-26]`; every
commit builds and passes `make test-fast PKG=./internal/gui/...`.

- **F-r1-01 (medium):** no help line clipped at 80x24, 100x30, 140x40, 200x50 that was whole
  on the base commit (`8af0ef1`). The widened key column (`j/k up/down`, `h/l left/right`)
  must not push descriptions off the line: shorten the key texts (`j/k ↑↓`, `h/l ←→`) or
  re-balance the column, your pick. Help still lists j/k, h/l, Shift+Tab. Test: no `…` in
  any help frame at those four sizes (both themes) — fails on HEAD.
- **F-r1-02 (low):** the footer shows `tab next` (lowercase, as the brief); help may keep
  `Tab`. Update `flows/footer_test.go`, README and goldens.
- **F-r1-03 (low, ruled):** the modal key row uses the muted `·` separator
  (`[y]es · [n]o · esc cancel`), at the same dim role as the footer. Frame test.
- **F-r1-04 (low, orchestrator):** a key listed in both groups shows once. On Status (and an
  open overlay on Status) `R refresh` appears in the panel group and the global group at 140
  and 200 columns: keep it in the global group only. Test: no key appears twice in one row
  on any panel at 140 columns.

Gate: `GOFLAGS=-p=1 make ci` before the receipt. Never launch the TUI. Append "Fix r2" (≤ 8
lines) to `work/inbox/runner-lb29-footer-redesign.md`, commit, do not edit afterwards; reply
with one done line.
