# Receipt: lb22-update-default-and-leftovers (runner)

## Commits (one per AC, in order)

- AC-01 `fc82630` lb22 AC-01: update_on_start defaults to true
- AC-02 `dcc09ce` lb22 AC-02: search errors surface, stale re-marks are dropped
- AC-03 `6c5ee76` lb22 AC-03: synctest for coalescing tests, debug log closed on exit

## Gate

- `make ci` (after AC-03): exit 0. All 14 packages `ok`; `All coverage floors met`.
  Floors unchanged: app 70.0 ≥ 64, brew 88.8 ≥ 78, config 93.1 ≥ 85, gui 90.8 ≥ 83.
- `GOOS=darwin go build ./...`: exit 0.
- `deadcode ./cmd/...`: no output, exit 0.
- `staticcheck ./...`: 5 findings, all present at `5d58668` (diffed, no new ones):
  keymap_test.go SA1019 ×2, launch_check_test.go SA4006 ×3 (pre-existing `m = launchOf(m)`).
- Per commit `make test-fast PKG=...` ok: AC-01 config + gui/...; AC-02 brew/... + gui/...;
  AC-03 brew/... + cmd/... + app/....

## AC-01: default true

- `config.Default()` → `UpdateAlways`. Sample: `update_on_start: true`, then
  `# update_on_start: auto` and `# update_on_start: false` with their meanings.
  The bool is written unquoted (`sampleValue` maps `UpdateMode` to true/false).
- Tests: `TestProductionDefaultUpdatesOnEveryLaunch` PASS (one `brew update`,
  Outdated known after it). `TestOfflineLaunchShowsUpdateFailureAndLoadsOldMetadata`
  PASS (error toast `Update: test error`, Status loaded, Outdated from old metadata,
  key `4` still switches to PanelOutdated). Sample tests added and passing.
- Shared helpers set `UpdateNever` in one place each: `testConfig()` (gui/main_test.go)
  and `testutil.NewModel` (flows). Launch tests that need auto set it: `autoConfig()`
  in launch_check_test.go (8 call sites); flows `TestLaunchCheckStaleFlow` and
  `TestLaunchCheckFreshFlow` use `testutil.NewTestModelWithUpdate(..., UpdateAuto)`.
  Without these, 7 gui and 1 flows test failed under the new helper default.
- Goldens changed: none. Every gui and flows frame is unchanged.
- README (config block and paragraph) and DESIGN.md row updated.

## AC-02: LB-20 two edges

(a) Failing first:
  `--cask failed, --formula is empty: Search err = <nil>, want the failed read's error (results [])`
  → `TestSearchErrorsWhenNoRowsAndAReadFailed` FAIL; PASS after the fix.
  gui `TestSearchFailureShowsErrorToast` before: `a failed search shows no error toast; toasts = []`;
  PASS after. Error toast text is `Search: <err>`. Search errors only when there are no rows.
  Rows of the read that worked are still kept (`TestSearchKeepsTheReadThatWorked` PASS).
  Behaviour note: a search where both reads fail now also toasts (same text).
(b) Failing first: `TestSlowRemarkCannotRestoreAStaleMark` FAIL
  (`a slow re-mark from before the second uninstall restored the installed mark`).
  Fix: `remarkGen` on the model, bumped per re-mark; `SearchMarksMsg.gen` must match.
  With the guard disabled (`if false && ...`) the test fails again; restored, PASS.

## AC-03: leftovers

1. Coalescing tests with 20 ms sleeps now use `synctest.Wait`:
   `TestCasksOutdatedCoalescesConcurrentCalls`, `TestFormulaeOutdatedCoalescesConcurrentCalls`,
   `TestBrewReadsConcurrentCallsRunOnce`.
   `CGO_ENABLED=1 go test ./internal/brew/ -run 'Coalesces|ConcurrentCallsRunOnce|Singleflight' -race -count=100`
   → `ok  github.com/lthiagol/lazybrew/internal/brew  8.492s`.
2. Log file: `ResetLogger` closes the file `EnableFileLogging` opened; a second
   `EnableFileLogging` closes the one it replaces; `cmd/lazybrew/main.go` calls
   `brew.ResetLogger()` after `p.Run()`. Tests `TestResetLoggerClosesTheLogFile` and
   `TestEnableFileLoggingClosesThePreviousFile`: before, `undefined: logFile` (build
   failed); after, both PASS. Signatures unchanged.
3. `ResetLogger` doc now says init and main call it.

## Seen, not done

- brew `cached_test.go` invalidation tests (~:208, ~:239) still sleep 20 ms; they
  are not coalescing tests. `formulae_test.go` TTL tests sleep 75 ms.

## Questions filed

- None. Decisions taken: the error-toast wording `Search: <err>`; the log file is
  closed by `ResetLogger` called from main.

## Fix r2

- F-r1-01 (`e5b6b16`): `TestSingleflightCoalescesConcurrent`, `TestSingleflightKeysAreIndependent`, `TestSingleflightPanicUnblocksWaiters` use `synctest.Test` + `synctest.Wait`; no sleeps. `CGO_ENABLED=1 go test ./internal/brew -run Singleflight -race -count=100` → `ok  github.com/lthiagol/lazybrew/internal/brew  1.169s`.
- Mutations (restored after each, tests re-run green): sharing off (`inflight[key]` lookup `&& false`) → CoalescesConcurrent FAIL `expected exactly 1 fn invocation, got 16`, PanicUnblocksWaiters FAIL `waiter must not invoke fn while sharing panicked call`, PropagatesError FAIL. Key collapse (`inflight[""]`) → KeysAreIndependent FAIL `got a=0 b=1`. Panic cleanup removed (`call.wg.Done()` dropped) → PanicUnblocksWaiters FAIL `waiter blocked after leader panic`, then synctest deadlock panic.
- F-r1-02 (`276f126`): `testConfig` doc names `TestProductionDefaultUpdatesOnEveryLaunch`.
- F-r1-03: no action.
- Gate: `make test-fast PKG="./internal/brew/... ./internal/gui/..."` exit 0. `make ci` exit 0; floors met (app 70.0, brew 88.8, config 93.1, gui 90.8, anim 88.2, modal 89.2, presentation 100.0, style 89.5, task 96.2, logtail 92.3).
