# Review brief — slice `lb04-lb10-small-fixes`

> The slice contract is `work/briefs/lb04-lb10-small-fixes.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-lb04-lb10-small-fixes.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands (see the brief's Safety section).

## Verify

- **V1 gate:** `make ci`; also `GOOS=darwin go build ./...`.
- **V2 chunking (AC-01):** read `fetchTapInfoBatch`; both bounds enforced
  (count and bytes), order kept, every name sent once, a typical list is one
  call with today's args, a failed chunk fails the batch, a cancelled ctx stops
  further chunks; `List`'s `TrustUnknown`/no-cache fallback unchanged; the
  tests prove each of these (break the chunking by hand — e.g. drop the last
  chunk — and check a test fails).
- **V3 cut row (AC-02):** the `…` row lines up with the list column in the
  saved frames and in the test; an uncut dialog is byte-identical to before
  (compare against the pre-slice commit); box height ≤ `dialogMaxRows`, width
  exact.
- **V4 invariants:** no other behaviour/frame change; tests not weakened; one
  commit per AC.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-lb04-lb10-small-fixes-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports lb04-lb10-small-fixes rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-lb04-lb10-small-fixes-rN.md`. Do not touch other panes.
