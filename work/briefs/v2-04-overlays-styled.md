# Slice: v2-04-overlays-styled — confirm dialogs by action, dimmed backdrop, stacked toasts

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Fifth slice of `LB-08`. `v2-02` made modals and toasts real compositor layers;
`v2-03` gave the app its palette, icons and titled panels; `v2-03b` made the
sidebar the list. This slice gives the **overlays** the change-gate look
(read-only reference: `/home/agent/coding/moises-training-troubleshoot/tools/change-gate-tui/internal/ui/view.go`,
`confirmPanel()` and `actionSpecs` in `styles.go`).

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes the tests; `TestEveryFrameIsTerminalSize` passes at every commit.

### AC-01 — every confirm carries an action kind and its accent

1. `modal.NewConfirmModal` takes an **action kind** (e.g. `ActionDanger`,
   `ActionUpgrade`, `ActionInstall`, `ActionNeutral`) instead of relying on the
   title text. Each kind maps to a theme role: danger → Error (red), upgrade →
   Warning (yellow), install → Info (blue), neutral → Accent (mauve).
2. Every call site picks its kind: uninstall / zap / untap / cleanup /
   autoremove / services cleanup / services stop (when it exists) → danger;
   upgrade / upgrade all / batch upgrade / reinstall → upgrade; install /
   tap add / repair → install; everything else → neutral. List every call
   site and its kind in the receipt.

### AC-02 — the confirm dialog's new face

Modelled on change-gate's `confirmPanel`:
1. **Thick** border in the action's accent; padding `1, 3`.
2. A label chip on top: the action's name upper-cased, bold, base-coloured
   text on the accent background (` UNINSTALL `).
3. The question in bold text; below it, one subtle line saying what will
   happen (e.g. uninstall: `removes bash; 2 packages depend on it` using the
   dependents the uninstall check already computes; untap: the existing
   warning text).
4. Key line: `[y]es  [n]o  esc cancel` with the bracketed keys in the accent;
   the current Yes/No selection (←/→, h/l, enter) stays and is visible (the
   selected option bold + underlined or reverse).
5. Width: `min(frameWidth-8, 72)`, content centred, every line cut to the
   inner width (never wrapping past the box). At 80×24 the box fits.
6. Input and menu modals get the same frame (thick border in Accent, chip with
   their title, key line for their keys); menu items keep their selection
   marker in the accent.

### AC-03 — the backdrop dims while a modal is open

1. While any modal is open, the base frame behind it is drawn **dimmed**:
   every cell's foreground mapped to the theme's Muted/overlay role (or a
   faint attribute) — implement it at the compositing step, not by
   re-rendering each panel with special styles.
2. The modal itself is not dimmed. The footer keeps showing the modal's keys
   (v2-03 AC-05.2), readable (not dimmed).
3. Test: with a modal open, a cell outside the box carries the dim
   style/colour; the same cell without the modal does not; text content of
   cells outside the box is unchanged (only style differs).

### AC-04 — toasts: icon, box, stack, dismiss

1. A toast is a small rounded box in its type's role colour (border + icon):
   success → OK icon, error → Error icon, warning → Warn icon, info → Info
   icon (icon set from v2-03, honouring `show_icons`). Message in Text,
   cut with `…` to `min(frameWidth/2, 60)`.
2. Toasts **stack**: up to 3 visible, newest at the bottom, bottom-right above
   the footer, never covering the header or the footer. A 4th pushes the
   oldest out. Each auto-dismisses after its own lifetime (keep today's
   3 s; errors 6 s).
3. `x`/`esc` does **not** dismiss toasts (those keys are taken); a mouse click
   on a toast dismisses it. Keep the existing toast tick mechanism (tea.Cmd,
   no goroutines).
4. Fixes the pre-existing overlap noted in v2-03: at 80×24 a toast must not
   sit on a box border row in a way that breaks the border — it is a layer
   with its own border; check the 80×24 frame.
5. Tests: three toasts stack in order; a 4th evicts the oldest; per-type
   lifetime; click dismiss; frame size unchanged with 3 toasts at 80×24.

### AC-05 — evidence and bookkeeping

1. Harness copied to `work/verification/v2-04-overlays-styled/`; same screens
   and sizes **plus** `12-confirm-upgrade` (Formulae → `u` on an item, the
   upgrade confirm), `13-toast-stack` (three toasts of different types), and
   `14-input-modal` (Taps → `a`). `before/` = `v2-03b-sidebar-lists/after/`;
   `after/` with `.ansi` for 140×40 — **the dim backdrop must be visible in
   the `.ansi`**.
2. `result.md`: per screen what changed; anything else is a defect.
3. `work/STATE.md`: `| v2-04-overlays-styled | — | implemented | awaiting reviewer |`.

## Invariants

- Confirm semantics unchanged: same keys (y/n/enter/esc/←/→/h/l), same
  default selection, same result handling in `handleModalResult`.
- No change to what triggers a modal or a toast, or to toast messages' text
  (except cutting with `…`).
- No sidebar/main/header/footer restyle (v2-03/v2-03b), no animation (v2-06).
- Every frame exactly the terminal size. No keybinding change.
  `internal/brew/**` untouched. No new dependencies.
- No test deleted, skipped or weakened; changed expectations named in the
  receipt. Coverage floors unchanged.
- Never run brew write commands on this machine.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.

## Report protocol

`work/inbox/runner-v2-04-overlays-styled.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-v2-04-overlays-styled.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-v2-04-overlays-styled-rN.md` → fix every finding, one commit
each (subject names the id), gate, receipt
`work/inbox/runner-v2-04-overlays-styled-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
