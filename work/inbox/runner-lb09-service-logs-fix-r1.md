# lb09-service-logs fix r1 receipt
Source: work/inbox/reviewer-lb09-service-logs-r1.md. Brief: work/briefs/fix-lb09-service-logs-r1.md.
Commits: F-r1-01 b9a586d; F-r1-02 608f031.

## F-r1-01 — 60 sites (pre-fix tree 02af339, its line numbers)
Rule: `c := m.X(...)` first, then `return m|nm, ...c`.
- commands.go (40): 137,199,201,203,205,363,366,387,390,563,566,579,666,672,678,
  780,783,817,819,1028,1031,1077,1080,1109,1112,1159,1162,1209,1212,1344,
  1593,1596,1628,1631,1651,1654,1674,1677,1697,1700
- gui.go (17): followLogs 278,280; update 440,509,584,589,645,670,778,839,908,914,937,939,942,944,953
- mouse.go (3): 106,162,214
Re-scan, single and multi-line: none left. `m.tasks.RunNext()` and
`m.activeModal.Init()` do not mutate the model; untouched.

## F-r1-02
TestLogsSwitchServiceShowsNewLogAndDropsOld: after `j`, only beta's lines show; an
old-gen read of alpha's lines leaves the view identical. Removing the gen guard fails it.

## Gate
- `make test-fast PKG=./internal/gui/...`: exit 0 after each fix.
- `make ci`: exit 0 (gofmt, vet, -race, floors: brew 80.0, presentation 98.5, gui 85.6, modal 90.2; build).
- `GOOS=darwin go build ./...`: exit 0. gui coverage 86.2 -> 85.6%, floor unchanged.
Rewritten returns hand back the same values; only statement order changed.
