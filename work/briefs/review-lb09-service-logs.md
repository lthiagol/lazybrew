# Review brief — slice `lb09-service-logs`

> The slice contract is `work/briefs/lb09-service-logs.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-lb09-service-logs.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands (see the brief's Safety section) — do not
start/stop/restart moshi-hook.

## Verify

- **V1 gate:** `make ci`; also `GOOS=darwin go build ./...`.
- **V2 tail (AC-01):** bounded read from the end (never the whole file);
  every listed edge case tested; terminal-unsafe bytes neutralised.
- **V3 tab (AC-02):** run `./bin/lazybrew` (read-only: Services → `]` → Logs,
  `e`, `esc`; never `s/S/r/f/c`), compare with the saved frames; footer/help
  as specified; every empty state; frame exactly the terminal size.
- **V4 follow (AC-03):** read the tick/read loop: Bubble Tea rules held (no
  goroutines, no `program.Send`); stale ticks and reads dropped; one loop at
  most (switch tabs/services quickly in a test or by hand); pin-to-bottom vs
  scrolled-up behaviour; truncation; no work when unchanged. Try it by hand
  with a temp file you append to (a fake service in the mock runner, or the
  tests).
- **V5 invariants:** other panels/Info/actions unchanged; tests not
  weakened; one commit per AC.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression, a Bubble Tea rule broken) · major (AC not met, scope
breach, unexplained change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-lb09-service-logs-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports lb09-service-logs rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-lb09-service-logs-rN.md`. Do not touch other panes.
