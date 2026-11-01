# Reviewer receipt — lb03 fix cycle (slice `lb03-dashboard-outdated-count`)

All three findings (F-01, F-02, F-03) are answered with biting tests, the gate passes, and the M9 invariant is now structurally pinned. Contract: `work/briefs/fix-lb03-dashboard-outdated-count.md`; re-review brief: `work/briefs/review-fix-lb03-dashboard-outdated-count.md`.

## Evidence (command → result)
- **S1 (F-01)** `cat internal/gui/outdated_count_test.go` → `TestNonRRefreshEmittersStayCacheOnly` asserts on the **emitted message** (`refresh.LoadOutdated` must be false) for `TaskCompletedMsg` ×2 and `MutationResultMsg` ×2, via `collectRefreshMsgs` → `drainCmds`. `go test ./internal/gui/ -run TestNonRRefreshEmittersStayCacheOnly -count=1 -v` → **PASS** (4/4 subtests).
- **S2 (F-02)** `go test ./internal/gui/ -run TestStatusRefreshKeepsOutdatedCache -count=1 -v` → **PASS**: primes `KeyOutdatedFormulae` **and** `KeyOutdatedCasks`, sends a plain `RefreshMsg{}`, asserts both keys still survive, 0 `brew outdated` calls, and the dashboard renders `Outdated:  1 packages` (not the not-loaded rendering).
- **S3 (F-03)** `grep -n -A28 'case "R":'` → `gui.go:591-594`: the "no update running" branch now returns `tea.Batch(m.tasks.RunNext(), func() tea.Msg { return RefreshMsg{LoadOutdated: true} })`. `grep -n -A20 'msg.ID == "brew-update"'` → `gui.go:302` stays **plain** `RefreshMsg{}`; that handler is shared with the Init path (`StartUpdateMsg` → `gui.go:258-268`), so the count is never loaded at startup (M9). **No invalidation added:** the `R` path runs `m.client.Runner.ExecuteStream` (`gui.go:577`); only `diagnosticsWriter.Update` (`internal/brew/doctor.go:217`) calls `InvalidateFor("update")` — so the loaded count sticks.
- **S4 (scope)** `git diff --name-only 80065c4~1..93a036b` → `internal/gui/gui.go` + `internal/gui/outdated_count_test.go` only; the same diff restricted to `work/ internal/brew/ internal/gui/presentation/ internal/gui/modal/ internal/gui/task/ go.mod go.sum .github/ Makefile scripts/` → **empty**. (`git diff --name-only 3cf36b7..93a036b` also lists `work/briefs/fix-lb03-dashboard-outdated-count.md` — that is the orchestrator's own brief commit `1f58194`, not the fix lane.) `commands.go:29,34,42` → timeouts still 30s/15s/45s. The formatter, known/unknown logic, `RefreshMsg` plumbing and `invalidatePanelCache` are untouched — the only `gui.go` hunk is the `R` branch.
- **S5 (gate, run by me)** `make build` ok · `make lint` ok · `go test ./... -race -count=1` → all packages ok · `make cover-check` → brew 64.3/62, presentation 96.2/91, gui 59.3/36, modal 41.4/41 — "All coverage floors met".

## Mutations (all reverted; `git status` clean)
| id | mutation | result |
|---|---|---|
| M1 | `gui.go:345` + `:385` → `LoadOutdated: true` | **FAILS** `TestNonRRefreshEmittersStayCacheOnly` (all 4 subtests) — bites |
| M2 | outdated keys back into `invalidatePanelCache`'s `PanelStatus` | **FAILS** `TestStatusRefreshKeepsOutdatedCache` (both keys + the render) — bites |
| M3 | `R` branch → `return m, m.tasks.RunNext()` | **FAILS** `TestRKeyLoadsOutdatedWhenUpdateOnStart` ("expected a RefreshMsg from the R key") — bites |
| M4 | `gui.go:302` → `LoadOutdated: true` | **FAILS** `TestBrewUpdateCompletionStaysPlain` — bites |

## Findings
- None. F-01, F-02 and F-03 are each answered by a test that fails when the behaviour is reverted. Every emitter is now pinned: tick `gui.go:181`, brew-update completion `:302`, task completion `:345`, mutation result `:385`, `R` path `:593/:599/:604`.

## Questions
- **Q-04 — answered, ruling holds.** The per-class reading remains the contract (`work/questions.md`); the fix cycle touched neither the outdated readers nor `internal/brew/**`. **No new questions** are raised by this cycle.

Verdict: ok
