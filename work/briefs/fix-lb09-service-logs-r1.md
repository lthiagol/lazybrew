# Fix brief — lb09-service-logs, round r1

Source: `work/inbox/reviewer-lb09-service-logs-r1.md`. Fix every finding, one
commit each, subject names the id. Brief `work/briefs/lb09-service-logs.md`
(Safety, Invariants, Gate) still applies.

- **F-r1-01 (major)** — `internal/gui/gui.go` ~:278 and ~:280: `return nm,
  tea.Batch(cmd, nm.syncLogs())` depends on unspecified evaluation order.
  Call `syncLogs` into a local first in **both** branches
  (`c := nm.syncLogs(); return nm, tea.Batch(cmd, c)`). Grep the gui package
  for the same pattern (a mutating method call in the same return as the
  model it mutates) and fix every instance; list them in the receipt.
- **F-r1-02 (minor)** — `internal/gui/services_logs_test.go`: add a test with
  two services, each with its own stdout temp file; after `j` the view shows
  B's lines and none of A's, and an old-gen `logsReadMsg` carrying A's lines
  delivered after the switch is dropped (view unchanged).

Gate: `make test-fast PKG=./internal/gui/...` per commit, then `make ci`.
Receipt `work/inbox/runner-lb09-service-logs-fix-r1.md` (≤ 30 lines, committed
once), worktree clean, reply with exactly one line:
`done: work/inbox/runner-lb09-service-logs-fix-r1.md`.
