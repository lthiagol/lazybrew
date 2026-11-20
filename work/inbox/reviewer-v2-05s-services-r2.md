verdict: ok

# Review r2 — slice v2-05s-services (wip 7162d41..999ad9a)

**V0 safety:** after the review's own harness run, `systemctl --user show sh.brew.moshi-hook`:
MainPID=800, NRestarts=0, ActiveEnterTimestamp=Wed 2026-11-18 11:37:48 -03 (same as the
pre-slice record); `journalctl --user -u sh.brew.moshi-hook --since 2026-11-20` has no entries.
Harness still sends only `S` then `esc` on Services (no `y`/enter); the r1 change only adds a
7 s sleep and a `ToastTickMsg`. No test builds a real runner. No blocker.
**V1 gate (CGO_ENABLED=1):** build, lint, `go test ./... -race -count=1` (9 pkgs ok), cover-check (brew 76.5, presentation 98.2, gui 83.0, modal 90.1) — all pass.
**V2 frames:** harness re-run against real brew; `diff -r` vs `after/`: identical. 40 `.txt` all
exactly W×H. `before/` vs `after/` differ only in 19-* and 20-*. `internal/gui/dump` removed.

## Earlier findings

| id | status | evidence |
|---|---|---|
| F-r1-01 (major) | fixed | befbf7e: `serviceStoppable` = started/error/scheduled (details.go), matches brew `formula_wrapper.rb` to_hash (only `:none` is not loaded; error/scheduled are loaded) and `cli.rb` stop (acts on `loaded?`); cited in the comment. Guard uses it for S/r; `s start` allowed on any not-running service (brew's `report_service_running_or_loaded?` only refuses a running pid — correct). Key line and footer share `serviceActionHints(s)`. List status is the raw brew string, so `error`/`scheduled` reach the guard. Tests: `TestStopAndRestartOnErrorAndScheduledConfirm` (S/r × error/scheduled: confirm kind/chip, no toast, confirm runs `brew services <verb> <name>` on mock), `TestServiceHintsFollowLoadedState`, `TestServiceInfoKeyLineOffersStartStopRestartForError`, `TestStartOnErrorAndScheduledRunsWithoutConfirm`, `TestServiceStoppableFollowsLoadedState`. `stopped` stays refused per AC-03.1 (runner flagged it for the orchestrator). |
| F-r1-02 (minor) | fixed | a97a13a: failed/cancelled services tasks send `servicesReloadMsg` → invalidate Services keys (list+info) and `fetchPanelData(PanelServices)` only, no `m.refreshing`, so no "Data refreshed". Success keeps `RefreshMsg`. `TestServiceFailureStillReloads` renamed to `TestServiceFailureReloadsWithoutRefreshToast` and strengthened (asserts reload, no RefreshMsg, error toast last); plus cancel/success/no-toast tests. |
| F-r1-03 (minor) | fixed | 190ff4c: harness waits past toast life and ticks; after/80x24/19 shows `S stop  r restart  c cleanup` unobstructed; 19/20 re-captured at both sizes; result.md section added; 01–18 unchanged. |

## New problems in the fix commits

None. Checked: no test removed (one renamed and strengthened, 9 added); no new key; only Services
code paths touched; one commit per finding with the id in the subject; receipt ≤ 25 lines.

Note (not a finding): a `servicesReloadMsg` landing while a user `R` refresh is in flight
counts toward `m.refreshing`, so "Data refreshed" could show one panel early. Cosmetic, rare.
