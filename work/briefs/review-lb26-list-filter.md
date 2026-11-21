# Review brief — slice `lb26-list-filter`

> Contract: `work/briefs/lb26-list-filter.md` (with `cr-common.md`). You review; you
> do **not** edit code, tests or commits. Review everything on `wip` after the
> commit that added the slice brief
> (`git log --oneline -- work/briefs/lb26-list-filter.md`) up to HEAD. A receipt is
> a claim, not evidence. `CGO_ENABLED=1` for the race gate (use `GOFLAGS=-p=1` if
> /tmp is tight). No brew write commands; never launch the TUI against the real
> brew; scratch under /home/agent/.cache and deleted when done; do not touch other
> panes; ignore untracked Screenshot*.png.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 behaviour (own harness, mock runner, live-shaped data):** Formulae and Casks:
  `/` opens the filter; typing narrows case-insensitively; first match selected;
  `up`/`down` move, `j/k/h/l/y/q/?/digits` type; Enter keeps it and `j/k` work;
  Esc clears (twice from the unfocused state); no match shows the message; title
  `m/total`. Other panels: `/` still jumps to Search and focuses it; `7` then `/`
  from Formulae works.
- **V3 correctness:** the selected filtered row drives Info/Deps tab fetches, `x`,
  `u`, `i`, `r`, `p`, `y`, `F` (args carry that row's name, not the unfiltered
  index): test several rows, including after Enter and after a scroll. Mouse click
  and wheel hit filtered rows. Leaves (`L`) on/off with a query, in both orders.
  Refresh/data reload/auto-refresh/uninstall completion with an active query:
  reapplied, cursor kept on the same item or clamped, no panic on an empty
  result, no stale item left selected. Switching panels keeps each filter.
- **V4 keys:** no collision in the table, `ctrl+c` still quits from the input,
  the keymap test passes and fails if `/ filter` is dropped from Formulae
  (mutate); help and README right.
- **V5 frames:** own diff base vs HEAD for every panel (80x24, 140x40, both
  themes): only Formulae/Casks footers, help, and filtered states differ.
- **V6:** one commit per AC; floors not lowered; no new goroutines; no
  `program.Send`; deadcode/staticcheck: nothing new.

## Report
`work/inbox/reviewer-lb26-list-filter-rN.md` (≤ 50 lines): first line `verdict: ok`
or `verdict: changes-required`; AC table; findings `F-rN-01 (severity) — file:line —
what — evidence — fix must achieve`. Commit once (`inbox: reviewer reports
lb26-list-filter rN`), worktree clean, reply with one line:
`done: work/inbox/reviewer-lb26-list-filter-rN.md`.
