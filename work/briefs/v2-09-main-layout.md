# Slice: v2-09-main-layout — the main area uses its space

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

User feedback (ruling `Q-v2-next` in `work/questions.md`): the sidebar
looks good but the main panel looks loose — content is squeezed on the left
and the space is not used. The frames in `work/verification/v2-06-animations/
after/140x40/` show it: text touches the left border (no inner padding);
Formulae → Deps is one narrow list in a 89-column area; Tap Info is six short
lines over ~20 empty rows; the Log box always takes ~11 rows even when it
says `No commands executed yet`. Modelled on the change-gate TUI
(`/home/agent/coding/moises-training-troubleshoot/tools/change-gate-tui`):
read how it lays out its main area and tables before you start.

## ⚠ Safety

Never run brew write commands on this machine (`brew install|uninstall|
upgrade|reinstall|tap|untap|trust|untrust|pin|unpin|cleanup|autoremove|update`,
`brew services start|stop|restart|run|cleanup|kill`). `moshi-hook` is a live
service. Tests use the mock runner only.

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes the tests; `TestEveryFrameIsTerminalSize` passes at every commit.

### AC-01 — inner padding

1. Every main-area view (all panels and tabs, help excluded, the Operation
   view included) has one column of padding left and right inside the box.
   The tab row and its underline line up with the content.
2. Width maths: helpers that cut or wrap to the main width use the padded
   width (nothing overflows, nothing touches the border).

### AC-02 — lists become full-width tables

1. One table helper (aligned columns, a Muted header row, the last text
   column takes the slack and is cut with `…`, rows never wrap). Use
   lipgloss v2's `table` package or a small helper — no new dependency.
2. Used by: Formulae `Deps` and `Used By` (icon, name, installed version,
   description from data already loaded — empty cell when unknown, never a
   new brew call), Casks `Deps`, Search result deps, Taps `Formulae` (name,
   installed icon, a grid of several columns when names are short is
   acceptable instead), Formulae `Files` (path, keeping the v2-05 styling).
   Group headings (`Runtime`/`Build`, `Formulae`/`Casks`) stay.
3. Narrow widths (80×24) drop the least important columns first
   (description, then version) — state the order in `result.md`.

### AC-03 — two-column Info views when wide

1. When the main area's inner width is ≥ 100 columns, Formulae Info, Casks
   Info, Outdated Info, Tap Info and Services Info split into two columns:
   fields left, prose and lists (description, caveats, actions line,
   dependencies summary or tap's formula count…) right — choose per view and
   list the split in `result.md`. Below 100 columns, today's single column.
2. Prose wraps inside its column with the v2-05 hanging indent.

### AC-04 — the Log box gives the space back when idle

1. While no task runs and the log is empty, the Log box is its minimum
   (border + 1 row: `No commands executed yet`); with log lines but idle, at
   most 5 rows; while a task runs, today's size. The main box gets the rest;
   main + Log always exactly fill the right column.
2. The height change is instant (no animation); a test asserts the heights
   in each state at 80×24 and 140×40.

### AC-05 — evidence and bookkeeping

1. Harness copied to `work/verification/v2-09-main-layout/`, same screens,
   `before/` = `v2-08-preload/after/`, `after/` with `.ansi` at 140×40.
2. `result.md`: per screen what changed and why; `work/STATE.md` row
   `| v2-09-main-layout | — | implemented | awaiting reviewer |`.

## Invariants

- No sidebar, header, footer, modal, toast or animation change. No key change.
- No content dropped from any view (a column dropped at narrow width keeps
  its data reachable elsewhere — e.g. the Info tab).
- No new brew command.
- Every frame exactly the terminal size. No test deleted, skipped or
  weakened; changed expectations named. Coverage floors unchanged. No new
  dependencies.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.

## Report protocol

`work/inbox/runner-v2-09-main-layout.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-v2-09-main-layout.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-v2-09-main-layout-rN.md` → fix every finding, one commit each
(subject names the id), gate, receipt
`work/inbox/runner-v2-09-main-layout-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
