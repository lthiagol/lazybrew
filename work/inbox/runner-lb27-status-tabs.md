# runner lb27-status-tabs — receipt

## Commits (AC order)

- AC-01 Missing tab — `6ab7914`
- AC-02 Vulns tab, check on v only — `cdbcb03`; fix-up `3310921` (cdbcb03 did not compile its tests: helper named `openStatusTab`, called as `pressStatus`)
- AC-03 d/m/v open the tabs, overlay runners removed — `f535d23`
- AC-01..03 lack the `Co-Authored-By` trailer: not amended (no-amend rule). Later commits have it.

## Evidence

- **AC-01** (tests first). Red: `undefined: missingView`, `undefined: statusMissingTab`. Green: `TestMissingViewRows`, `TestMissingTabFetchesRows`, `TestMissingTabRealErrorShowsError`, `TestMissingTabLoadsOnceAndStays` (cached across `2`/`1`; R reloads once), `TestMissingTabFrames` (80x24, 140x40, dark, light).
- **AC-02**. Tests and code written in one pass; not run red first. Green: `TestVulnsTabOpenRunsNothing` (0 calls), `TestVulnsKeyRunsOnceWhileRunning` (second v: nil cmd), `TestVulnsResultShapes` (report, empty, exit 1 with report, error), `TestVulnsReportSurvivesRefreshAndPanelSwitch`, `TestVulnsTabFrames`. Timeout: `brewVulnsTimeout = 2 * time.Minute` (OSV.dev; local reads use 30 s).
- **AC-03** (tests first). Red: `d: active tab 0, want 2`; `m: active tab 0, want 3`; `two d presses ran brew doctor 1 times, want 2`; `TestStatusTabsFlow: WaitFor: condition not met`. Green: `TestStatusKeysOpenTheirTabs`, `TestStatusDoctorKeyRefreshes`, `flows.TestStatusTabsFlow`.

## Gate

- `GOFLAGS=-p=1 make ci` at f535d23: exit 0 (fmt, vet, `-race` all ok, floors met: gui 91.7 ≥ 83, build).
- `GOOS=darwin go build ./...`: ok. `deadcode ./cmd/...`: clean (flagged `linesTask`, `sendLines` after AC-03; removed).
- `staticcheck ./...`: only pre-existing hits in untouched files (`keymap_test.go` SA1019; `launch_check_test.go` SA4006).
- Per commit: `make test-fast PKG=./internal/gui/...` exit 0.

## Removed tests

- `task_build_test.go`: `TestLinesTaskSendsLinesInOrder`, `…EmptyResult`, `…Error`, `…CancelMidStream`, `…CancelledBeforeStart`, `…CancelThroughManager`; helpers `errWithin`, `hundredLines`, `completionWithin`. `linesTask` had no caller left.
- `gui_test.go`: `TestRunDoctor`, `TestRunVulns`, `TestRunMissing`.
- Goldens: none changed. `hot_path_test.go` needsTabFetch table gained `statusMissingTab` (AC-01).

## Decisions

- `d`/`m` drop the `diagnostics` cache group first, as R does (also drops Config).
- R clears Doctor/Missing content as before; the Vulns report survives R.
- README had no Status keys section; four lines added under Keybindings.

## Seen, not done

- `SMOKE-CHECKLIST.md:132-134` cites the removed `TestRunDoctor`/`TestRunMissing`/`TestRunVulns`.
- No copy test covers `y` on the Vulns tab.

## Next session

- Gate passed. No brew write command run; TUI not launched.

## Fix r2 (work/briefs/fix-lb27-status-tabs-r2.md)

- F-r1-01 `b548db4`: `brew.Vulns` fails on exit 1 with no stdout or an `Error:` stderr line; exit 1 with a report stays a result. Red before the fix: `TestVulnsResultShapes/exit_1_no_report`; also `TestDiagnosticsVulns*` (3).
- F-r1-02 `b450312`: Status bar drops the literal gap either side of each name (`tabGap`: 0 on Status, 1 elsewhere). At 80x24, both themes, all five names and the underline under Vulns show: `TestStatusTabRowShowsEveryName` (red before). Formulae gap held by `TestOtherTabBarsKeepTheirGap`. Status tabs are also tighter at 140.
- F-r1-03 `05f7aa5`, `1610d81`: `TestVulnsTimeoutShowsErrorAndHint`, `TestVulnsRunningFrameFitsAt80x24`, `TestYCopiesTheVulnsReportWhole`. They pass on arrival (pin behaviour the reviewer checked by hand). `1610d81` drops an unused assignment (SA4006).
- F-r1-04 `d99091e`: SMOKE-CHECKLIST cites `TestStatusKeysOpenTheirTabs`, `TestMissingTabFetchesRows`, `TestVulnsKeyRunsOnceWhileRunning`, `flows.TestStatusTabsFlow` and others, all existing.
- Gate at `1610d81`: `GOFLAGS=-p=1 make ci` exit 0; `GOOS=darwin go build ./...` ok; deadcode clean; staticcheck: no new hit.
- Not done: SMOKE-CHECKLIST `d`/`m` lines still say "output in modal / status path" (no overlay now). Wording only; left alone.
