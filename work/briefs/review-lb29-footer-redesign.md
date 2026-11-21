# Review brief — slice `lb29-footer-redesign`

> Contract: `work/briefs/lb29-footer-redesign.md` (with `cr-common.md`). You review; you
> do **not** edit code, tests or commits. Review everything on `wip` after the commit
> that added the slice brief (`git log --oneline -- work/briefs/lb29-footer-redesign.md`)
> up to HEAD. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate
> (`GOFLAGS=-p=1` if /tmp is tight). No brew write commands; never launch the TUI against
> the real brew; scratch under /home/agent/.cache, deleted when done; do not touch other
> panes; ignore untracked Screenshot*.png.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 the footer as drawn:** build your own harness (mock runner) and dump the last
  row for every panel and tab (Status 5 tabs, Formulae, Casks, Outdated, Taps, Services
  incl. Logs, Search with and without results), with a modal, the help screen and a
  failed/running overlay, at 80x24, 100x30, 140x40, 200x50, both themes. Compare with
  the brief's target and narrow rules: groups and order, dim `·` and `│` separators
  (check the SGR is the muted role, not the accent), `←↑↓→ navigate`, `tab next`,
  `1-7 jump`, `/ filter` on Formulae/Casks and `/ search` elsewhere, `? help`/`? more`;
  never wider than the terminal, never a trailing separator, no clipped word, the
  drop order exactly as stated (try widths 60..160 step 2 on three panels).
- **V3 keys still work and are documented:** `j k h l down up left right`, `shift+tab`,
  `tab`, `1-7` behave as before on every panel; `?` help lists `j/k`, `h/l`,
  `Shift+Tab`; the keymap test passes and fails when the arrow hint is dropped
  or `h`/`j` is removed from help (mutate both).
- **V4 regression:** own frame diff base vs HEAD for every panel/overlay: **only the
  last row differs** (and the help screen lines the brief names). Goldens changed in
  the diff are footer rows only.
- **V5:** one commit per AC; floors not lowered; no new dependency; deadcode/staticcheck
  nothing new; README matches the real footers.

## Report
`work/inbox/reviewer-lb29-footer-redesign-rN.md` (≤ 55 lines): first line `verdict: ok`
or `verdict: changes-required`; AC table; findings `F-rN-01 (severity) — file:line —
what — evidence — fix must achieve`. Commit once (`inbox: reviewer reports
lb29-footer-redesign rN`), worktree clean, reply with one line:
`done: work/inbox/reviewer-lb29-footer-redesign-rN.md`.
