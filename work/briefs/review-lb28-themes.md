# Review brief — slice `lb28-themes`

> Contract: `work/briefs/lb28-themes.md` (with `cr-common.md`). You review; you do
> **not** edit code, tests or commits. Review everything on `wip` after the commit
> that added the slice brief (`git log --oneline -- work/briefs/lb28-themes.md`) up
> to HEAD. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate
> (`GOFLAGS=-p=1` if /tmp is tight). No brew write commands; never launch the TUI
> against the real brew; scratch under /home/agent/.cache, deleted when done; do not
> touch other panes; ignore untracked Screenshot*.png.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 colors are right (the main check):** the palette hex in the brief were
  written from memory. Fetch the upstream sources (curl the raw files; read-only,
  into /home/agent/.cache, small): `catppuccin/palette` `palette.json` (Mocha,
  Macchiato, Frappé, Latte: every token), the Dracula spec, the Nord spec, and
  compare every value used in `palette.go` / the new themes with them. Any
  mismatch is a finding, even when the repo's own hex test passes (the test only
  repeats what the code says). If the network is unavailable, say so and verify by
  an independent second source.
- **V3 names/config:** each name loads; case-insensitive; `dark`/`light` equal
  Mocha/Latte values; unknown name errors with the list and exit non-zero; an
  old config with `theme: light` still works; `--config-sample` lists every name
  and the sample tests pass; README right.
- **V4 contrast and legibility:** recompute contrast for every theme and role
  yourself; look at a rendered frame (Formulae selected row, overlay, modal,
  footer, tabs, toasts, Taps trusted/untrusted, outdated) per theme in
  all eight themes at 80x24 and 140x40: nothing invisible
  (foreground equal to background), selection and borders distinguishable,
  the `Muted`/`Subtle` text readable.
- **V5 regression:** Mocha and Latte frames byte-identical to the base commit
  for every panel and overlay (own harness diff).
- **V6:** one commit per AC; floors not lowered; no new dependency; deadcode/
  staticcheck nothing new; no live switch, no new key, no env override was added.

## Report
`work/inbox/reviewer-lb28-themes-rN.md` (≤ 50 lines): first line `verdict: ok` or
`verdict: changes-required`; AC table; findings `F-rN-01 (severity) — file:line —
what — evidence — fix must achieve`. Commit once (`inbox: reviewer reports
lb28-themes rN`), worktree clean, reply with one line:
`done: work/inbox/reviewer-lb28-themes-rN.md`.
