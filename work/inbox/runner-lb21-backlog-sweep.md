# Receipt: lb21-backlog-sweep (runner)

## Commits (one per AC, in order)

- AC-01 LB-13 `88f5f32` footers show tabs; Status shows navigation
- AC-02 LB-14 `90b1c5d` error test uses synctest, not a sleep
- AC-03 LB-15 `557e6a4` SetDebug replaced by ResetLogger
- AC-04 LB-18 `d0c4000` search-tab cycle test drives the model

## AC-01 (LB-13)

- Status, Outdated, Taps footers gain `h/l tabs`; Status also `j/k navigate`. footerExempt loses those three entries.
- 80x24: footers never wrap. `footerLine` cuts the tail and ends `? more`, as before. Newly cut: Status `B brewfile v vulns`, Outdated `a select all`, Taps `/ search`.
- Goldens changed (footer only): `TestFooterLeadsWithThePanelsKeys`, `TestTapsFooterLeadsWithTheTapActions`, `TestTapsFooterAndInfoAgreeOnActions` (skips `h/l`: tabs are not tap actions).
- View diff vs HEAD, 7 panels x 2 tabs at 80x24: 6 footer rows differ, nothing else.

## AC-02 (LB-14)

- `synctest.Test` + `synctest.Wait()`; the 20 ms sleep is gone.
- `-race -count=200`: 200 PASS.
- Sharing reverted: `singleflight_test.go:150: second caller must not invoke fn while first is in flight` → FAIL.

## AC-03 (LB-15)

- `SetDebug` removed; `ResetLogger()` added. Production enables debug only via `app.enableDebugLog` → `EnableFileLogging`.
- Tests: `TestEnableFileLoggingWritesToTheFileNotStderr`, `TestResetLoggerDiscardsDebug`. `app/debug_log_test.go` calls `ResetLogger`.
- Mutation (file logger writes to stderr): `logger_test.go:31: stderr should stay empty` → FAIL.
- `deadcode -test ./...`: no output, same as before. `staticcheck ./...`: same 5 findings as before.

## AC-04 (LB-18)

- One `gui.Model`, `View()` after each key. 21 s → 0.62 s under `-race`.
- lb19 Tab fix reverted: Tab branch → `navigation_test.go:122: after 2 keys: main title is not Status`; Shift+Tab branch → `after 9 keys: ... not Services`. Both FAIL.

## Gate

- `make ci`: exit 0 (fmt, vet, race tests, 10 coverage floors, build).
- `GOOS=darwin go build ./...`: ok.

## Notes

- Decision for orchestrator: tabs sit second, as the brief says. Putting them last in Status and Outdated keeps the actions visible at 80x24.
- Staticcheck findings, untouched: `keymap_test.go:39,431` SA1019; `launch_check_test.go:172,212,445` SA4006.
- Questions filed: none.
- Seen, not done: `singleflight_test.go` keeps 20 ms sleeps in three other tests; `EnableFileLogging` never closes its file.
- No brew write commands. TUI not launched.
