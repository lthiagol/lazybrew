# Review brief — slice `lb31-cleanup-tabs`

> Contract: `work/briefs/lb31-cleanup-tabs.md` (with `cr-common.md`). You review; you do
> **not** edit code, tests or commits. Review everything on `wip` after the commit that
> added the slice brief (`git log --oneline -- work/briefs/lb31-cleanup-tabs.md`) up to
> HEAD. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate
> (`GOFLAGS=-p=1` if /tmp is tight). No brew write commands (read-only `brew cleanup -n`,
> `brew autoremove -n` are fine); never launch the TUI against the real brew; scratch under
> /home/agent/.cache, deleted when done; do not touch other panes; ignore untracked
> Screenshot*.png.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 real shapes:** run `brew cleanup -n` and `brew autoremove -n` here (read-only) and
  feed the exact output through the parsers: every `Would remove` line (sizes in
  MB/KB/GB/B, file counts with thousands separators as brew prints them: check `1,700`),
  the total line, odd lines kept raw, empty output, exit codes. Revert AC-01: tests fail.
- **V3 tabs:** own harness (mock runner, live-shaped data): Cleanup table biggest first,
  sizes right-aligned, large-third in warning colour, paths shortened and cut in the middle,
  header total; Unused list; empty ✔ lines; error bodies; loaded when opened (one read),
  cached across panel switches, reloaded by `R` once opened and never before; the
  7-name tab row whole at 80x24 and 140x40, both themes; `y` copies the whole text.
- **V4 keys:** `c`/`A` off the tab switch and load; on the tab open the confirm (names
  size/count and the consequence); confirm → exactly `cleanup` / `autoremove` through
  `task.Manager`; cancel runs nothing; after completion the tab and dashboard refresh;
  Services `c` (service cleanup) unchanged; the old preview overlay flow is gone and
  nothing is dead (deadcode/staticcheck nothing new vs base); keymap test passes and fails
  when a hint is dropped (mutate).
- **V5 frames:** own diff base vs HEAD every panel (80x24, 140x40, both themes): only
  Status (tab row, new tabs, footer/help) differs.
- **V6:** one commit per AC; floors not lowered; no new goroutines in handlers; no
  `program.Send`.

## Report
`work/inbox/reviewer-lb31-cleanup-tabs-rN.md` (≤ 55 lines): first line `verdict: ok` or
`verdict: changes-required`; AC table; findings `F-rN-01 (severity) — file:line — what —
evidence — fix must achieve`. Commit once (`inbox: reviewer reports lb31-cleanup-tabs rN`),
worktree clean, reply with one line: `done: work/inbox/reviewer-lb31-cleanup-tabs-rN.md`.
