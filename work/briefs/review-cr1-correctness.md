# Review brief — slice `cr1-correctness`

> The slice contract is `work/briefs/cr1-correctness.md` (with
> `work/briefs/cr-common.md`); it wins where they differ. You review; you do
> **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-cr1-correctness.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands — do not start/stop/restart moshi-hook.

## Verify

- **V1 gate:** `make ci`; also `GOOS=darwin go build ./...`.
- **V2 per AC (AC-01 … AC-06):** each finding in the brief is really fixed in
  the code, not just in a test. For every regression test, revert the fix
  (`git stash`-free: check out the pre-fix file into a temp worktree) and
  confirm the test **fails** there; a test that passes without the fix is a
  finding.
- **V3 AC-01:** grep shows `extractPackageName`/`selectedItemName` have no
  production caller; call sites listed in the receipt match the grep.
- **V4 AC-05:** exactly one auto-refresh chain and one update ticker; the
  pause-on-keypress behaviour unchanged.
- **V5 invariants:** `p.items`, `presentation.Format*`, `extractPackageName`
  still exist; footer/help/goldens unchanged; no scope creep (cr3/cr4/cr5
  items untouched); Bubble Tea rules held; one commit per AC.

## Severity

blocker (gate fails, test deleted/weakened, regression, a Bubble Tea rule
broken) · major (AC not met, test passes without the fix, scope breach) ·
minor (docs/process, polish).

## Report

`work/inbox/reviewer-cr1-correctness-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports cr1-correctness rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-cr1-correctness-rN.md`. Do not touch other panes.
