# Review brief — slice `v2-08-preload`

> The slice contract is `work/briefs/v2-08-preload.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-08-preload.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands (see the brief's Safety section).

## Verify

- **V1 gate:** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`.
- **V2 frames:** run the committed harness against real Homebrew; every
  frame exactly W×H; check the changes vs `v2-07-op-dismiss/after/` are only
  the ones `result.md` lists (sidebar boxes filled / loading state).
  Remove `internal/gui/dump` after.
- **V3 scheduling (AC-01):** read the scheduler: Status first, the stated
  order, never more than 2 reads in flight (a test proves it), each through
  `boundedRead`; no goroutine/`program.Send`; a panel the user opens early is
  fetched once, not twice; `update_on_start` path still correct;
  `gui.preload: false` → today's lazy behaviour exactly.
- **V4 input (AC-02):** keys work while preloading (a test); a failed read
  shows its error in that box and does not stop the others.
- **V5 loading state (AC-03):** spinner/`loading…` while in flight, the
  count once loaded, `not loaded` only when preload is off.
- **V6 invariants:** tests and harness behaviour explicit about preload; no
  new brew command; tests not weakened; one commit per AC.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-v2-08-preload-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports v2-08-preload rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-08-preload-rN.md`. Do not touch other panes.
