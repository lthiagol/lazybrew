# Slice: v2-02-layout-bugs — every frame fits the terminal, overlays float, data is honest

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Second slice of `LB-08` (visual overhaul). `v2-01-charm-v2` moved lazybrew to
the Charm v2 stack with no visual change. This slice fixes the **layout and
data bugs** found in the orchestrator's review, **in the current style**. The
restyle (colours, title-in-border panels, header, icons, the lazygit-style
sidebar) is the **next** slice, `v2-03` — do not start it here.

The bugs, as observed on real frames (`work/verification/v2-01-charm-v2/baseline/`):

| # | bug | where |
|---|---|---|
| B1 | the frame does not fill the terminal: at 140×40 the main view is 38 rows and 138 used columns — 2 rows at the bottom and 2 columns at the right are dead | `render()` in `internal/gui/gui.go`, `renderMainPanel` / `renderBottomBar` in `internal/gui/render.go` |
| B2 | a modal is **appended below** a view that is already full height, so it lands off-screen (`07-confirm-modal.txt` is 42 rows in a 40-row terminal) | `internal/gui/gui.go` `render()`, the `m.activeModal != nil` block |
| B3 | the confirm modal shows `Loading...` until the terminal is resized: it only becomes ready on a `tea.WindowSizeMsg`, which it never receives after it is opened | `internal/gui/modal/confirm.go` (`ready`); check `input.go` and `menu.go` for the same pattern |
| B4 | a toast is appended **below** the bottom bar, so every toast pushes the whole layout down one row | `internal/gui/gui.go` `render()`, the `toastView` block |
| B5 | a panel that has never been loaded says `(empty)` (Casks says `(empty)` with 2 casks installed) — indistinguishable from "loaded, nothing there" | `internal/gui/panel.go:301` `renderSidebarContent` |
| B6 | the Log box is rendered **inside** the main panel's border — a double border, and it eats main-panel width | `renderMainPanel` / `renderCommandLogBox` in `internal/gui/render.go` |
| B7 | the Formulae **Deps** tab says `No dependencies` for every installed formula (`bash` really depends on `ncurses` and `readline`) | `internal/brew/casks.go:65` `flexStrings` + `internal/brew/formulae.go` `parseFormula` |

**B7 root cause (verified by the orchestrator):** for an installed formula,
`brew info --json=v2` gives `installed[0].runtime_dependencies` as objects
shaped `{"full_name": "ncurses", "version": "6.6", ...}`. `flexStrings` only
reads a `"name"` key from objects, so every entry is dropped and the list is
empty. Run `brew info --json=v2 bash` yourself to see the real shape.

## Tools

Lip Gloss v2 has a compositor: `lipgloss.NewLayer(content).X(x).Y(y).Z(z)`,
`lipgloss.NewCompositor(layers...)`, `lipgloss.NewCanvas(w, h)` /
`canvas.Compose(...)` / `canvas.Render()`. Read
`$(go env GOMODCACHE)/charm.land/lipgloss/v2@v2.1.1/layer.go`, `canvas.go`
and its `README.md`/examples before you use them — use the real API, not
memory. This is how B2 and B4 must be solved: the modal and the toast are
**layers drawn on top of** the full-size base frame, not strings joined to it.

## Acceptance criteria

One commit per criterion, in order. Stage by name — never `git add -A`, never
`git add .`. Every commit must build and pass the tests.

### AC-01 — every frame is exactly the terminal size (B1)

1. In every state — normal (each of the 7 panels), help open, a modal open,
   a toast showing, an operation running/finished, terminal-too-small —
   `View().Content` is **exactly** `m.height` rows and every row is **exactly**
   `m.width` cells (`lipgloss.Width` per line; pad, never overflow).
2. The 2 dead rows and 2 dead columns go to the existing layout: the main
   panel and sidebar grow to fill the height, the main panel to fill the
   width. No new content is added to fill them.
3. A table-driven test in `internal/gui` asserts the size rule for every state
   above at **140×40, 80×24 and 200×60**. This test is the guard for the rest
   of the overhaul; make it easy to add states to.

### AC-02 — modals float over the screen (B2)

1. While a modal (confirm, input, menu) is open, the frame is the normal
   screen with the modal box **composited on top**, centred horizontally and
   vertically, using the Lip Gloss v2 compositor.
2. Rows and cells outside the modal's box are identical to the frame without
   the modal (test it: render with and without, compare the rows above and
   below the modal box).
3. The modal's own look (its box, title, message, buttons) does not change in
   this slice beyond what centring requires.
4. Mouse: while a modal is open, today's mouse behaviour must not regress
   (read `internal/gui/mouse.go`; if clicks were ignored while a modal was
   open, they still are).

### AC-03 — modals render their content on the first frame (B3)

1. A modal shows its real content on the very first frame after it opens —
   never `Loading...`. Give it the size it needs when it is created or when
   the model opens it (e.g. pass the current width/height, or have the model
   forward its last `tea.WindowSizeMsg` to the new modal); remove the
   `ready` gate if it is no longer needed.
2. A later terminal resize while the modal is open still re-sizes it.
3. Test: open the uninstall confirm on a model that has already received its
   `WindowSizeMsg` (no resize afterwards) and assert the first `View()`
   contains the confirm's title and question and does **not** contain
   `Loading...`. Same for the input modal (tap add) and a menu modal
   (brewfile or trust menu).

### AC-04 — toasts float, the layout never moves (B4)

1. A toast is a layer composited over the base frame, bottom-right, just
   above the bottom bar (inside the frame, never below it). Keep today's
   toast styling (colour per type, bold, message text).
2. Test: render the same model with and without a toast; every row except
   the row(s) the toast covers is identical, and the frame size is unchanged.

### AC-05 — "never loaded" is not "empty" (B5)

1. A panel that has never been fetched shows a subtle `not loaded` (same
   `SubtleText` style) in the sidebar. A panel that was fetched and has no
   items keeps `(empty)`. Loading (`...`) and error rendering are unchanged.
2. Track this with explicit state (e.g. a `loaded bool` set when a fetch
   result lands, error or not), not by inferring from `items == nil` in the
   renderer. The R-refresh rule at `internal/gui/gui.go` (`if p.items == nil
   { continue // never loaded }`) keeps its meaning; if you switch it to the
   new flag, keep the behaviour identical.
3. Tests: fresh model → Casks sidebar says `not loaded`; after a fetch result
   with zero casks → `(empty)`; after a fetch result with items → the items.

### AC-06 — the Log is a sibling, not a box inside a box (B6)

1. The command Log is drawn **below** the main panel, as its own box at the
   same width as the main panel — not inside the main panel's border. There is
   exactly one border around the main content and one around the Log.
2. Collapsing the Log (`C`) gives its rows back to the main panel, as today.
3. The Log keeps its title, its busy/idle border colour rule (`logBusy`), and
   its content.
4. Mouse hit-testing that depends on the main panel's geometry (tabs, list
   rows) still hits the right element — update `internal/gui/mouse.go` if the
   geometry moved, and keep/extend its tests.

### AC-07 — the Deps tab shows real dependencies (B7)

1. `flexStrings` accepts objects carrying `"full_name"` (and still `"name"`;
   prefer `"full_name"` when both exist, since that is what brew uses for
   tap-qualified deps).
2. In `parseFormula`, an installed formula's runtime dependencies come from
   `installed[0].runtime_dependencies`; a formula with **no** installed entry
   falls back to the top-level `dependencies` list. Build deps are unchanged.
3. Tests in `internal/brew` with a fixture shaped like the **real** output
   (objects with `full_name`, `version`, `declared_directly`, …) — e.g. extend
   `testdata/formulae_installed.json` or add a new fixture — asserting `bash`'s
   deps are `ncurses` and `readline`. Keep every existing parser test passing.
4. Check the other places `flexStrings` is used (casks) still parse their
   fixtures identically.

### AC-08 — evidence and bookkeeping

1. Port the v2 frame harness
   (`work/verification/v2-01-charm-v2/frame_dump_v2_test.go.txt`) into a new
   `work/verification/v2-02-layout-bugs/frame_dump_test.go.txt`: the same nine
   screens and sizes **plus** two new screens — `10-toast` (any toast
   showing over the Formulae list) and `11-deps-bash` (Formulae → select
   `bash` → Deps tab). Run it against real Homebrew; commit the
   ANSI-stripped frames to `work/verification/v2-02-layout-bugs/after/`.
   `internal/gui/dump/` is never committed and must not exist when you report.
2. `work/verification/v2-02-layout-bugs/result.md`: for each of the nine old
   screens, `diff` the v2-01 baseline frame against the new one and state in
   one line which bug(s) explain every changed region. **Any change not
   explained by B1–B7 is a defect — fix it, don't explain it away.**
3. `work/STATE.md`: add `| v2-02-layout-bugs | — | implemented | awaiting reviewer |`.

## Invariants

- **No restyle.** No new colours, no theme changes, no border style changes,
  no new glyphs/icons, no header bar, no sidebar restructure. Those are v2-03.
- No keybinding changes. No change to `task.Manager`, the Operation overlay's
  content, or the refresh/lazy-load policy (except AC-05's explicit flag,
  behaviour-identical).
- `internal/brew/**` changes are limited to AC-07 (`flexStrings`,
  `parseFormula`, their tests and fixtures).
- Bubble Tea rules in `AGENTS.md` hold. No new dependencies.
- No test deleted, skipped, or weakened. Today's test count is the floor:
  `grep -rho "^func Test\w*" --include=*_test.go . | wc -l` (410 at
  `143ee61`). Tests whose expected strings encode the bugs (e.g. a frame that
  expected the modal below the screen) are **updated to the correct
  behaviour**, and each such update is named in the receipt.
- Coverage floors unchanged; `make cover-check` passes.
- One commit per AC, in order, staged by name. Do not amend, rebase, revert,
  or push. `master-plan/` is not to be read or modified. Worktree clean when
  you report.

## Gate

```bash
make build
make lint
CGO_ENABLED=1 go test ./... -race -count=1   # CGO is off by default on this machine
make cover-check
```

All four must pass; record real output in the receipt.

## Report protocol

1. Write `work/inbox/runner-v2-02-layout-bugs.md` (≤ 40 lines): commit ids in
   AC order, each AC with its evidence (command + output), the gate, the test
   count before/after, every test whose expectation you changed and why, any
   contradiction found. Get it under the limit **before** committing; commit
   it **once**:
   `git add work/inbox/runner-v2-02-layout-bugs.md && git commit -m "inbox: runner reports v2-02-layout-bugs"`.
2. `git status` clean. Reply in your session with exactly one line:
   `done: work/inbox/runner-v2-02-layout-bugs.md`. Do **not** send keys or text
   to any other pane.

## Fix rounds

Findings come back in `work/briefs/fix-v2-02-layout-bugs-rN.md`. Fix **every**
finding, one commit per finding (subject names the id), gate passing, receipt
`work/inbox/runner-v2-02-layout-bugs-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` (never rewrite) and keep going on the criteria
the question does not affect.
