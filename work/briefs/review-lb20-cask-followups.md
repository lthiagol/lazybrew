# Review brief — slice `lb20-cask-followups`

> Contract: `work/briefs/lb20-cask-followups.md` (with `cr-common.md`). You
> review; you do **not** edit code, tests or commits. Review everything on `wip`
> after the commit that added the slice brief
> (`git log --oneline -- work/briefs/lb20-cask-followups.md`) up to HEAD.
> A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate. No brew
> write commands; read-only `brew info/search/list` are fine; never launch the
> TUI; scratch under /home/agent/.cache; do not touch other panes.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2:** revert each fix in a temp worktree; its test must fail there.
- **V3 real shape:** run `brew info --json=v2 --installed --cask` and `brew info
  --json=v2 --cask claude-code` (read-only) through the parser: `FullName` is
  `claude-code`/the tap-qualified form brew reports; the fixture matches the
  live shape. Check `installedFromTap`, the untap confirm and the trust menu
  now list installed casks and that formulae are unaffected; a cask in the
  default `homebrew/cask` tap and one from a third-party tap both behave.
- **V4 search info:** formula row → `--formula`, cask row → `--cask`; a name
  that is both; info output parses; the Search info pane frames are unchanged
  for normal rows (own `View()` diff at the base and HEAD, 80x24 and 140x40).
- **V5:** one commit per AC; floors not lowered; no new goroutines.

## Report
`work/inbox/reviewer-lb20-cask-followups-rN.md` (≤ 40 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table; findings `F-rN-01
(severity) — file:line — what — evidence — fix must achieve`. Commit once
(`inbox: reviewer reports lb20-cask-followups rN`), worktree clean, reply with
one line: `done: work/inbox/reviewer-lb20-cask-followups-rN.md`.
