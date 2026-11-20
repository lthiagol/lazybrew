# Review brief — slice `v2-10-tap-actions`

> The slice contract is `work/briefs/v2-10-tap-actions.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-10-tap-actions.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands (see the brief's Safety section).

## Verify

- **V1 gate:** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`.
- **V2 frames:** run the committed harness against real Homebrew; every
  frame exactly W×H; new screens present; the confirm screens were closed
  with `esc` (read the harness — a `y`/`enter` on a tap confirm is a
  **blocker**). Remove `internal/gui/dump` after.
- **V2b trust unknown (AC-00):** a timed-out or failed tap-info gives a flagged, uncached list; the panel says `trust unknown — press R`; `t`/`T` refused; other keys work.
- **V3 state-aware actions (AC-01, AC-02):** each tap state (official,
  trusted, untrusted) shows the right action line and footer; refused
  actions toast and never enqueue (tests with the mock runner).
- **V4 confirms (AC-03):** untrust and untap confirms with the stated kind,
  chip and consequence; untap names installed packages from that tap.
- **V5 after the action (AC-04):** list and info reload, selection kept by
  name; errors readable.
- **V6 invariants:** every write through `task.Manager`; no brew write run on
  the machine; tests not weakened; one commit per AC.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-v2-10-tap-actions-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports v2-10-tap-actions rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-10-tap-actions-rN.md`. Do not touch other panes.
