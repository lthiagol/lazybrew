# v2-11-cancel-and-test-speed — evidence

## Gate, final state (commit `02e9cac`, AC-02b)

| command | result |
|---|---|
| `make build` | pass (`go build -o bin/lazybrew ./cmd/lazybrew`) |
| `make lint` | pass (`go vet ./...`, no output) |
| `CGO_ENABLED=1 go test ./... -race -count=1 -json` | pass, 0 failing tests |
| `make cover-check` | `brew` 79.2% ≥ 62, `gui/presentation` 98.5% ≥ 91, `gui` 85.3% ≥ 36, `gui/modal` 90.1% ≥ 41. Floors unchanged. |
| `GOOS=darwin go build ./...`, `GOOS=darwin go vet ./internal/brew/ ./internal/gui/...` | pass |
| `GOOS=windows go build ./...`, `GOOS=windows go vet ./internal/brew/` | pass |

Race run time: **38.9 s before → 20.8 s after**, on this machine, same command.
Per package (summary, s): brew 34.15 → 1.62; gui 26.59 → 18.19; flows 6.52 → 6.81; the other packages unchanged within 0.02 s.

`TestExecuteStreamCancelReportsCanceled`: 30.00 s → 0.01 s.

## What changed, by criterion

**AC-01 (`d76b8de`).** Both runner paths start brew in its own process group (`internal/brew/procgroup_unix.go`, `//go:build unix`; `procgroup_other.go` is the no-op for other platforms). On cancel the whole group gets SIGKILL, and `cmd.WaitDelay` is 2 s.
- `ExecuteStream` now hands `cmd.Stdout` and `cmd.Stderr` to exec as writers instead of `StdoutPipe`/`StderrPipe`. With pipes, `Wait` cannot run until the reader has drained, so a stray holder of the pipe hung the stream and `WaitDelay` never applied. The line splitting is unchanged (`lineWriter`). Stderr is collected in a buffer and read after `Wait`.
- `finishErr`: `exec.ErrWaitDelay` means brew exited 0 and only a stray child still held its output past the delay. It is not reported as a failure.
- Tests (`procgroup_unix_test.go`): a grandchild holding the pipe is gone after cancel and the stream closes (`ExecuteStream` and `Execute`); a `setsid` child that escapes the group does not hang the stream; a successful brew with such a child reports no error. Bite checks: without the group kill the grandchild tests fail; without the `WaitDelay` path the stray test fails; without `finishErr` the after-exit test fails.

**AC-02 (`f11d2a6`).** Seams are package variables or fields that tests override. Production values are unchanged.
- `toastTickInterval` (gui, 500 ms), `operationAutoClose` (gui, 2 s), `killGroupDelay` (brew, 2 s): set in gui `TestMain` (1 ms) or per test (brew, 100 ms).
- `modal.ToastTick` takes its interval as a parameter.
- `Cache.now`: the cache reads time through a field, so TTL tests move a clock forward instead of sleeping.
- `sfCall.waiters`: a count of callers queued behind an in-flight call, read under `sf.mu`. `TestSingleflightPropagatesError` releases `fn` once the second caller is queued. It used to poll a channel that closed only after `fn` had returned, so it always waited the full 2 s.

Changed tests, with what each still proves:
- `TestToastTickMsgEventuallyClearsToast`: no longer waits out the 3 s lifetime. A fresh toast survives its first tick and the chain re-arms (asserted as before). The toast is then aged past its lifetime and the next tick dismisses it (asserted as before).
- `TestSingleflightPropagatesError`: same assertions (fn is not invoked while the first call is in flight; both callers get the leader's error). The wait is now on the queued-caller count, not on a timer.
- `TestCacheTTL`, `TestCacheSetWithTTLUsesPerEntryTTL`, `TestCacheSetWithTTLZeroInheritsDefault`, `TestCacheSetWithoutTTLStillUsesDefault`, `TestCacheSetWithTTLAllKeys`: the same expiry and no-expiry assertions on a fake clock.
- `TestExecuteStreamCancelReportsCanceled`: waits for a marker file written by the fake brew, not 100 ms. Same assertion.
- `TestExecuteStreamCancelWinsOverReaderError`: reads three lines before cancelling, not 80 ms of output. The script's chatter interval went from 20 ms to 5 ms. Same assertion, 5 iterations.
- `TestFetchStatusDataTimesOut`, `TestFetchStatusDataRespectsTotalBudget`, `TestFetchStatusDataTotalBudgetBoundsTheAggregate`: budgets shortened (read 50→10 ms, total 500→100 ms; read 100→20 ms, total 200→40 ms; total 150→30 ms). The expected timeout text changed `50ms` → `10ms` in three assertions. The upper and lower bounds on elapsed time are unchanged.
- `TestInstallFlow`, `TestUninstallFlow` (flows): the fixed sleeps are replaced by waits on rendered text (`▸ ripgrep`, `Install ripgrep?`, `▸ ✓ testformula`, `Uninstall testformula?`) and on the recorded brew call. The key sequence is unchanged. `flows/wait_test.go` holds the helpers.
- `TestServiceFailureReloadsWithoutRefreshToast`: runs the batch once and checks both predicates on the same messages. Running a `tea.Tick` twice blocks on the second run, which the old helper turned into a 200 ms timeout. Same assertions.

**AC-02b (`02e9cac`).** `wip-ci.yml` and `stable-ci.yml`: the `go test` step is `go test ./... -race -count=1` with `CGO_ENABLED: "1"`. Nothing else changed. YAML validated with the repo's `yaml.v3`. **Not pushed.** The `wip-ci` run after the push must be checked by the orchestrator.

## Waits still in the suite (≥ 100 ms by test, after this slice)

- `gui/TestNewBatchStartsItsBarEmpty`, `TestProgressBarSpringsToTheBatchShare`, `TestProgressBarCountIsExactEveryFrame` (about 0.37 s each): these run bubbles' `progress` frame ticks through `pumpBar`. A `progress.FrameMsg` cannot be built outside bubbles, so the frame tick has to run. I tried synthetic frame times in `pumpBar` and it changed the spring result; I reverted it. Not reduced.
- `brew/TestExecuteStreamStrayHolderDoesNotHang`, `TestExecuteStrayHolderAfterExitIsNotAFailure` (0.1 s each): the 100 ms `WaitDelay` they test. This is the time under test.
- Most other entries above 100 ms are render cost under `-race` (for example `TestEveryFrameIsTerminalSize`, 3.6 s; `TestModalFloatsOverFrame`, 2.0 s), not sleeps. Without `-race` the same tests are under 100 ms. The flows (`TestInstallFlow` 1.7 s, `TestTabsFlow`, `TestHelpFlow`, …) are teatest programs under `-race`; `TestInstallFlow` now has no fixed sleeps and takes 0.11 s without `-race`. `flows` cannot reach the gui-package seams because it is an external test package.

## Judgement calls for the reviewer

- **Brief inconsistency, Report protocol.** The brief names the receipt `work/inbox/runner-v2-10-tap-actions.md` and the reply `done: work/inbox/runner-v2-10-tap-actions.md`. That is the v2-10 slice's name, so it looks copied from there. This lane wrote its receipt to `work/inbox/runner-v2-11-cancel-and-test-speed.md`, following the `<lane>-<slice>.md` rule in AGENTS.md and `work/README.md`. The reply line names that file. See `work/questions.md`.
- **STATE row.** The brief gives the row with `—` in the commits column. The commit ids are in the receipt, and the row is written as the brief specifies.
- **Attribution.** Commit `d76b8de` (AC-01) has no `Co-Authored-By` trailer. I had not yet read the attribution rule when I made it. The brief forbids amending, so it stays. Commits `f11d2a6`, `02e9cac` and the evidence commit carry the trailer.
- **Nothing pushed.** The brief's AC-02b says "after the push of this slice" and the orchestrator checks `wip-ci`. AGENTS.md forbids pushing without being asked.


# v2-11-cancel-and-test-speed — evidence

Machine: linux/arm64, go1.27.4 toolchain, `CGO_ENABLED=1`. Commands:

- before: `CGO_ENABLED=1 go test ./... -race -count=1 -json` on `c094035` (the slice's base), wall time 38.9 s.
- after: the same command on `02e9cac` (AC-02b), wall time 20.8 s, 0 failing tests.

Per-package wall time (go test summary, s): brew 34.15 → 1.62, gui 26.59 → 18.19, flows 6.52 → 6.81, the rest unchanged within 0.02 s.

## Slowest 10 tests

Before:

| test | s |
|---|---|
| brew/TestExecuteStreamCancelReportsCanceled | 30.00 |
| gui/TestEveryFrameIsTerminalSize | 3.40 |
| gui/TestToastTickMsgEventuallyClearsToast | 3.31 |
| gui/TestNonRRefreshEmittersStayCacheOnly | 2.01 |
| gui/TestModalFloatsOverFrame | 2.00 |
| brew/TestSingleflightPropagatesError | 2.00 |
| flows/TestInstallFlow | 1.75 |
| flows/TestUninstallFlow | 1.26 |
| gui/TestToastStackLeavesCellsBesideBoxesAlone | 1.18 |
| gui/TestMainAreaHasInnerPadding | 1.03 |

After:

| test | s |
|---|---|
| gui/TestEveryFrameIsTerminalSize | 3.56 |
| gui/TestModalFloatsOverFrame | 1.99 |
| flows/TestInstallFlow | 1.74 |
| gui/TestToastStackLeavesCellsBesideBoxesAlone | 1.17 |
| gui/TestMainAreaHasInnerPadding | 1.06 |
| flows/TestUninstallFlow | 1.06 |
| gui/TestOverlayFramesAreTerminalSize | 0.87 |
| gui/TestToastStackLeavesCellsBesideBoxesAlone/140x40 | 0.87 |
| flows/TestRefreshFlow | 0.60 |
| gui/TestMovingAwayClosesAFinishedOperation | 0.58 |

## Tests at or above 100 ms

Every test at or above 100 ms in the before run, with its after time. Two tests are new (AC-01 stray-holder tests) and are listed last.

| test | before | after |
|---|---|---|
| brew/TestExecuteStreamCancelReportsCanceled | 30.00 s | 0.01 s |
| gui/TestEveryFrameIsTerminalSize | 3.40 s | 3.56 s |
| gui/TestToastTickMsgEventuallyClearsToast | 3.31 s | 0.00 s |
| gui/TestNonRRefreshEmittersStayCacheOnly | 2.01 s | 0.01 s |
| gui/TestModalFloatsOverFrame | 2.00 s | 1.99 s |
| brew/TestSingleflightPropagatesError | 2.00 s | 0.00 s |
| flows/TestInstallFlow | 1.75 s | 1.74 s |
| flows/TestUninstallFlow | 1.26 s | 1.06 s |
| gui/TestToastStackLeavesCellsBesideBoxesAlone | 1.18 s | 1.17 s |
| gui/TestMainAreaHasInnerPadding | 1.03 s | 1.06 s |
| gui/TestToastStackLeavesCellsBesideBoxesAlone/140x40 | 0.88 s | 0.87 s |
| gui/TestOverlayFramesAreTerminalSize | 0.87 s | 0.87 s |
| gui/TestMovingAwayClosesAFinishedOperation | 0.58 s | 0.58 s |
| gui/TestExistingToastsSpringWhenOneArrives | 0.52 s | 0.53 s |
| gui/TestFailedBatchTaskDecrementsCounter | 0.51 s | 0.00 s |
| gui/TestNonRRefreshEmittersStayCacheOnly/install_completion | 0.50 s | 0.00 s |
| gui/TestNonRRefreshEmittersStayCacheOnly/uninstall_completion | 0.50 s | 0.00 s |
| gui/TestNonRRefreshEmittersStayCacheOnly/install_mutation_result | 0.50 s | 0.00 s |
| gui/TestNonRRefreshEmittersStayCacheOnly/uninstall_mutation_result | 0.50 s | 0.00 s |
| gui/TestBatchUpgradeQueueFullStillStartsBatch | 0.50 s | 0.00 s |
| gui/TestBatchUpgradeQueueFullAtFirstEnqueueSurfacesToast | 0.50 s | 0.00 s |
| gui/TestBatchUpgradeZeroValidSelectionSurfacesToast | 0.50 s | 0.00 s |
| gui/TestRefreshDuringBatchConfirmStillSurfacesToast | 0.50 s | 0.00 s |
| gui/TestSlidingFramesAreTerminalSize | 0.49 s | 0.49 s |
| flows/TestTabsFlow | 0.49 s | 0.54 s |
| gui/TestModalSlidesDownIntoItsPlace | 0.45 s | 0.44 s |
| gui/TestToastSlidesInFromTheRight | 0.44 s | 0.44 s |
| flows/TestNavigationFlows | 0.44 s | 0.50 s |
| brew/TestExecuteStreamCancelWinsOverReaderError | 0.42 s | 0.07 s |
| gui/TestDimBackdropMutesEveryRowButTheFooter | 0.41 s | 0.43 s |
| flows/TestModalFlow | 0.41 s | 0.48 s |
| gui/TestServiceFailureReloadsWithoutRefreshToast | 0.40 s | 0.00 s |
| gui/TestToastFloatsAboveBottomBar | 0.40 s | 0.40 s |
| flows/TestRefreshFlow | 0.40 s | 0.60 s |
| gui/TestModalFloatsOverFrame/input/200x60 | 0.39 s | 0.38 s |
| flows/TestSearchFlow | 0.39 s | 0.43 s |
| gui/TestModalFloatsOverFrame/menu/200x60 | 0.38 s | 0.37 s |
| gui/TestNewBatchStartsItsBarEmpty | 0.38 s | 0.39 s |
| gui/TestModalFloatsOverFrame/confirm/200x60 | 0.37 s | 0.37 s |
| gui/TestProgressBarSpringsToTheBatchShare | 0.37 s | 0.37 s |
| gui/TestProgressBarCountIsExactEveryFrame | 0.37 s | 0.37 s |
| flows/TestHelpFlow | 0.37 s | 0.46 s |
| gui/TestToastStackLeavesCellsBesideBoxesAlone/80x24 | 0.31 s | 0.30 s |
| gui/TestModalsShowContentOnFirstFrame | 0.30 s | 0.31 s |
| gui/TestBackdropIsDimAtOnce | 0.30 s | 0.30 s |
| gui/TestModalBoxKeepsItsOwnStyling | 0.26 s | 0.26 s |
| gui/TestDimBackdropMutesEveryRowButTheFooter/200x60 | 0.23 s | 0.24 s |
| gui/TestToastFloatsAboveBottomBar/200x60 | 0.23 s | 0.23 s |
| gui/TestSuccessfulOperationClosesOnItsTimer | 0.22 s | 0.21 s |
| gui/TestServicesInfoDrawsTheReloadOnceTheRunCloses | 0.22 s | 0.22 s |
| gui/TestFetchStatusDataRespectsTotalBudget | 0.20 s | 0.04 s |
| gui/TestModalFloatsOverFrame/input/140x40 | 0.20 s | 0.21 s |
| gui/TestModalFloatsOverFrame/menu/140x40 | 0.20 s | 0.19 s |
| gui/TestModalFloatsOverFrame/confirm/140x40 | 0.19 s | 0.21 s |
| gui/TestModalBoxKeepsItsOwnStyling/200x60 | 0.15 s | 0.15 s |
| gui/TestFetchStatusDataTotalBudgetBoundsTheAggregate | 0.15 s | 0.03 s |
| gui/TestEveryFrameIsTerminalSize/confirm_modal/200x60 | 0.15 s | 0.16 s |
| gui/TestMovingAwayClosesAFinishedOperation/shift+tab | 0.15 s | 0.15 s |
| gui/TestEveryFrameIsTerminalSize/input_modal/200x60 | 0.14 s | 0.15 s |
| gui/TestEveryFrameIsTerminalSize/menu_modal/200x60 | 0.14 s | 0.15 s |
| gui/TestEveryFrameIsTerminalSize/toast/200x60 | 0.14 s | 0.15 s |
| gui/TestMovingAwayClosesAFinishedOperation/1 | 0.14 s | 0.15 s |
| gui/TestMovingAwayClosesAFinishedOperation/click | 0.14 s | 0.14 s |
| gui/TestMovingAwayClosesAFinishedOperation/tab | 0.14 s | 0.14 s |
| gui/TestLogBoxHasInnerPadding | 0.13 s | 0.13 s |
| gui/TestDimBackdropMutesEveryRowButTheFooter/140x40 | 0.12 s | 0.14 s |
| gui/TestScrollKeysMoveAndClamp | 0.12 s | 0.10 s |
| gui/TestToastFloatsAboveBottomBar/140x40 | 0.12 s | 0.12 s |
| gui/TestEveryFrameIsTerminalSize/services_info_running/200x60 | 0.11 s | 0.09 s |
| gui/TestPreloadKeysWorkMidPreload | 0.11 s | 0.09 s |
| gui/TestEarlySwitchCountReachesTheDashboard | 0.11 s | 0.11 s |
| brew/TestCacheTTL | 0.10 s | 0.00 s |
| gui/TestFetchStatusDataTimesOut | 0.10 s | 0.02 s |
| gui/TestEveryFrameIsTerminalSize/services_info_scheduled/200x60 | 0.10 s | 0.10 s |
| gui/TestFailedReadForgetsItsCount | 0.10 s | 0.09 s |
| brew/TestExecuteStrayHolderAfterExitIsNotAFailure (new) | — | 0.10 s |
| brew/TestExecuteStreamStrayHolderDoesNotHang (new) | — | 0.11 s |
