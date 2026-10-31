# reviewer: lb07-bound-remaining-reads

The slice meets the contract. No findings.

## Evidence (command → result)

- **V1 / G-1** — `grep "Formulae.Uses\|Formulae.Leaves" internal/gui/commands.go` → `:972` (tab-content, untouched), `:1055` (`boundedRead(brewReadTimeout, "uses", …)`), `:1160` (`boundedRead(brewReadTimeout, "leaves", …)`). Timeout vars unchanged: `brewReadTimeout` 30s, `brewSearchTimeout` 15s, `brewStatusTotalTimeout` 45s (only line numbers shifted by the new `errors` import). `timeoutError` and `boundedRead` diff clean against `4b90480` — detection rule still `ctx.Err() == context.DeadlineExceeded`. No new timeout var.
- **V2 / G-2** — `internal/gui/commands.go:1053-1073`: the error is no longer discarded (`uses, err := boundedRead(...)`); on `errors.As(err, &timeoutErr)` the message gains, **inside the closure**, `"\n\nDependency check timed out after " + timeoutErr.Timeout.String() + " — the list of packages that depend on " + name + " may be incomplete."` (at the 30s cap: *"Dependency check timed out after 30s — the list of packages that depend on ripgrep may be incomplete."*). Any other error takes the untouched path — no dependents, no complaint. `DepCheckMsg` is returned on every branch, and `internal/gui/gui.go:401-405` always opens the confirm modal, so a timeout never swallows the flow.
- **V3 / G-3** — `internal/gui/commands.go:1157-1167`: the read is bounded and the error still flows through the existing `MutationResultMsg{Name: "leaves", …, Err: err}`; no new message type, no new toast. `internal/gui/gui.go:382-383` renders `leaves: brew leaves timed out after 30s`.
- **V4 / AC-03** — `internal/gui/fetch_timeouts_leaves_test.go` has all three tests, all passing under `-race -count=2`; every wait bounded by `awaitMsg(…, 2*time.Second)`; `brewReadTimeout` overridden to 50ms and restored via `defer` in each test. No `t.Parallel()` in the package, so the global override is safe.
- **V5 / G-4 + scope** — `git diff --name-only 4b90480..7b21fc0` → `internal/gui/commands.go`, `internal/gui/fetch_timeouts_leaves_test.go`, `work/STATE.md`, `work/backlog.md`. Forbidden paths diff empty. `fetchTabContentCmd` diff empty (tab-content untouched). `statusDashboard` fan-out unchanged. `grep "context.Background()"` → exactly four hits: `:74` helper, `:762` task-manager cancel, `:936` tab-content timeout, `:1383` dashboard total budget.
- **V6 / success paths** — the dependents-listing block and the `Uninstall …? This cannot be undone.` / `Zap …?` wording are unchanged; `toggleLeaves`' success branch is byte-identical.
- **V7 / gate, run by me** — `make build` → 0; `make lint` (`go vet ./...`) → 0; `go test ./... -race -count=1` → all packages ok (gui 2.308s); `make cover-check` → "All coverage floors met" (brew 64.3%≥62%, presentation 95.4%≥91%, gui 58.0%≥36%, modal 41.4%≥41%).
- **V8 / AC-04** — `git show 7b21fc0`: `LB-07` status cell → `closed` with a one-line note; no other backlog row touched; `work/STATE.md` row is exactly `| lb07-bound-remaining-reads | — | implemented | awaiting reviewer |`.

## Mutations (all reverted; worktree clean)

| id | mutation | result |
|---|---|---|
| N1 | removed the timeout notice from the confirm message | `TestConfirmUninstallDepCheckTimesOut` **FAILS** ("does not report the timeout and its duration", "may be incomplete") |
| N2 | uninstall read back to bare `context.Background()` | `TestConfirmUninstallDepCheckTimesOut` **FAILS** ("cmd hung past 2s") |
| N3 | `toggleLeaves` read back to bare `context.Background()` | `TestToggleLeavesTimesOut` **FAILS** ("cmd hung past 2s") |
| N4 | notice fires on any error, not only a timeout | `TestConfirmUninstallDepCheckNonTimeoutError` **FAILS** ("a non-timeout error must not be reported as a timeout") |

Every mutation was reverted; `git status --porcelain` is empty.

## Questions

`work/questions.md` has no open questions; none was answered, missed, or newly raised by this slice. The brief's line references were slightly stale (it cites `commands.go:1053-1054` / `:1145-1148`; the actual pre-slice lines were `:1053` and `:1147` — same code, no contradiction).

## Process observation (not a defect)

The runner committed its receipt three times while reflowing it under the 35-line limit: `275f2df` (41 lines), `85b83dd` "tighten" (40 lines), `2ddef85` "reflow" (17 lines). History is not rewritten and this is not a finding against the slice — but the runner should have measured the line count before committing.

Verdict: ok

