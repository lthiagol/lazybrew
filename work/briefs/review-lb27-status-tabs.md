# Review brief — slice `lb27-status-tabs`

> Contract: `work/briefs/lb27-status-tabs.md` (with `cr-common.md`). You review; you do
> **not** edit code, tests or commits. Review everything on `wip` after the commit
> that added the slice brief (`git log --oneline -- work/briefs/lb27-status-tabs.md`)
> up to HEAD. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate
> (`GOFLAGS=-p=1` if /tmp is tight). No brew write commands; read-only `brew
> missing` is fine, `brew vulns` at most once (network). Never launch the TUI;
> scratch under /home/agent/.cache, deleted when done; do not touch other panes;
> ignore untracked Screenshot*.png.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 Missing:** real `brew missing` output (exit 0 empty, exit 1 with rows) through
  the service and tab: rows or the satisfied line; cached across a panel switch;
  `R` reloads it once opened; error body; revert the fix: tests fail.
- **V3 Vulns:** opening the tab runs nothing (mock runner records zero calls);
  `v` runs once, not twice while running; not-run hint, running, report, empty,
  exit-1, error and timeout states; `R` never runs it; leaving and returning
  keeps the report; the real one-off `brew vulns` output (once) renders sanely
  wrapped at 80 and 140 cols, and `y` copies it whole.
- **V4 keys/tabs:** `d`, `m`, `v` switch to their tab from any Status tab and from
  the dashboard; `h/l`/arrows cycle all five tabs; Status footer/help/README right;
  the overlay runners are gone and nothing is dead (deadcode/staticcheck: nothing
  new vs the base); keymap test passes and fails when a hint is dropped (mutate).
- **V5 frames:** own diff base vs HEAD for every panel (80x24, 140x40, both
  themes): only Status (tabs row, new tabs, footer, help) differs.
- **V6:** one commit per AC, in order; floors not lowered; no new goroutines in
  handlers, no `program.Send`; the vulns read is bounded.

## Report
`work/inbox/reviewer-lb27-status-tabs-rN.md` (≤ 50 lines): first line `verdict: ok`
or `verdict: changes-required`; AC table; findings `F-rN-01 (severity) — file:line —
what — evidence — fix must achieve`. Commit once (`inbox: reviewer reports
lb27-status-tabs rN`), worktree clean, reply with one line:
`done: work/inbox/reviewer-lb27-status-tabs-rN.md`.
