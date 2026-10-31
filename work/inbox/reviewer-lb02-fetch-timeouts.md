# reviewer: lb02-fetch-timeouts

All evidence below is what I ran myself, not what the receipt claimed.

**Gate (V8), run by me — all pass:** `make build`; `make lint` (`go vet ./...`); `go test ./... -race -count=1` (9/9 packages ok); `make cover-check` (brew 64.3%≥62, presentation 95.4%≥91, gui 57.4%≥36, modal 41.4%≥41). `gofmt -l .` empty. `go test ./internal/gui/ -race -count=5` ok.

**V1/G-1:** `brewReadTimeout` 30s (`commands.go:28`), `brewSearchTimeout` 15s (`:33`), `brewStatusTotalTimeout` 45s (`:41`) — package-level `var`s. Every read in the four paths goes through `boundedRead`/`boundedReadIn` (`:1252,1264,1277,1278,1295,1306,1374-1391`). Remaining `context.Background()` hits are only `:73` (the helper), `:915` (tab content), `:1033`/`:1127` (LB-07 reads), `:741` (`serviceCleanup` write) — no unbounded read left in the four paths.

**V2/G-2:** `statusDashboard` (`:1349`) parents every read in `context.WithTimeout(Background, brewStatusTotalTimeout)`; `boundedReadIn` trims each read by the parent's remaining budget (`:90-94`) — two distinct bounds, the total is not the per-call cap. With cap 30s < total 45s the per-read cap binds first in production and the total is the backstop; the aggregate test covers the total-is-tighter case.

**V3/G-3:** `timeoutError` (`:60`) detects `ctx.Err() == context.DeadlineExceeded` and returns `&brew.TimeoutError{...}` — the existing type (`internal/brew/errors.go:42`), reused. Carriers: `DataLoadedMsg.Err`, the status ⚠ lines + aggregated errs, `SearchDoneMsg.Err`, `SearchInfoLoadedMsg.Err`; all reach the user (`gui.go:220-238,435-445`). `Diagnostics.Config`'s error still ignored (`:1388`).

**V4/AC-03:** six tests in `internal/gui/fetch_timeouts_test.go`, named as briefed; every wait is bounded (`awaitMsg`); each asserts `errors.As(*brew.TimeoutError)` *and* that the message names command and duration (`assertTimeout`). Not vacuous — see M1–M5.

**V5/G-4 + scope:** `diff` of `fetchTabContentCmd` pre/post prints nothing; `fetchTabContentTimeout` untouched. `git diff --name-only 1ec84ec..16c24f4` = exactly `internal/gui/commands.go`, `internal/gui/fetch_timeouts_test.go`, `work/backlog.md`, `work/STATE.md`; nothing in `internal/brew/`, `go.mod`/`go.sum`, `.github/`, `Makefile`, `scripts/`; `git diff 1ec84ec..16c24f4 -- internal/gui/gui_test.go` prints nothing.

**V6/success path:** same `FormatStatusDashboard` args; `""` + `FormatDoctorStatus` still appended after the items; ⚠ order formulae, casks, taps, services, doctor — identical to pre-slice and confirmed empirically in the M1 output; outdated still cache-only (`:1400-1410`); config error still ignored.

**V9/AC-04:** LB-02 `closed` + note; LB-07 added (low, active, "found by lb02-fetch-timeouts") with the **current** lines 1034/1128 (the brief's 951/1045 were pre-slice numbers); LB-01/03/04/05/06 and LB-DROP-01 untouched; `work/STATE.md` row matches the contract.

**Mutations — all reverted, `git status` clean.** M1 `errors.Is(err,…)` instead of `ctx.Err()` → 5 tests FAIL. M2 parent with no deadline → `TestFetchStatusDataTotalBudgetBoundsTheAggregate` FAILs (2.0s vs 150ms). M3 per-read deadline removed → panel+search tests FAIL on the bounded wait ("cmd hung past 2s"). M4 raw `err` instead of `*brew.TimeoutError` → 5 tests FAIL on `errors.As`. M5 tab-content `WithTimeout` removed → `TestFetchTabContentCmdTimesOut` FAILs. The runner's mutation claims are accurate.

**V7 — the concurrency decision: sound, and the right shape.** Each goroutine writes only its own local and `wg.Wait()` (`:1393`) joins before any read; shared state is `brew.Cache` (RWMutex, `cache.go:46`), the singleflight (Mutex, `singleflight.go:11`), a stateless `DefaultRunner` and write-once TTL fields — race detector clean over 5 runs. Concurrency is *required* by G-2, not a free choice: sequential reads trimmed by the remaining budget would let one hung source eat 30s of the 45s and leave the last read ~0. Cost: six brew processes at once (`brew info --json=v2 --installed`, `…--installed --cask`, `brew tap`, `brew services list --json`, `brew config`, `brew doctor`); the two `brew info` calls each load the full Homebrew API JSON, so the peak is two API parses plus doctor. Acceptable: Init and manual `R` only (auto-refresh defaults off), each process capped at 30s, dashboard at 45s. ⚠ order is fixed by construction (sources appended after the join), never by completion order.

**F-01 | minor |** `boundedReadIn` reports a **negative** duration when the parent's budget is already spent: `commands.go:91` sets `budget = time.Until(parentDeadline)` unclamped. Probe (temp file, deleted): expired parent → `brew list timed out after -1s`. Latent — today's four paths start all six reads at once, so none reaches it. Fix: clamp `budget` to `>= 0`.

**F-02 | minor |** the fan-out puts both heavy `brew info --json=v2` API loads and `brew doctor` in the same instant. Correct and bounded, but a higher peak than the sequential shape. Optional fix: sequence the two `brew info` reads or run doctor outside the fan-out (doctor is cached only on success, so a failing doctor re-runs every refresh).

**Questions:** `work/questions.md` — none open, none answered, none missed; this slice raises none.

Verdict: ok
