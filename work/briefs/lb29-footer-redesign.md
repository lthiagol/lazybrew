# Slice: lb29-footer-redesign — grouped footer, dim separators, `←↑↓→ navigate`

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb28-themes` closed. User request and rulings (2026-11-21), backlog
LB-26. Never launch the TUI against the real brew; no brew write commands.

## Today

`footerLine` (`footer.go`) joins `key desc` pairs with two spaces: the panel's own
hints, then `globalHints`, and cuts with `? more` when the row is too wide. Each panel
also lists `j/k navigate` and `h/l tabs`, and the global list has `S-Tab prev`.

## Target (140 columns, Formulae)

```
 i install · x uninstall · r reinstall · u upgrade · p pin · L leaves · y copy  │  ←↑↓→ navigate · tab next · 1-7 jump · / filter · R refresh · C log · ? help · q quit
```

- **Two groups**, in this order: the panel's own actions, then the global keys,
  separated by a wider gap with a dim `│`. Hints inside a group are separated by a
  dim `·` (one space either side). Separators use the muted style (`style.MutedText`
  or the role the theme gives the dimmest text that is still visible); keys keep the
  accent style, words the subtle style, as today.
- **One arrow hint**: `←↑↓→ navigate` replaces the per-panel `j/k navigate` and
  `h/l tabs`. The arrows act as today: up/down move the list, left/right switch
  tabs (the hint says `navigate`; no tabs hint any more).
- **Hidden from the footer, kept on the `?` help screen**: `j/k`, `h/l`, and
  `S-Tab prev` (`Shift+Tab`). They keep working. Help lists all of them (Scroll list,
  Switch tabs, Previous panel) so nothing is undocumented.
- Global group: `←↑↓→ navigate · tab next · 1-7 jump · / search (or filter) · R
  refresh · C log · ? help · q quit`. The `/` hint is `search` or `filter` as
  `globalFooterHints` decides today (lb26).
- Modals and the help screen use the same `·` separator, with no groups.

## Narrow widths

The row is one line (`bottomBarRows`). At 80 columns everything cannot fit, so it
degrades by priority, never mid-word, never wrapping: first the panel actions
from the right end drop one at a time, **but** at least the first three panel
actions stay while any global hint beyond `? help` and `←↑↓→ navigate` remains;
global hints drop in this order: `C log`, `R refresh`, `1-7 jump`, `tab next`,
`/ search`, `q quit`; `←↑↓→ navigate` and `? help` stay last. When anything was
dropped, `? more` ends the row as today (it replaces `? help`). The builder checks
the rule against every panel at 80x24 and 100x30 and reports the 80-col result of
each panel in the receipt (the row as drawn).

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — the footer renderer: groups, dim separators, priority fit
- [ ] AC-02 — the key tables: one arrow hint; `j/k`, `h/l`, `S-Tab` help-only
- [ ] AC-03 — frames, help, README, flows

### AC-01
Rewrite `footerLine` (and modal/help hint rows) per the target. The hint lists
carry group membership (panel vs global) without breaking `keymap_test`'s model:
extend `keyHint` only if needed. Tests first: table of widths (60..200) × a
representative hint set: the row is never wider than the width, never ends with a
separator, keeps `? more` rules, drops in the stated order; separators drawn in the
muted role (assert the SGR); both themes. Pure function tests; no model needed.

### AC-02
New `keyArrows` hint (`←↑↓→`, keys `up down left right`, desc `navigate`) in the
global list; `keyNavigate` and `keyPanelTabs` stop being footer hints (their keys
`j k h l` become help-only entries: no `desc`, a `help` line); `keyPrev` loses its
footer `desc` (help-only). Remove the per-panel `keyNavigate`/`keyPanelTabs` from
`panelHints`/`tapsHints`/`servicesHints`, and the `footerExempt` entries that this
makes unnecessary. `keymap_test.go` passes unchanged in spirit (footer and help and
dispatch agree) and **fails** if the arrow hint is dropped (mutate it) or if `h`/`j`
disappears from help. Left/right on panels with no tabs (Search) and the Services
Logs tab behave as before.

### AC-03
Frames at 80x24 and 140x40, both themes, **all eight panel/tab families** plus modals
and an open overlay: the footer matches the target and the narrow rules; no clipped
word. README key tables and the `?` help text updated (and `y`, the filter). A
teatest flow asserting the footer text at 140 cols on Formulae and Status. Name each
golden that changes (every frame's last row). Own `View()` diff base vs HEAD: only
the footer row differs.

## Done means

`make ci` green (`GOFLAGS=-p=1` if /tmp is tight); `GOOS=darwin go build ./...`;
floors not lowered; deadcode/staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Disk is tight: no large scratch. Receipt
`work/inbox/runner-lb29-footer-redesign.md` (≤ 45 lines, committed, not edited
afterwards), one commit per AC, tests first, subjects end `[LB-26]`.
