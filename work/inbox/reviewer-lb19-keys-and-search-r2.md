verdict: ok

# Review r2 — lb19-keys-and-search

Scope: b1a5c97..067bc35; the r2 fix commits are 85290cd..42696cb (one per finding, in order, staged by name;
each commit is green on `go test ./internal/brew/ ./internal/gui/` + vet). Code changes are only in
`brew/search.go` and `gui/tasks_packages.go`; the rest are tests. No rendering code changed, so the r1
V6 frame diff still holds.

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 | met | F-r1-01 closed (below); the unit tests are still red on revert (r1). |
| AC-02 | met | F-r1-04 and F-r1-05 closed; Q-lb19-search-partial is implemented (`Search` runs both reads and returns an error only when both fail) and pinned in both directions. |
| AC-03 | met | F-r1-06 closed; the r1 mutation of the main content is unchanged. |
| AC-04 | met | F-r1-02 and F-r1-03 closed; args and the `task.Manager` route are unchanged since r1. |

## r1 findings — I reverted each fix in a temp worktree and the new test fails

| Finding | Mutation | Result |
|---|---|---|
| F-r1-01 | drop `tab`+`shift+tab` under focused search (keys.go) | `TestSearchTabCyclesAllPanelsFlow`: `after 2 keys: main title is not Status` |
| F-r1-01 | drop only `shift+tab` | `after 9 keys: main title is not Services` (the backward walk is pinned too) |
| F-r1-02 | delete `m.remarkSearch()` from the TaskCompletedMsg batch (gui.go:452) | `the row still says installed after the uninstall` + the item still shows `installed` |
| F-r1-03 | drop `&& m.searchResults[i].Name == r.Name` | `row "wget" became installed from marks for another search` |
| F-r1-04 | tasks_packages.go from before 4c78f98 | `formula list unreadable: jq Installed = false, want true`; same for the cask case |
| F-r1-05 | search.go from before ee85b8f | `--cask failed: Search err = test error, want the other read's rows` |
| F-r1-06 | add a `pgdown` case to `helpKey` | `PgUp/PgDn scrolled the help screen: offset 2->7` |

The F-r1-01 flow checks `View()` (the main box's title) after each of the 15 keys, through all 7 panels
in both directions. `-race -count=10` passed with no flakes. The doc comments for `MarkInstalled` and
`markInstalled` now match what the code does.

Gate: `make ci` exit 0 (brew 87.8% ≥ 78, gui 90.6% ≥ 83, all floors met); `GOOS=darwin go build ./...` ok.

## Findings

None.

## Notes (not findings)
- Gate time: under `-race`, `TestSearchTabCyclesAllPanelsFlow` takes about 21 s (it runs one teatest
  program per key prefix, 15 in all), and the flows package goes from about 12 s to about 35 s. Without
  race the test takes about 1 s. Worth a backlog item if gate speed matters (`Q-v2-builder-speed`).
- Q-lb19-search-partial edge case: if one read exits 1 (no results) and the other fails, the user gets an
  empty list with no error. This follows the ruling as written, but it reads as "no results". The
  orchestrator can decide whether that needs its own ruling.
- If two writes complete back to back, their re-marks could land out of order, and a slow first re-mark
  could restore a stale mark. This predates r2, the timing makes it unlikely, and the name guard does
  not cover it.
- Carried from r1: Search → info has no `--cask` for a name that is both a formula and a cask.
