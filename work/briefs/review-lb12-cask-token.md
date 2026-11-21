# Review brief — slice `lb12-cask-token`

> The slice contract is `work/briefs/lb12-cask-token.md` (with
> `work/briefs/cr-common.md`); it wins where they differ. You review; you do
> **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-lb12-cask-token.md`), including
fix commits. Rounds ≥ 2: confirm each earlier finding is fixed, then look for
new problems. Orchestrator commits touching only `work/` are not findings. A
receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate. Never run
brew write commands; do not start/stop/restart moshi-hook.

## Verify

- **V1 gate:** `make ci`; also `GOOS=darwin go build ./...`.
- **V2 regression tests:** revert the fix (check the pre-fix `casks.go` into a
  temp worktree) and confirm each new test fails there. The receipt says the
  brew-args assertion in `TestCaskUpgradeSendsTokenFromBrewPayload` was never
  shown failing (the confirm-text check fails first): prove the args check
  fails on its own, or report it.
- **V3 real shape:** read-only `brew info --json=v2 --installed --cask` and
  `brew outdated --json=v2 --cask` are allowed. Confirm the fixture matches
  the live shape (`token`, `name` list, `full_token`) and that the Casks panel
  shows the token for every installed cask.
- **V4 every consumer:** grep each reader of `Cask.Name`/`DisplayName`; any
  place that still sends a display name to brew, or keys a map by one, is a
  finding. Check the details header, the Outdated rows, search, the tap menu,
  batch actions, zap, and the dependency views.
- **V5 frames:** build your own harness: diff `View()` at the commit before AC-01
  and at HEAD for the Casks, Outdated and Search panels with a token that
  differs from its display name, at 80x24 and 140x40. Only the cask name text
  and the new "Display" row may differ; any other change is a finding.
- **V6 invariants:** Bubble Tea rules held; one commit per AC; floors not lowered.

## Severity

blocker (gate fails, regression, a display name still sent to brew) · major (AC
not met, test passes without the fix) · minor (docs/polish).

## Report

`work/inbox/reviewer-lb12-cask-token-rN.md` (≤ 50 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must achieve`.
Commit once (`inbox: reviewer reports lb12-cask-token rN`), worktree clean, reply
with exactly one line: `done: work/inbox/reviewer-lb12-cask-token-rN.md`. Do
not touch other panes. Write scratch output under /home/agent/.cache, not /tmp.
