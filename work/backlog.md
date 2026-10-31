# Backlog — parked work

Work that is real, known, and **not** in any active slice. When a slice picks an
item up, move it into that slice's brief and drop the row here.

The `source` column points at where the item came from — the old `mp` plan, which
is now a frozen local archive (untracked by git, history only, not an input to any
decision). The source is provenance, not a spec.

| id | work | source | priority | status |
|---|---|---|---|---|
| LB-01 | Verify the M13 Deps-tab fix by running the Deps test suite (`TestDepsTabLoads`, `TestDepsSelectionChangeUpdatesContent`) | M13 (execution done, independent review never passed) | medium | active |
| LB-02 | Apply context timeouts to the four brew fetch paths: `fetchPanelData`, `fetchStatusData`, `fetchSelectedSearchInfo`, `executeSearch` | B-02 (residual from M13 F-02/F-07) | medium | active |
| LB-03 | Status dashboard outdated count stays 0 until the first refresh | B-03 (M09 F-02 residual) | low | active |
| LB-04 | Chunk `brew tap-info --json` batches when the tap count risks ARG_MAX | B-04 (M11 F-01 residual) | low | active |
| LB-05 | Full interactive human smoke before the stable tag — **human gate**, not an agent task | B-05 | medium | active |
| LB-06 | The M13 independent review itself — closed by verification together with LB-01; all 7 findings were already fixed, so the review is not left silently dropped | M13 | medium | active |
| LB-DROP-01 | Dead header fix in the `master-plan-bkp` directory (frozen, untracked, local-only archive) | B-01 | n/a | dropped |

**LB-DROP-01 reason:** the target directory does not exist on this machine — the
`master-plan-bkp` directory is part of the frozen, untracked, local-only archive —
so the fix cannot be applied and there is nothing to preserve.
