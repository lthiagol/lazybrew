# Slice: v2-03-look-foundation — palette, titled panels, icons, header, one-line footer

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Third slice of `LB-08` (visual overhaul). `v2-01` moved to Charm v2, `v2-02`
made every frame exactly the terminal size with real overlays. This slice
builds the **look's foundation** in the style of the change-gate operator TUI
(read-only reference: `/home/agent/coding/moises-training-troubleshoot/tools/change-gate-tui/internal/ui/styles.go`
and `view.go` — `panel()`, `fit()`, `spread()`, `header()`, `footer()`).

It does **not** restructure what the sidebar and the main area show — that is
`v2-03b` (lazygit-style sidebar, user ruling `Q-v2-sidebar`). Here the
existing layout gets the new skin.

User rulings that apply (`work/questions.md`): `Q-v2-look` (Nerd Font icons by
default, Unicode fallback via `gui.icons: unicode` or `LAZYBREW_ICONS=unicode`;
accent stays Catppuccin **mauve** `#cba6f7`), `Q-v2-hints` (one-line footer),
`Q-v2-small` (state/count in the border title).

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes the tests. `TestEveryFrameIsTerminalSize` must pass at every commit.

### AC-01 — a real palette with semantic roles

1. `internal/gui/style` defines the full Catppuccin palette as two token sets:
   **Mocha** (dark) and **Latte** (light — replace the "pre-M14" light values
   with real Latte). Tokens at least: base, mantle, crust, surface0/1/2,
   overlay0/1/2, subtext0/1, text, mauve, lavender, blue, sapphire, teal,
   green, yellow, peach, red, pink. Use the official hex values.
2. Semantic roles on top, used by the rest of the GUI instead of raw tokens:
   Accent (mauve), Success (green), Warning (yellow), Error (red), Info (blue),
   Pinned (teal), Muted (overlay0), Subtle text (subtext0), Text, Border
   (surface1), Focus border (accent), Background (base). Keep the exported
   names other packages already use (`style.AccentColor`, `SubtleText`, …)
   working, mapped onto the roles, so the change is not a rename storm.
3. Remove the `t.BgColor == LightTheme().BgColor` detection hack in
   `ApplyTheme`; each theme carries its own surface tokens.
4. `Model.View()` sets the `tea.View` background and foreground colours to the
   theme's base and text (as change-gate does), so the app owns its canvas.
5. Tests: both themes define every token (non-nil, valid hex), the light theme
   is Latte (spot-check `base #eff1f5`, `mauve #8839ef`), and `ApplyTheme`
   switches every role.

### AC-02 — icons: Nerd Font by default, Unicode fallback, off switch

1. One `Icons` struct with two complete sets (Nerd Font, Unicode) — fields at
   least: App, Status, Formula, Cask, Outdated, Tap, Service, Search,
   Installed, Upgradable, Pinned, Running, Stopped, Error, Warn, OK, Info,
   Marker, Loading/Refresh.
2. Selection: `gui.icons: nerd|unicode` in config (default `nerd`, validated;
   unknown value → config error like other invalid fields), and
   `LAZYBREW_ICONS=nerd|unicode` in the environment **overrides** the config.
   The existing `gui.show_icons: false` turns icons **off** entirely (no
   glyph, no stray space) — it finally does something.
3. Tests: every Nerd glyph is in a Unicode Private Use Area; no Unicode-set
   glyph is; both sets fill every field; config default/override/off matrix;
   README's config block documents `icons` and `show_icons`.
4. In the frame harness and all tests, use the **Unicode** set (set the env
   var / config in the harness) so frames are readable in a plain file.

### AC-03 — one panel primitive with the title in the border

1. A single helper draws every box: rounded border, the **title inside the top
   border** at the left (`╭─ 2 Formulae ───── 125 ╮`), an optional right-hand
   label in the top border (count, `not loaded`, spinner, …), body lines
   fitted to the inner width (`fit`/truncate with `…`, never wrapping, never
   overflowing). Focused box: accent border and bold accent title. Unfocused:
   Border-role border, subtext title. Modelled on change-gate's `panel()`.
2. All sidebar boxes, the main panel and the Log box use it. The old
   `renderBox` and the in-box title lines go away. Sidebar boxes gain the row
   the title line used to take.
3. Sidebar boxes: title `N Name` (with the panel icon when icons are on), the
   right label shows the count when loaded, `not loaded`, `…`/spinner while
   loading, or `!` (error role) on error — so a collapsed box at 80×24 still
   shows its state (`Q-v2-small`).
4. Main panel: title is the breadcrumb (`Formulae › Info`). The tab bar stays
   inside the main panel body for now.
5. Log box: title `Log`, right label shows the running task title while busy;
   the busy/idle border colour rule (`logBusy`) is kept.
6. Mouse hit-testing still lands on the right element (tabs, list rows,
   sidebar boxes) — update `internal/gui/mouse.go` and keep/extend its tests.

### AC-04 — a header bar

1. Row 0 is a one-line header, full width:
   left — app icon + `lazybrew` (bold accent) + Homebrew version when known;
   middle — the current running operation (`⟳ upgrading curl`) or, when idle,
   nothing;
   right — badges: formulae/casks counts when known, `N outdated` (warning
   role) when known and > 0, and the update status (`updated 2m ago` /
   `never updated` / `updating…`) moved here from the bottom bar.
2. Unknown values are **omitted**, never shown as `0`. The header never wraps
   or overflows: badges drop from the left of the right group when space is
   short (`spread`-style), the app name always stays.

### AC-05 — a one-line, context-aware footer

1. The bottom bar becomes **one** row (the header took the other): the active
   panel's keys first, then the globals, styled `key` (bold accent) + `desc`
   (subtle), separated by spacing. Cut at the frame width; when anything is
   cut, the line ends with `? more`, which is always visible (`Q-v2-hints`).
2. While a modal is open, the footer shows that modal's keys (e.g. `y yes  n no
   esc cancel`); while help is open, help's keys.
3. `panelHints` stays the single source of truth for panel keys (its comment
   rule — every listed key has a handler — still holds). Tests: footer at 80
   and 140 columns for at least Formulae and Status; `? more` present when cut
   and absent when everything fits; modal footer.

### AC-06 — evidence and bookkeeping

1. Copy the v2-02 frame harness to
   `work/verification/v2-03-look-foundation/frame_dump_test.go.txt`, add
   Unicode-icon selection, same screens + sizes (22 frames). `before/` =
   copy of `work/verification/v2-02-layout-bugs/after/` (or re-captured from
   the slice's start commit — say which). `after/` = this slice's frames.
   Also save the `.ansi` frames for `after/140x40` (colour evidence for the
   reviewer and the user checkpoint).
2. `result.md`: one line per screen naming what changed; anything not
   explained by AC-01..AC-05 is a defect — fix it.
3. `work/STATE.md`: `| v2-03-look-foundation | — | implemented | awaiting reviewer |`.

## Invariants

- No sidebar/main **content** restructure (that is `v2-03b`): the same
  information appears in the same box as before, just restyled.
- No modal or toast restyle beyond palette roles (that is `v2-04`). No
  animation (`v2-06`).
- Every frame exactly the terminal size (`TestEveryFrameIsTerminalSize`) at
  every commit, all states, 140×40 / 80×24 / 200×60.
- No keybinding change. `task.Manager`, refresh and lazy-load policy untouched.
  `internal/brew/**` untouched.
- Tests whose expected strings encode the old look (e.g. `1 Status` on its own
  line) are updated to the new look — each named in the receipt. No test
  deleted, skipped, or weakened. Test count floor: the count at the slice's
  start commit.
- Coverage floors unchanged. No new dependencies.
- Never run brew write commands (`install`, `uninstall`, `upgrade`,
  `services start|stop|restart|run|cleanup`, `tap`, …) on this machine.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.
  `master-plan/` is not to be read. Worktree clean when you report.

## Gate

```bash
make build
make lint
CGO_ENABLED=1 go test ./... -race -count=1
make cover-check
```

## Report protocol

1. `work/inbox/runner-v2-03-look-foundation.md` (≤ 40 lines): commit ids in AC
   order, evidence per AC, gate, test count before/after, every changed test
   expectation, contradictions. Committed **once**.
2. Worktree clean. Reply with exactly one line:
   `done: work/inbox/runner-v2-03-look-foundation.md`. Do not touch other panes.

## Fix rounds

Findings come back in `work/briefs/fix-v2-03-look-foundation-rN.md`. Fix every
finding, one commit per finding (subject names the id), gate passing, receipt
`work/inbox/runner-v2-03-look-foundation-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` (never rewrite) and keep going on unaffected ACs.
