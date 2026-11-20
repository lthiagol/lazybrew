# Review brief — slice `v2-05-content-polish`

> The slice contract is `work/briefs/v2-05-content-polish.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-05-content-polish.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands.

## Verify

- **V1 gate:** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`.
- **V2 frames:** run the committed harness against real Homebrew, diff vs
  `after/` (must match), every frame exactly W×H cells. Remove
  `internal/gui/dump` after.
- **V3 detail fields (AC-01):** one helper used by every listed view; empty
  values omitted; URLs Info role; prose wraps with hanging indent, other
  values cut with `…`; presentation snapshot changes named; floor kept.
- **V4 deps (AC-02):** installed/not-installed icons from loaded data;
  Runtime/Build grouping; cask `depends_on` object parsed against a fixture
  shaped like real `brew info --json=v2 --cask --installed` (run it — read
  only); Files styling.
- **V5 dashboard (AC-03):** cards; outdated gauge never shows `0` when
  unknown; doctor blocks; side-by-side vs stacked; no overflow at 80×24.
- **V6 help (AC-04):** every help line reachable at 80×24 by scrolling (test
  it); position label; keys only active inside help; two-column alignment.
- **V7 operation (AC-05):** step parsing from `==>` lines with a fixture (no
  brew write was run to make it — check how it was made); elapsed uses the
  existing tick (no goroutine); final states; batch bar `n/total`; Log
  unchanged.
- **V8 every frame change explained (`result.md`);** restyle of other slices'
  parts or animation = **major**.
- **V9 invariants:** no new brew commands, `internal/brew/**` only AC-02.2,
  only `bubbles/v2/progress` added, test count, removed assertions replaced,
  one commit per AC.

Also `cat after/140x40/*.ansi` in your terminal and flag anything that reads
badly as a **minor** finding with the frame name.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-v2-05-content-polish-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports v2-05-content-polish rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-05-content-polish-rN.md`. Do not touch other panes.
