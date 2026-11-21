# Review brief — slice `lb30-taps-and-trust`

> Contract: `work/briefs/lb30-taps-and-trust.md` (with `cr-common.md`). You review; you do
> **not** edit code, tests or commits. Review everything on `wip` after the commit that
> added the slice brief (`git log --oneline -- work/briefs/lb30-taps-and-trust.md`) up to
> HEAD. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate
> (`GOFLAGS=-p=1` if /tmp is tight). No brew write commands (read-only `brew trust
> --json=v1`, `brew tap-info` are fine); never launch the TUI against the real brew;
> scratch under /home/agent/.cache, deleted when done; do not touch other panes; ignore
> untracked Screenshot*.png.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 real shape:** feed the real `brew trust --json=v1` output of this machine through
  the reader: every tap, formula, cask and command appears with the right type and tap
  part; empty lists; one list missing from the JSON; malformed. Revert AC-01: tests fail.
- **V3 the pane (own harness, mock runner, live-shaped data):** taps then formulae, casks,
  commands (alphabetical in each); title `Taps & Trust` and count; item rows draw ✔ name
  and a right-aligned type, long names truncate with `…` keeping the type; the cursor,
  wheel and mouse hit the right row across both groups; selecting a trusted row shows the
  single `Info` tab with Name/Type/Tap/Tap trusted and the sentence; `h/l` do nothing there;
  tap rows keep all their tabs and behaviour (`t`, `T`, `x`, `r`, `a`) — compare with the
  base commit.
- **V4 actions:** `t` on an item → confirm naming the consequence → exactly `untrust
  --formula|--cask|--command <name>` through `task.Manager` (mock records args; a cask and
  a command too); `x`/`r` on an item toast and run nothing; `T` on an item row says it needs
  a tap; after completion both lists reload, the item is gone, the cursor stays on the same
  row or the nearest; an untrust that fails leaves the row and shows the error.
- **V5 reloads/failures:** `R`, auto-refresh, preload and the post-action reload keep both
  groups in step; a trust-list read failure keeps the taps usable, no item rows, one error
  toast (not repeated every refresh); an empty trust list is silent.
- **V6 frames/keys:** own frame diff base vs HEAD, every panel (80x24, 140x40, both
  themes): only the Taps pane (title, rows, Info, footer/help) and help pages that name the
  pane differ; the keymap test passes and fails if the item-row hints are dropped.
- **V7:** one commit per AC; floors not lowered; deadcode/staticcheck nothing new; no
  `program.Send`/raw goroutines; README right.

## Report
`work/inbox/reviewer-lb30-taps-and-trust-rN.md` (≤ 55 lines): first line `verdict: ok` or
`verdict: changes-required`; AC table; findings `F-rN-01 (severity) — file:line — what —
evidence — fix must achieve`. Commit once (`inbox: reviewer reports lb30-taps-and-trust rN`),
worktree clean, reply with one line: `done: work/inbox/reviewer-lb30-taps-and-trust-rN.md`.
