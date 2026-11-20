# Slice: cr4-gui-dedup — collapse the copy-paste in `internal/gui`

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `cr3-render-perf` closed.

## Goal

A refactor with **no behaviour change**. Remove the repeated enqueue tails and
task bodies, the duplicate helpers, the hand-synced key tables and the
oversized `Update` switch. Toasts, frames, commands sent to brew, and test
expectations stay the same. The receipt reports `wc -l` for
`internal/gui/commands.go` and `internal/gui/gui.go` before and after.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — `enqueue` helper [R1]
- [ ] AC-02 — `errTask` / `streamTask` / `linesTask` [R2, R4]
- [ ] AC-03 — selection and writer choice [R5, R6, Q2]
- [ ] AC-04 — shared small helpers [R7, R8, R10, R11]
- [ ] AC-05 — one keymap [R9, R12]
- [ ] AC-06 — `Update` cleanups [R13, Q3, Q4, Q5]
- [ ] AC-07 — named Casks tabs, `initPanels` [Q6]
- [ ] AC-08 — split the big files [Q1]

### AC-01 — `enqueue` helper [R1]

1. Add `func (m *Model) enqueue(t *task.Task) tea.Cmd`. It holds the
   repeated tail: `Enqueue`, the "Queue full" toast, the "Queued behind
   running task" toast, and `RunNext`.
2. All ~17 copies use it. At `c1aa732` they are in `executeTrustAction`,
   `trustItemCmd`, `executeUntap`, `executeRepair`, `executeBrewfileAction`,
   `doMutation`, `executeTapAdd`, `runServiceTask`, `togglePin`,
   `executeServicesCleanup`, `runVulns`, `runMissing`, `runDoctor`,
   `brewCleanup`, `runAutoremove`, `executeCleanup` and `executeAutoremove`;
   grep for more. The `brew update` helper from cr2 AC-04 uses it too. Keep
   the special handling it has today when the enqueue fails (the reset of
   `isUpdating`).
3. Tests: the queue-full and queued-behind toasts for one representative
   task each still pass, unchanged.

### AC-02 — `errTask` / `streamTask` / `linesTask` [R2, R4]

1. Add `errTask(id, title string, run func(ctx) error) *task.Task` for the
   single-error body (~9 copies) and
   `streamTask(id, title string, run func(ctx) (<-chan string, <-chan error)) *task.Task`
   for the stream body (~8 copies; it handles a nil line channel).
2. [R4] Add `linesTask(...)` for `runDoctor`, `runMissing` and `runVulns`,
   which today each write a goroutine that sends lines with a ctx guard. One
   helper with the same buffering for all three (`runVulns` currently
   differs; say what you settled on and why it is safe).
3. `executeCleanup`, `executeAutoremove`, `brewCleanup` and `runAutoremove`
   become one-liners on top of these helpers and `enqueue`.
4. Unit tests for each helper: success, error, ctx cancel mid-stream, nil
   line channel.

### AC-03 — selection and writer choice [R5, R6, Q2]

1. [R5] Every remaining `panel.selected >= len(panel.items)` check, name
   extraction and `name == ""` guard goes through `p.selectedName()` (added in
   cr1). In `doMutation`, the four `if casks { CasksWrite.X } else {
   FormulaeWrite.X }` arms become a writer chosen once, through a small
   unexported interface (Install/Uninstall/Reinstall/Upgrade).
2. [R6] Delete `selectedOutdatedFormula` (same body as `selectedFormula`)
   and switch its callers. Merge `selectServiceNamed`/`selectTapNamed` into
   one generic `selectNamed`.
3. [Q2] Task `Run` closures capture only what they use (`client :=
   m.client`, `isCask := …`), never the whole `m`. List the closures you
   changed.

### AC-04 — shared small helpers [R7, R8, R10, R11]

1. [R7] There is one `withIcon` and one bool-to-text helper (`boolStr`/
   `yesNo`), exported from `presentation` (or `style`). The duplicates in
   `gui/icons.go` and `render.go` are deleted.
2. [R8] One clip helper and one fit helper (cut and pad to width) replace
   `clipLine`, `cutCell` and the raw `ansi.Truncate(…, "…")` calls in
   `internal/gui` and `internal/gui/modal`. Put them in a small package, e.g.
   `internal/gui/textutil`, so `modal` can import it. Rename one of the two
   `fitLine`s so the names stop clashing. Tests: ANSI-styled input, wide
   runes, width 0 and 1.
3. [R10] `handleTabBarClick` uses `tabWidth()` instead of re-rendering each
   tab, and `handleMouseWheel` uses `l.sideW`. The existing mouse tests stay
   green.
4. [R11] `modal/menu.go` calls `leftBlock` (`frame.go`) instead of its inline
   copy.

### AC-05 — one keymap [R9, R12]

1. One table per context (each panel, global, and each modal) describes
   key → action, label and help text. `buildHelp` (`help.go`) and
   `panelHints`/`modalHints` (`footer.go`) are generated from it. Each modal
   exposes `Hints()` from the same source, and the duplicate structs
   (`helpItem`, `keyHint`) merge into one type.
2. Key dispatch for the global keys reads the same table where that stays
   simple. Panel-specific handlers can stay as switches, but their keys must
   appear in the table. A test fails if a key handled in dispatch is missing
   from help or the footer, or the other way round.
3. [R12] `"1"`…`"7"` become a single case computing the `PanelID`, and
   `[`/`left` and `]`/`right` are each combined into one case.
4. Help and footer frames are byte-identical to the current goldens.

### AC-06 — `Update` cleanups [R13, Q3, Q4, Q5]

1. [R13] Move the leaves filter out of the `Update` switch into a
   `panelData` method. In the `MutationResultMsg` handler, choose the toast
   once and return once.
2. [Q3] `confirmMutation` stores its label in `pendingLabel`, and the confirm
   path stops rebuilding the label map. No silent "Install" fallback remains.
3. [Q4] `confirmUninstall` doesn't run `brew uses` for casks. Test: a cask
   uninstall confirm records no `uses` call in the mock runner. This is the
   one intended behaviour change: one fewer brew call.
4. [Q5] Strings built with `+=` in loops (`commands.go` ~:1340, ~:1454) use
   `strings.Join` or a `strings.Builder`.

### AC-07 — named Casks tabs, `initPanels` [Q6]

1. The Casks tabs get named constants, like the other panels. All tab
   constants live in one file. `render.go` and `panel.go` use the names, with
   no bare `0`/`1`/`2`.
2. `initPanels` derives titles from `PanelID.String()`, sizes from a
   panel-count constant (not `make(…, 7)`), and drops the leftover
   `loading := false`.

### AC-08 — split the big files [Q1]

1. Move code only, with no edits other than imports: split `commands.go` and
   `gui.go` by concern, e.g. `keys.go` (key dispatch), `land.go` (data
   arrival), `tasks_packages.go`, `tasks_taps.go`, `tasks_services.go`,
   `tasks_diagnostics.go`. Choose the names and list the mapping in the
   receipt.
2. `git diff --stat` shows only moves. The reviewer will check with
   `git diff -M --color-moved=dimmed-zebra`. No file over ~800 lines.

## Invariants

The common invariants apply. In addition: no behaviour change except AC-06.3.
All goldens, snapshots and teatest flows are unchanged. Run deadcode and
staticcheck before the receipt; this slice must not add new unreachable
functions.
