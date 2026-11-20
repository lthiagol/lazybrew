# Slice: v2-05-content-polish — detail views, dashboard, scrollable help, operation progress

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Sixth slice of `LB-08`. The frame, palette, icons, panels, sidebar lists and
overlays are done (`v2-02` … `v2-04`). This slice polishes **what the main
area says**: the detail views, the Status dashboard, help, and the Operation
view. Rulings that apply (`work/questions.md`): `Q-v2-help` (help scrolls),
`Q-v2-cask-deps` (show cask dependencies).

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes the tests; `TestEveryFrameIsTerminalSize` passes at every commit.

### AC-01 — one label/value style for every detail view

1. A single helper renders detail fields: label in the Muted role, padded to
   one column per view; value in Text; URLs in the Info role; empty values
   omitted (never `Label:` followed by nothing). Long values cut with `…` at
   the main width — or wrapped with a hanging indent for prose fields
   (description, caveats).
2. Used by: Formulae Info, Casks Info, Outdated Info, Taps Tap Info, Services
   Status, Search result info, Status Config. `presentation.FormatFormulaInfo`
   / `FormatCaskInfo` move to it (keep `internal/gui/presentation` coverage ≥
   its floor; update its snapshot tests — each changed expectation named).
3. Status values use the icon + role: `✓ installed` (Success), `↑ outdated
   1.2 → 1.3` (Warning), pinned (Pinned role), keg-only noted.

### AC-02 — dependency lists that say something

1. Formulae `Deps` and `Used By`, Casks `Deps`, Search result dependencies:
   one row per package with an icon showing whether it is **installed**
   (Success) or **not installed** (Muted), using data already loaded (the
   Formulae list); grouped under `Runtime` / `Build` headings where known.
2. Cask dependencies (ruling `Q-v2-cask-deps`): parse brew's object shape for
   cask `depends_on` (e.g. `{"formula": ["python@3.14"], "cask": [...],
   "macos": {...}, "arch": [...]}`) in `internal/brew` and show `Formulae`,
   `Casks`, and requirements (`macOS ≥ …`, `arch`) when present. Fixture from
   real output (`brew info --json=v2 --cask --installed` on this machine —
   read-only). Existing cask parser tests keep passing.
3. Formulae `Files`: paths styled (directory part Muted, file name Text), a
   count line on top.

### AC-03 — the Status dashboard as cards

1. `Dashboard` tab: small titled sections (Homebrew, Packages, Taps,
   Services, Health) instead of one text block. Packages shows formulae /
   casks counts and an **outdated gauge** (a bar like change-gate's
   `━━━━╌╌╌`, Warning when > 0, Success when 0; `not loaded — press R` when
   unknown, never `0`). Services shows running / stopped with icons. Health
   shows the doctor summary when known (`✓ ready to brew` / `⚠ N warnings`)
   and `press d to run doctor` otherwise.
2. `Doctor` tab: each warning as its own block — a Warn-icon heading line and
   the body in Text, separated by a blank line; `✓ Your system is ready to
   brew.` in Success when clean.
3. Layout adapts: cards side by side when the main width allows (≥ 2 × card
   width), stacked otherwise; nothing overflows at 80×24.

### AC-04 — help scrolls and lines up

1. Help (`?`) is a scrollable view: `j/k`, arrows, `PgUp/PgDn`, `Home/End`,
   mouse wheel; a position hint (`12–34 of 61`) in the border's right label;
   `?`/`esc` close. At 80×24 every line of help is reachable.
2. Sections line up: one key column width per section, headings in Accent,
   keys bold Accent, descriptions Subtle; two columns when width allows, one
   otherwise; no section heading stranded at the bottom of a column without
   its rows.
3. The `(g/G outside Search)` wrap noted in v2-03b r3 is fixed by the new
   layout.

### AC-05 — the Operation view shows progress

1. The Operation view (main area while a task runs / just finished) gets:
   a title line with the action icon and the package; **elapsed time**
   (`12s`, `1m04s`) updating while running (use the existing tick, no new
   goroutine); a **step list** built from brew's `==> ` lines (completed steps
   `✓` Success, current step with the spinner, later output lines Muted under
   their step); the final state `✓ Completed in 34s` / `✗ Failed: <err>` /
   `Cancelled`.
2. **Batch** operations (batch upgrade, upgrade all) show a real progress bar
   (`charm.land/bubbles/v2/progress`) with `n/total` — this bar is static
   width updates here; spring animation comes in `v2-06`.
3. The raw output remains available: the last lines stay visible under the
   steps (tail), and the Log box keeps the full command log as today.
4. Tests: step parsing from a recorded `brew upgrade` output fixture (make it
   from real output **without** running an upgrade — use a known transcript
   shape or brew's documented `==>` lines), elapsed formatting, final states,
   batch bar `n/total`.

### AC-06 — evidence and bookkeeping

1. Harness copied to `work/verification/v2-05-content-polish/`; same screens
   and sizes plus `16-help-scrolled` (help, scrolled down a page),
   `17-cask-info` (Casks → first cask Info), `18-doctor` only if doctor output
   is cached/cheap — otherwise skip and say so (do **not** run `brew doctor`
   repeatedly). The Operation view is covered by tests, not frames (no brew
   writes). `before/` = `v2-04-overlays-styled/after/` (frames 01–15); `after/` with `.ansi`
   for 140×40.
2. `result.md`: per screen what changed; anything else is a defect.
3. `work/STATE.md`: `| v2-05-content-polish | — | implemented | awaiting reviewer |`.

## Invariants

- No sidebar/header/footer/panel-primitive restyle (v2-03/03b), no modal or
  toast restyle (v2-04), no animation (v2-06).
- No new brew command is run except read-only ones already used (`info`,
  `list`, `services list`, `tap-info`, `outdated`, `config`, `doctor` where
  already wired). `internal/brew/**` changes limited to AC-02.2.
- Every frame exactly the terminal size. No keybinding change except help's
  scroll keys (inside help only).
- New dependency allowed: `charm.land/bubbles/v2/progress` (already in the
  bubbles module). Nothing else.
- No test deleted, skipped or weakened; changed expectations named.
  Coverage floors unchanged.
- Never run brew write commands on this machine.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.

## Report protocol

`work/inbox/runner-v2-05-content-polish.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-v2-05-content-polish.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-v2-05-content-polish-rN.md` → fix every finding, one commit
each (subject names the id), gate, receipt
`work/inbox/runner-v2-05-content-polish-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
