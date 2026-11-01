# Runner receipt — lb03-dashboard-outdated-count

Commits, in AC order: `8a7b191` (AC-01) · `4ae32bd` (AC-02) · `99e54b7` (AC-03) ·
`45fc3f8` (AC-04) · `9fd23dd` (AC-05) · `adb8418` (AC-06). One extra non-AC commit:
`310256c` (question log only).

**The two renderings** (em dash = U+2014):

- not loaded: `Outdated:  — not loaded (press R to load)`
- known: `Outdated:  0 packages` / `Outdated:  7 packages`

## AC-01 — the formatter expresses "not loaded"

`presentation.OutdatedInfo{Count, Known}` added; `FormatStatusDashboard` takes it
instead of an `int`; `Known` → `Outdated:  %d packages`, otherwise the not-loaded
line. Both callers updated (`formatters_test.go`, `snapshot_test.go`) plus new
known / not-loaded / known-zero cases. Evidence: `grep -rn "not loaded"
internal/gui/presentation/formatters.go` → `formatters.go:123`.

## AC-02 — the dashboard computes known from unknown

`statusDashboard(client, loadOutdated)` and `fetchStatusData(client, loadOutdated)`.
Cache-only mode: known only when `KeyOutdatedFormulae` **and** `KeyOutdatedCasks`
are cache hits (an empty slice hit is a known zero; a partial hit is not known).
Load mode: both outdated entries read through the brew reader with
`boundedReadIn(totalCtx, brewReadTimeout, "outdated", …)`, joining the existing
fan-out; no new timeout var. A failed/timed-out outdated read adds a `⚠ outdated:`
line and falls back to not-known, never to 0.

## AC-03 — R loads it, nothing else does

`RefreshMsg{LoadOutdated bool}`. The R key returns `RefreshMsg{LoadOutdated: true}`
(`gui.go:590`, `gui.go:595`); the handler passes it through (`gui.go:475`); Init
passes `false` (`gui.go:143`). Every other emitter stays `RefreshMsg{}` — the
auto-refresh tick (`gui.go:181`), install/uninstall (`gui.go:302`, `gui.go:345`)
and task completion (`gui.go:385`).

## AC-04 — a status refresh stops discarding the outdated cache

`invalidatePanelCache`'s `PanelStatus` case no longer invalidates
`KeyOutdatedFormulae` / `KeyOutdatedCasks`; the comment says the dashboard reads
outdated from the cache and does not own that class. `PanelOutdated` and the
mutation-level invalidations are unchanged.

## AC-05 — the tests

`TestStatusDashboardDoesNotShellCallOutdated`: the no-shell-call assertion is
**byte-identical** to before; only the rendering expectation changed to the
not-loaded line (plus a "no `0 packages`" check and a comment saying why). New:
known-zero from an empty cached slice; partial cache is not known; explicit load
= exactly one `brew outdated` call + real count; cold load = one call per class;
failed load → not-loaded + `⚠ outdated:`; plain `RefreshMsg{}` → **zero**
`brew outdated` calls; `RefreshMsg{LoadOutdated: true}` → one call + real count;
R key sets the flag; the auto-refresh tick does not. Snapshot covers the
not-loaded rendering. Every wait is bounded (mock runner + synchronous cmds).

## AC-06 — bookkeeping

`work/backlog.md`: LB-03 `status` → `closed`, one-line note added under the
table. `work/STATE.md`: row `| lb03-dashboard-outdated-count | — | implemented |
awaiting reviewer |` added.

## Gate (real output)

`make build` → ok · `make lint` → ok · `go test ./... -race -count=1` → all
packages ok · `make cover-check` → "All coverage floors met" (brew 64.3%/62%,
presentation 96.2%/91%, gui 58.4%/36%, modal 41.4%/41%).

## What the next session must know

1. **Commit boundaries.** AC-01 also touched the single production caller in
   `commands.go` (transitional `Known: true`) so the tree kept compiling; the
   known/unknown logic is in AC-02. AC-02 also carried the compile-fix call
   sites (`gui.go` Init/handler + `fetch_timeouts_test.go`, `gui_test.go`), so
   the Init call site reads `fetchStatusData(m.client, false)` from AC-02 on
   instead of AC-03 — the end state matches G-3. Between AC-02 and AC-05 the
   tree had exactly one expected failure (the `0 packages` assertion AC-05
   replaces); the gate was run at the end and is green.
2. **AC-05.5 wording.** A cold load costs one `brew outdated` call *per class*
   (formulae + casks are two readers with two cache keys); one combined call
   would require `internal/brew/**`, which the brief forbids. Both readings are
   covered by two tests. Raised as Q-04 in `work/questions.md`; the
   orchestrator owns the answer.
3. **Extra edit in AC-06.** The "Next slice" paragraph of `work/STATE.md` still
   called LB-03 "blocked on a product decision"; it now says the slice is
   implemented. No table row other than the new one was touched.
4. **Untouched, as required:** `internal/brew/**`, the tab-content path,
   `fetchPanelData`, the Outdated panel's own behaviour, the mutation
   invalidations, `master-plan/`, the timeout values.
