# runner: lb07-bound-remaining-reads

Commits (AC order): AC-01 `bd6dfd7` · AC-02 `b5c1cc5` · AC-03 `ea71699` · AC-04 `7b21fc0`.

**AC-01.** `confirmUninstall`: the dependency check goes through `boundedRead(brewReadTimeout, "uses", …)` — no new timeout var, no value changed, detection rule untouched. The error is no longer discarded. On a `*brew.TimeoutError` the confirm message gains, inside the closure after the read: *"Dependency check timed out after 30s — the list of packages that depend on ripgrep may be incomplete."* (duration = `timeoutErr.Timeout`; at the production 30s cap it renders "30s"). The dialog still opens in every case; any other error keeps today's behaviour exactly — no dependents, no complaint.

**AC-02.** `toggleLeaves` uses `boundedRead(brewReadTimeout, "leaves", …)`. The `MutationResultMsg` error path is untouched, so a timeout reaches the user through the existing toast as `leaves: brew leaves timed out after 30s`.

**AC-03.** `internal/gui/fetch_timeouts_leaves_test.go`, three tests using `blockingRunner()` and `brewReadTimeout` overridden to 50ms (restored after). All four mutations bite: removing the notice fails the timeout test; reverting either read to `context.Background()` hangs and fails it (`cmd hung past 2s`); making the notice fire on *any* error fails the non-timeout test.

**AC-04.** `work/backlog.md`: LB-07 → `closed` + a one-line note, no other row changed. `work/STATE.md`: row `| lb07-bound-remaining-reads | — | implemented | awaiting reviewer |`.

**Gate — all four pass.** `make build` → 0. `make lint` → 0. `go test ./... -race -count=1` → all packages ok (gui 2.121s). `make cover-check` → "All coverage floors met" (brew 64.3%≥62%, presentation 95.4%≥91%, gui 58.0%≥36%, modal 41.4%≥41%).

**Evidence.** `Formulae.Uses`/`Formulae.Leaves` now at :972 (tab-content path, untouched), :1055 and :1160 (both inside `boundedRead`). The only `context.Background()` left are :74 (bounded-read helper), :762 (task-manager cancel), :936 (tab-content timeout), :1383 (status total budget) — exactly the four the brief allows.

**Next session.** `boundedRead`, `boundedReadIn`, `timeoutError` and all three timeout values are unchanged; `internal/brew`, `internal/gui/gui.go`, the tab-content path, the dashboard fan-out and `master-plan/` were not touched. `LB-04` is next in `work/backlog.md`.
