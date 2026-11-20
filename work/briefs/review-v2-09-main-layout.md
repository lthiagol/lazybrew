# Review brief — slice `v2-09-main-layout`

> The slice contract is `work/briefs/v2-09-main-layout.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-09-main-layout.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands (see the brief's Safety section).

## Verify

- **V1 gate:** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`.
- **V2 frames:** run the committed harness against real Homebrew; every
  frame exactly W×H at 80×24 and 140×40; compare each screen with
  `v2-08-preload/after/` and check every change is one `result.md` names.
  Look at the `.ansi` frames in your terminal. Remove `internal/gui/dump` after.
- **V3 padding (AC-01):** one column each side in every main-area view, tabs
  row and underline aligned with the content; nothing touches the border.
- **V4 tables (AC-02):** full width, columns aligned, last column takes the
  slack, cut with `…`, never wraps a row; installed icons still correct.
- **V5 two-column Info (AC-03):** only at the stated threshold; single column
  below it is unchanged in content; prose wraps inside its column.
- **V6 Log box (AC-04):** idle height, busy height, the sum of main + Log
  heights always the right column's height; no jump during an operation
  other than the one stated.
- **V7 invariants:** sidebar/header/footer/overlays untouched; no content
  dropped from any view; tests not weakened; one commit per AC.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-v2-09-main-layout-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports v2-09-main-layout rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-09-main-layout-rN.md`. Do not touch other panes.
