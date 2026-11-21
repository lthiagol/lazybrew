# Fix brief — lb26-list-filter, round r2 (orchestrator-verified)

Findings: `work/inbox/reviewer-lb26-list-filter-r1.md` (read in full). Tests first,
one commit per finding, subjects end with the finding id.

- **F-r1-01 (medium):** from Formulae/Casks, `7` then `/` must land on Search with
  the box focused and **no `/` typed**: a `/` pressed while the Search box is
  focused and empty only keeps it focused (it never types `/`; a non-empty box
  types `/` as before). Test asserts `searchInput.Value()==""` and focused, and
  fails with the fix reverted; update `TestSlashStillSearchesOnOtherPanels`.
- **F-r1-02 (low):** the header badge shows the full installed total for
  Formulae/Casks while a query or leaves narrows the list (`fullFormulae()`/
  `fullCasks()`), with a test.

Gate: `GOFLAGS=-p=1 make test-fast PKG=./internal/gui/...` per commit, `make ci`
before the receipt (use `GOFLAGS=-p=1`). Never launch the TUI. Append "Fix r2" (≤ 8
lines, with the revert evidence) to `work/inbox/runner-lb26-list-filter.md`,
commit, do not edit afterwards; reply with one done line.
