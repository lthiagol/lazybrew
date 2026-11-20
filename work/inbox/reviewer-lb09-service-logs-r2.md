verdict: ok

# Review lb09-service-logs — round r2

Scope: everything after `65662ee`; new since r1: fix brief `02af339`,
F-r1-01 `b9a586d`, F-r1-02 `608f031`, receipt `db79fbe`/`7ec0d40`/`96d8b88`.

## Gate (run by the reviewer, tree `96d8b88`)

- `make ci`: exit 0. gofmt, vet, `go test ./... -race -count=1` ok, floors met
  (gui 85.6% >= 36%), build ok. `GOOS=darwin go build ./...`: ok.

## Earlier findings

| Id | Status | Evidence |
|---|---|---|
| F-r1-01 (major) | fixed | `gui.go:278-282` both `followLogs` branches now `c := nm.syncLogs(); return nm, tea.Batch(cmd, c)`. |
| F-r1-02 (minor) | fixed | `services_logs_test.go:351` `TestLogsSwitchServiceShowsNewLogAndDropsOld`: moshi-hook and backup each have their own stdout file; after `j` the view has `beta line 1` and no `alpha line`; an old-gen `logsReadMsg` carrying alpha's lines leaves `plainView` byte-identical. Without the gen check in the `logsReadMsg` case, `logsShown` holds and `applyLogs` would draw alpha's lines, so the test would fail. |

## The 60 rewritten returns (b9a586d)

- **Mechanical proof.** A script (scratchpad `drv/undo.py`) takes each
  `c := EXPR` that sits directly before `return m|nm, …c…` with `c` used once,
  puts `EXPR` back in place of `c`, and compares the result with `b9a586d~1`
  after removing whitespace. Result: commands.go 40, gui.go 17, mouse.go 3 sites;
  all three files are **identical** to the pre-fix tree. So the returned values
  are the same, and no call was dropped or added.
- **No call crosses a branch.** Each `c :=` is the statement just before its
  `return`, in the same block (the script only matches when the two are next to
  each other). So the call runs on exactly the paths it ran on before.
- **Call order is the same.** In the 5 multi-argument `tea.Batch` cases
  (gui.go 509, 585, 592, 845; commands.go batchUpgrade queue-full), the hoisted
  `setToast` was already the **first** argument, and `RunNext()`,
  `serviceReload`, etc. still run after it, as before. In the others the whole
  second operand was hoisted.
- **Nothing missed.** None of the 50 pointer-receiver `Model` methods is now
  called inside a `return` that also returns the model, on one line or across
  several. The ones left as they were (`autoRefreshCmd`, `updateTickerCmd`,
  `fetchSelectedSearchInfo`) have value receivers. `tasks.RunNext` changes the
  task manager, not the model. The two pointer-receiver sites in commands.go
  (`handleModalResult`, `toggleTrust`) return `*Model`: the rewrite does no harm
  there.

## gui coverage 86.2 → 85.6 explained

Coverage profiles of `./internal/gui` at `7b4b236` (pre-fix) and HEAD:
statements 3339 → 3399 (+60, one per rewritten site: the extra `c :=`), covered
2880 → 2909 (+29), uncovered 459 → 490 (+31). Per file: commands.go 917 → 957
(+40), gui.go 647 → 664 (+17), mouse.go 112 → 115 (+3). So 31 of the split
sites are on paths no test reaches (mostly the queue-full and queued-behind toast
branches), and each one's uncovered statements doubled. The percentage fell
(2880/3339 = 86.25% → 2909/3399 = 85.58%), but covered statements went up by
29 and none went down. No test was lost, and the floor (36%) is unchanged.

## New problems

None found. The fix commits touch only the three files named and one test file;
no test was deleted or weakened, and the AC table from r1 still holds
(re-ran `TestLogs*`: ok).

## Findings

None.

Safety: no brew write command run; moshi-hook not touched.
