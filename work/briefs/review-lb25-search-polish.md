# Review brief — slice `lb25-search-polish`

> Contract: `work/briefs/lb25-search-polish.md` (with `cr-common.md`). You review;
> you do **not** edit code, tests or commits. Review everything on `wip` after the
> commit that added the slice brief
> (`git log --oneline -- work/briefs/lb25-search-polish.md`) up to HEAD. A receipt
> is a claim, not evidence. `CGO_ENABLED=1` for the race gate. No brew write
> commands; never launch the TUI against the real brew; scratch under
> /home/agent/.cache (disk is tight: delete your scratch when done); do not touch
> other panes; ignore untracked Screenshot*.png.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 glyph:** `✔` in both icon sets, green, one cell; `show_icons: false` draws
  none; Formulae/Casks rows keep their alignment and the outdated `↑`; the
  glyph is not used where another meaning is intended (taps, services).
- **V3 Search rows:** own `View()` harness (mock runner, live-shaped results
  incl. an installed cask, a formula, a name >30 chars, a name that is both
  kinds) at 80x24 and 140x40, both themes: names left-aligned after the mark
  column, types right-aligned on one column, `✔` only on installed rows,
  selection highlight intact, `…` truncation keeps the type visible. After an
  uninstall the mark clears (lb19 re-mark path). Info pane and install/uninstall
  args unchanged.
- **V4 `I`:** removed from dispatch, table, footer, help, README; `i` works;
  keymap test passes and fails if `I` is re-added (mutate).
- **V5 frames:** own diff base vs HEAD for every panel: only rows with the
  installed glyph, Search rows and the Search footer/help differ.
- **V6:** one commit per AC; floors not lowered; deadcode/staticcheck: nothing
  new vs the base.

## Report
`work/inbox/reviewer-lb25-search-polish-rN.md` (≤ 45 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table; findings `F-rN-01
(severity) — file:line — what — evidence — fix must achieve`. Commit once
(`inbox: reviewer reports lb25-search-polish rN`), worktree clean, reply with one
line: `done: work/inbox/reviewer-lb25-search-polish-rN.md`.
