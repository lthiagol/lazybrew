# Review brief — slice `lb21-backlog-sweep`

> Contract: `work/briefs/lb21-backlog-sweep.md` (with `cr-common.md`). You review;
> you do **not** edit code, tests or commits. Review everything on `wip` after the
> commit that added the slice brief
> (`git log --oneline -- work/briefs/lb21-backlog-sweep.md`) up to HEAD. A receipt
> is a claim, not evidence. `CGO_ENABLED=1` for the race gate. No brew write
> commands; never launch the TUI; scratch under /home/agent/.cache; do not touch
> other panes.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 AC-01:** footers on every panel at 80x24 and 140x40, both themes: the new
  hints show, nothing overflows or wraps badly, the keymap test passes with the
  smaller `footerExempt` and fails if a footer hint is dropped (mutate it). Own
  `View()` diff base vs HEAD: only footer rows differ.
- **V3 AC-02:** `go test -race -count=200` on the singleflight test passes; it
  fails when the sharing is reverted; no sleep and no new production hook.
- **V4 AC-03:** with debug on, nothing reaches stderr (run the code path with
  stderr captured); the log file receives it; no exported symbol is left only
  for a test except the reset helper; run deadcode/staticcheck from the installed
  binaries (path in the brief) and compare with `tools` output at the base.
- **V5 AC-04:** the flow runs in the stated time under `-race`; revert the lb19
  Tab fix (and only `shift+tab`) in a temp worktree: it fails both ways.
- **V6:** one commit per AC; floors not lowered; no test deleted without its code.

## Report
`work/inbox/reviewer-lb21-backlog-sweep-rN.md` (≤ 50 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table; findings `F-rN-01
(severity) — file:line — what — evidence — fix must achieve`. Commit once
(`inbox: reviewer reports lb21-backlog-sweep rN`), worktree clean, reply with one
line: `done: work/inbox/reviewer-lb21-backlog-sweep-rN.md`.
