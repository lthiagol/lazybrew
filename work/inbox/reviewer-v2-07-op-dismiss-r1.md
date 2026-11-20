verdict: ok

# reviewer — v2-07-op-dismiss r1

Scope: `wip` after `2a7e2a5` (brief commit): `0dfd9ff` `ede3959` `21b3b21`
`21fb1a1` `39c15c0` `473b247`, plus orchestrator `cc99ed0` (questions only).
No brew write command ran; `update_on_start` defaults false, so the harness Init
starts no `brew update`.

## Verify

- **V1 gate (HEAD):** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`
  → exit 0; all 11 packages ok; floors met (brew 76.5, presentation 98.2, gui 84.1, modal 90.1).
  Per commit: `0dfd9ff` and `ede3959` checked out in temp worktrees, `go build ./...` +
  `go test ./internal/gui/... -race -count=1` ok; `21b3b21` code = HEAD code.
- **V2 frames:** committed harness (`frame_dump_test.go.txt`, identical to v2-06's) +
  `op_dump_test.go.txt` run against real Homebrew into scratch, `internal/gui/dump`
  removed after. vs committed `v2-07-op-dismiss/{after,op}`: `diff -r` empty.
  vs `v2-06-animations/after/`: only 19-services-info and 20-services-stop-confirm
  (.txt/.ansi 140×40, .txt 80×24) differ, each only in `PID: 1332457` vs `PID: 800`
  and its padding — accepted by ruling Q-v2-07-pid. All 43 .txt frames exactly W×H
  (display width).
- **V3 auto-close:** `opState` is set only in `TaskStartedMsg`; cleared by brew-update
  completion (unchanged), Esc, `switchPanel`, and `opAutoCloseMsg`. The latter clears
  only when `opState == msg.op && Done() && batchCount == 0` (gui.go:352).
  `autoCloseCmd` (operation.go:33) issues `tea.Tick` only for `opSuccess` with
  `batchCount == 0`, after batch bookkeeping. No goroutine, no `program.Send`, no sleep in tests.
- **V4 hints:** rows `✓ Completed in 0s · Esc to close`, `✗ Failed: … · Esc to close`,
  `Cancelled · Esc to close` (Subtle hint); long error truncated so the hint survives;
  footer first hint `esc close` only when `Done()`. `op/*.txt` 140×40 exact.
- **V5 moving away:** every switch path (keys 1–7, `/`, Tab/S-Tab, sidebar click and
  wheel, search results) goes through `switchPanel`, which clears a done op only when
  the panel changes; running op untouched. RefreshMsg reloads the active Services panel;
  test asserts the reloaded Status after close.
- **V6 invariants:** only footer.go, gui.go, messages.go, operation.go + new test file;
  no existing test changed or deleted; no new module dependency (x/ansi already used);
  one commit per AC in order.

## AC table

| AC | Commit | Verdict | Evidence |
|---|---|---|---|
| AC-01 | 0dfd9ff | met | 2 s `operationAutoClose` const; pointer-named msg; stale/running/batch tests assert `View()` |
| AC-02 | ede3959 | met | `TestFinishedRunsEndWithTheCloseHint` exact rows; footer test asserts `View()` |
| AC-03 | 21b3b21 | met | Tab, S-Tab, `1`, click clear done; running kept; Info shows reloaded state |
| AC-04 | 21fb1a1 | met | harness copied unchanged; `op/` frames; result.md; STATE row as specified; PID diff per ruling |

## Findings

None.

## Notes (not findings)

- The runner receipt took two commits (`39c15c0` at 51 lines, `473b247` trimmed to 38)
  where the brief says "committed once"; the second commit names why, no amend was used,
  and the final receipt meets the limit. Nothing left to fix.
- A done op is also cleared when async search results switch to Search; consistent
  with AC-03 (it is a panel change).
