verdict: changes-required

# Review r1 — lb22-update-default-and-leftovers (fc82630..6cc8fb5, base 5d58668)

| AC | Verdict | Evidence |
|---|---|---|
| V1 gate | ok | `make ci` EXIT 0, all floors met (app 70.0, brew 88.8, config 93.1, …); `GOOS=darwin go build ./...` ok |
| AC-01 default true | ok | `Default()` = `UpdateAlways`; sample `update_on_start: true` + commented `auto`/`false`; `TestSampleLoadsToDefault`, `TestUpdateOnStartValues`, keyless file → true all PASS; README/DESIGN agree. Reverting the default → `TestProductionDefaultUpdatesOnEveryLaunch` and `TestOfflineLaunch…` FAIL. `testConfig()` and `testutil.NewModel` set `UpdateNever` once each. Own `View()` diff base vs HEAD at 80x24 and 140x40 with `config.Default()`: only header line 1 differs (`never updated` before Init's `StartUpdateMsg`); after it both sizes read `updating…`. `-race -count=20` on gui Launch/ProductionDefault/Offline/RKey and flows LaunchCheck: ok |
| AC-02 (a) | ok | revert `search.go` → `TestSearchErrorsWhenNoRowsAndAReadFailed` and `TestSearchFailureShowsErrorToast` FAIL; drop the toast → gui test FAIL |
| AC-02 (b) | ok | guard disabled → `TestSlowRemarkCannotRestoreAStaleMark` FAIL ("restored the installed mark"); `remarkSearch` mutates the addressable `m` returned from `update` (gui.go ~:463) |
| AC-03.1 synctest | **changes-required** | F-r1-01 |
| AC-03.2 log closed | ok | revert close → both logger tests FAIL (`err = <nil>, want os.ErrClosed`); `ResetLogger` swaps to the discard logger before `Close`, so later logs go to `io.Discard` |
| AC-03.3 doc | ok | `ResetLogger` doc names init and main |
| V5 | ok (one low) | one commit per AC; Makefile/floors untouched; no new goroutines in handlers (re-mark is a `tea.Cmd`); deadcode: none at base or HEAD; staticcheck: same 5 findings at base and HEAD |

## Findings

F-r1-01 (medium) — internal/brew/singleflight_test.go:42, :108, :203 — AC-03.1 is
not done: the three tests LB-21 names ("three other singleflight tests … same
synctest fix as LB-14", i.e. the siblings of `TestSingleflightPropagatesError` in
this file) still `time.Sleep(20 * time.Millisecond)`:
`TestSingleflightCoalescesConcurrent`, `TestSingleflightKeysAreIndependent`,
`TestSingleflightPanicUnblocksWaiters`. The file is untouched since 90b1c5d. The
runner converted three other tests instead (cached/casks/formulae coalescing; those
conversions are sound: sharing reverted → all three FAIL; `-race -count=100` ok).
The receipt's `-count=100` run matched `Singleflight` but proves nothing about
sleeps. — Fix must achieve: the three tests above wait with `synctest.Test` +
`synctest.Wait` (no sleep), pass `CGO_ENABLED=1 go test ./internal/brew -run
Singleflight -race -count=100`, and still fail when `singleflight.Do` stops sharing
(panic test: still unblocks waiters). Keep the cached/casks/formulae conversions.

F-r1-02 (low) — internal/gui/main_test.go:30 — the `testConfig` doc names
`TestProductionDefaultUpdatesOnLaunch`; the test is
`TestProductionDefaultUpdatesOnEveryLaunch`. — Fix must achieve: the doc names the
real test.

F-r1-03 (low) — commits fc82630, dcc09ce, 6c5ee76 — cr-common says the subject ends
with the finding ids in brackets; none do (`[LB-20]`, `[LB-21]`). No amend: the fix
commits carry their ids (`[LB-21]` for F-r1-01).
