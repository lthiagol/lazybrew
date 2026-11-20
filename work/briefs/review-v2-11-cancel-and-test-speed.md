# Review brief — slice `v2-11-cancel-and-test-speed`

> The slice contract is `work/briefs/v2-11-cancel-and-test-speed.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-11-cancel-and-test-speed.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands (see the brief's Safety section).

## Verify

- **V1 gate:** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`; time the race run and compare with result.md. Also `GOOS=darwin go build ./...`.
- **V2 cancel (AC-01):** read both runner paths; process group set, the whole
  group killed on cancel, `WaitDelay` set; a grandchild is really gone after
  cancel (run the new test, and try it by hand with a fake brew script —
  never real brew writes); output/errors/logging unchanged.
- **V3 tests (AC-02):** every changed test still proves what it proved (read
  old vs new side by side); no assertion dropped; production durations
  unchanged.
- **V3b CI (AC-02b):** both workflows run `-race` with CGO on; no other workflow change.
- **V4 invariants:** no GUI change; tests not weakened; one commit per AC.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-v2-11-cancel-and-test-speed-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports v2-11-cancel-and-test-speed rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-11-cancel-and-test-speed-rN.md`. Do not touch other panes.
