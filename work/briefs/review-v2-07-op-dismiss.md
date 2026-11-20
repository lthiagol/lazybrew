# Review brief — slice `v2-07-op-dismiss`

> The slice contract is `work/briefs/v2-07-op-dismiss.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-07-op-dismiss.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands (see the brief's Safety section).

## Verify

- **V1 gate:** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`.
- **V2 frames:** run the committed harness against real Homebrew; `diff -r`
  vs `v2-06-animations/after/` must be empty; every frame exactly W×H.
  Remove `internal/gui/dump` after.
- **V3 auto-close (AC-01):** read every path that sets/clears `m.opState`.
  The dismiss message cannot clear a newer, running or batch-in-progress
  operation; no sleep in tests; no goroutine/`program.Send`.
- **V4 hints (AC-02):** rows and footer as specified; `op/` frames exact size.
- **V5 moving away (AC-03):** done → cleared on every panel-switch path
  (keys, Tab, mouse); running → not cleared; Services Info shows the reloaded
  state after.
- **V6 invariants:** no restyle/key change; tests not weakened; one commit
  per AC.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-v2-07-op-dismiss-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports v2-07-op-dismiss rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-07-op-dismiss-rN.md`. Do not touch other panes.
