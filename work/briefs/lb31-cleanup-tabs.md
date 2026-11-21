# Slice: lb31-cleanup-tabs — Cleanup and Unused tabs on Status

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb30-taps-and-trust` closed. User request (2026-11-21), backlog LB-29.
Never launch the TUI against the real brew; no brew write commands (read-only
`brew cleanup -n` and `brew autoremove -n` are fine to check the real shapes).

## Today

`c` on Status runs `brew cleanup -n` as a task (`cleanup-preview`, overlay), then asks to
run the real cleanup (`executeCleanup`); `A` runs `brew autoremove -n` the same way
(`runAutoremove`, `autoremove-preview`). The preview vanishes with the overlay. The
dry-runs go through the task manager (`streamWrite`).

## Behaviour

Status tabs become **Dashboard | Config | Doctor | Missing | Vulns | Cleanup | Unused**
(`Unused` = autoremove: shorter). All seven must be fully visible at 80x24 in both themes
(tighten the gap further if needed; do not abbreviate other names; if still too wide at 80,
say so and propose in a question rather than clipping).

- **Cleanup** loads when the tab is opened (read-only, cached in `tabContent`, kept across
  panel switches, reloaded by `R` once opened). Content: a header `Would free ~950MB` (brew's
  total line) and a table, biggest first: size (right-aligned, warning colour for the largest
  third), file count, path (shortened: strip the Homebrew prefix/cache dir to `~/…` or
  `<prefix>/…`, long paths cut in the **middle** with `…`, never mid-word otherwise). Empty:
  `✔ Nothing to clean up`. Error body as the other tabs.
- **Unused** loads the same way: the formulae `brew autoremove -n` would remove, one per
  line (name, version if available), with a header `N unused formulae would be removed`.
  Empty: `✔ Nothing unused`.
- **Keys** on Status: `c` and `A`. Not on the tab → switch to it and (re)load. **Already on
  that tab** → open the confirm to run the real command (text names the total/size or the
  count, and says it deletes files / removes formulae), through `task.Manager` as today
  (`executeCleanup`, the autoremove run). When it completes: toast as today, the tab reloads
  (the cache group is dropped) and shows the new state; the dashboard counts refresh.
- The old overlay preview flow (`cleanup-preview`, `autoremove-preview`, the pending-action
  chain in `gui.go`) goes away; delete what becomes dead.
- `y` copies the whole tab text (lb24), ANSI-free and uncut.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — brew layer: read-only previews and parsers
- [ ] AC-02 — the two tabs, rendering, 7 tabs fit at 80x24
- [ ] AC-03 — `c`/`A` open the tabs, confirm from the tab, remove the old preview flow

### AC-01
`Diagnostics.CleanupPreview(ctx) (CleanupPreview, error)` and `AutoremovePreview(ctx)`:
`Execute` (read-only, not a write, not the task manager) of `cleanup -n` / `autoremove -n`,
exit codes handled (check real behaviour of both on exit 0/1), parsed into typed data
(`Items{Path, Files, Bytes}` + `Total` for cleanup, parsing brew's `Would remove: <path> (N
files, 230.3MB)` and `This operation would free approximately 950MB`; names for autoremove).
Unparseable lines are kept as raw lines (nothing dropped silently); the view falls back to
raw text when nothing parsed. Fixtures: real `brew cleanup -n` output captured here
(paths anonymised), and autoremove output in both shapes (empty; with names: derive from
brew's source/docs if none is available here, and say so). Tests first.

### AC-02
Tab constants/entries, `fetchTabContentCmd` cases, view functions beside `doctorView`/
`missingView`, render cases, `needsTabFetch`, `R` reload, the tab-row spacing. Colour roles
from the theme (warning for large sizes, success for the ✔ empty lines). Frames at 80x24 and
140x40, both themes: each tab with data, empty, error; the 7-name tab row whole at 80x24.

### AC-03
`c`/`A` behaviour above; confirm text; completion reload; remove dead code (deadcode
clean); keymap/footer/help/README; a teatest flow: Status → `c` → tab shows the table →
`c` → confirm → brew args `cleanup` recorded → tab refreshed; same for `A`/`autoremove`.
Name each golden that changes.

## Done means

`make ci` green (`GOFLAGS=-p=1` if /tmp is tight); `GOOS=darwin go build ./...`; floors not
lowered; deadcode/staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Disk is tight: no large scratch. Receipt
`work/inbox/runner-lb31-cleanup-tabs.md` (≤ 45 lines, committed, not edited afterwards),
one commit per AC, tests first, subjects end `[LB-29]`.
