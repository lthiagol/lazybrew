# Review brief — slice `v2-06-animations`

> The slice contract is `work/briefs/v2-06-animations.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-06-animations.md`),
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
- **V3 engine (AC-01):** harmonica springs; snap on settle; one 60 fps
  tick only while animating — prove no idle tick (a test, and read the code:
  every path that schedules the tick checks for motion); no goroutines; config
  default on, env override, off → instant; tests/harness run with animations
  off (check `config.Default()` callers in tests and the harness).
- **V4 resting frames unchanged (AC-06.1):** run the harness, `diff -r` vs
  `v2-05s-services/after/` must be empty. Any difference is a **blocker**.
- **V5 intermediate frames:** every frame in `motion/` and every step the
  tests take is exactly W×H; sidebar heights sum to the sidebar height at
  every step; the selected row stays visible during the accordion.
- **V6 input never waits:** keys and clicks act at once mid-animation (a test
  presses a key on the first frame of a modal slide / focus change); closing
  is instant.
- **V7 progress spring (AC-05):** `n/total` exact at every frame; off → static.
- **V8 invariants:** only harmonica promoted to direct; no restyle/content/key
  change; DESIGN/README updated; test count; one commit per AC.

Also look at `motion/*.ansi` in your terminal and say whether the motion reads
as subtle and short (≤ ~300 ms of frames); flag anything jarring as minor.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-v2-06-animations-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports v2-06-animations rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-06-animations-rN.md`. Do not touch other panes.
