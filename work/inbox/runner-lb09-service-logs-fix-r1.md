# lb09-service-logs — fix round r1 receipt

Source: `work/inbox/reviewer-lb09-service-logs-r1.md` (changes-required).
Brief: `work/briefs/fix-lb09-service-logs-r1.md`.

## Commits (one per finding)

- F-r1-01 `b9a586d` fix(gui): no return mixes the model with a call that mutates it
- F-r1-02 `608f031` test(gui): a Logs switch to another service drops the old service's read

## F-r1-01 — every instance (60, pre-fix tree `02af339`, its line numbers)

Rule applied: the mutating `*Model` method runs first into a local `c`; the
return is `return m|nm, ...c...`. Found by scanning every return that names the
model and calls a pointer method on it, single-line and multi-line.

- commands.go: handleModalResult 137,199,201,203,205; executeUntap 363,366;
  executeRepair 387,390; doMutation 563,566; batchUpgrade 579,666,672,678;
  executeTapAdd 780,783; toggleTrust 817,819; runServiceTask 1028,1031;
  togglePin 1077,1080; executeServicesCleanup 1109,1112; runVulns 1159,1162;
  runMissing 1209,1212; confirmUntap 1344; runDoctor 1593,1596;
  brewCleanup 1628,1631; runAutoremove 1651,1654; executeCleanup 1674,1677;
  executeAutoremove 1697,1700
- gui.go: followLogs 278,280 (now the brief's `c := nm.syncLogs()` form);
  update 440,509,584,589,645,670,778,839,908,914,937,939,942,944,953
- mouse.go: handleMouseWheel 106; handleSidebarClick 162; handleTabBarClick 214

Re-scan after the change, single and multi-line: no instance left. Calls not on
the model (`m.tasks.RunNext()`, `m.activeModal.Init()`) were left as they were:
they do not mutate the model.

## F-r1-02 — the test

`TestLogsSwitchServiceShowsNewLogAndDropsOld` (services_logs_test.go): two
services, moshi-hook (alpha.log) and backup (beta.log). After `j`, beta's lines
show and no alpha line does. Then a read of the old gen carrying alpha's lines
lands: the view is byte-identical. Mutation check: removing the gen guard in the
`logsReadMsg` case fails this test; the guard was restored and the test passes.

## Gate

- Per commit: `make test-fast PKG=./internal/gui/...` exit 0 after each fix.
- `make ci` exit 0: gofmt, vet, `-race` tests all `ok`; "All coverage floors met"
  (brew 80.0 ≥ 62, presentation 98.5 ≥ 91, gui 85.6 ≥ 36, modal 90.2 ≥ 41); build ok.
- `GOOS=darwin go build ./...`: exit 0.

## Notes

- gui coverage is 85.6%, down from 86.2%. Floor unchanged.
- The rewritten returns hand back the same model and command values. Only the
  order of the statements changed.
- Nothing outside the fix scope was touched: no STATE row, no questions, no
  brief, no verification files.
