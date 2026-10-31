# Backlog — parked work

Work that is real, known, and **not** in any active slice. When a slice picks an
item up, move it into that slice's brief and drop the row here.

The `source` column points at where the item came from — the old `mp` plan, which
is now a frozen local archive (untracked by git, history only, not an input to any
decision). The source is provenance, not a spec.

| id | work | source | priority | status |
|---|---|---|---|---|
| LB-01 | Verify the M13 Deps-tab fix by running the Deps test suite (`TestDepsTabLoads`, `TestDepsSelectionChangeUpdatesContent`) | M13 (execution done, independent review never passed) | medium | closed |
| LB-02 | Apply context timeouts to the four brew fetch paths: `fetchPanelData`, `fetchStatusData`, `fetchSelectedSearchInfo`, `executeSearch` | B-02 (residual from M13 F-02/F-07) | medium | closed |
| LB-03 | Status dashboard outdated count stays 0 until the first refresh | B-03 (M09 F-02 residual) | low | active |
| LB-04 | Chunk `brew tap-info --json` batches when the tap count risks ARG_MAX | B-04 (M11 F-01 residual) | low | active |
| LB-05 | Full interactive human smoke before the stable tag — **human gate**, not an agent task | B-05 | medium | active |
| LB-06 | The M13 independent review itself — closed by verification together with LB-01; all 7 findings were already fixed, so the review is not left silently dropped | M13 | medium | closed-by-LB-01 |
| LB-07 | Apply context timeouts to the two brew reads still unbounded: the uninstall dependency pre-check (`internal/gui/commands.go:1034`, `Formulae.Uses`) and `toggleLeaves` (`internal/gui/commands.go:1128`, `Formulae.Leaves`) | found by lb02-fetch-timeouts | low | closed |
| LB-DROP-01 | Dead header fix in the `master-plan-bkp` directory | B-01 | n/a | dropped |

**LB-DROP-01 reason:** the target directory does not exist on this machine —
so the fix cannot be applied and there is nothing to preserve.

**LB-01 closed** — the M13 review is closed by this verification; the evidence
is in `work/verification/lb01-m13-deps.md`, carried by commit `789d506`.

**LB-02 closed** — the four brew fetch paths are bounded by context
timeouts: `internal/gui/commands.go` (commits `0027596`, `f7bc287`) and the
tests in `internal/gui/fetch_timeouts_test.go` (commit `00f73f2`); the
receipt is `work/inbox/runner-lb02-fetch-timeouts.md`.

**LB-07 closed** — the last two unbounded brew reads (the uninstall dependency pre-check and `toggleLeaves`) are bounded by `brewReadTimeout`, and a timed-out dependency check now says so in the uninstall confirm dialog instead of looking like "nothing depends on this package"; receipt `work/inbox/runner-lb07-bound-remaining-reads.md`.
