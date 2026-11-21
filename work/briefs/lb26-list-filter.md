# Slice: lb26-list-filter — `/` filters the Formulae and Casks lists

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb25-search-polish` closed. User request (2026-11-21), backlog
LB-23. Never launch the TUI against the real brew; no brew write commands.

## Behaviour

On **Formulae** and **Casks**, `/` opens a local filter on the loaded list (no
brew call). On every other panel `/` stays the global search (jump to Search and
focus its box). To reach global search from Formulae or Casks: `7`, then `/`.

- A one-line filter row (`/ <query>`) shows at the top of that panel's sidebar
  box while the filter is open or active; the title count reads `matches/total`
  (`3/70`) while a query is active.
- Typing narrows the list as you type: case-insensitive substring on the name.
  The cursor goes to the first match; `up`/`down` (and only those) move the
  selection while the input is focused; `j`/`k`/`h`/`l`/letters type.
- `Enter` leaves the input and keeps the filter (list stays narrowed, `j`/`k`
  work again, `/` edits the query). `Esc` in the input clears the query and the
  filter; `Esc` on an active unfocused filter clears it too.
- No match: the list shows `No matches for "<q>"`; `Esc` clears.
- The Info/Deps/Used By tabs and every action (`i x r u p …`, `y` copy, `F`) act on
  the **selected filtered item**. Mouse clicks and the wheel map to the filtered
  rows. Switching panel keeps each panel's own filter; a refresh (`R`, data
  reload, auto-refresh, after an install/uninstall) reapplies the query to the new
  data and keeps the cursor on the same item when it is still listed.
- It composes with the Formulae **leaves** filter (`L`): the query narrows the
  leaves list, and toggling `L` re-derives from the full data (see
  `unfilteredItems`/`unfilteredFormulae` in `panel.go`; do not break leaves).
- Design it generic (a filter on `panelData` with an `id`-based allow-list, so
  Outdated/Taps/Services can be enabled later); enable it on Formulae and Casks
  only.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — filter state on `panelData`: matching, composes with leaves, survives reloads
- [ ] AC-02 — the `/` input: keys, row, title count, empty state
- [ ] AC-03 — keymap, help, footer, README, flows

### AC-01

Tests first (pure `panelData`): query narrows `items`+`formulae`/`casks` together;
selection and `selectedFormula/Cask/nameAt` index the filtered list; leaves on +
query; leaves toggled with a query active; `landData` with an active query
reapplies it and keeps the cursor on the same name when still present, else
clamps; clearing restores the full list and cursor on the same item. No UI yet.

### AC-02

`/` on Formulae/Casks opens the input (not on other panels; the global search
path is unchanged there); the keys above; the filter row and `matches/total`
title; the empty state. Add the model field(s) for the focused input following
how `searchInput` is held (textinput), without a second text-input path leaking
keys (`y`, `q`, digits, `?` type into the filter while it is focused; `ctrl+c`
still quits). Frame tests at 80x24 and 140x40, both themes: open, with query,
no match, after Enter. `7` then `/` reaches global search from Formulae.

### AC-03

Single keymap (`keybindings.go`): Formulae and Casks footers show `/ filter`
(their `/ search` hint goes; other panels keep `/ search`); help lists both
meanings. Keymap test passes. README (keys, the `7` then `/` path). A teatest flow:
Formulae, `/`, type, select a filtered row, `x` opens the uninstall confirm for
that row's name (assert on `View()` and the brew args with the mock), Esc
clears. Name each golden that changes.

## Done means

`make ci` green; `GOOS=darwin go build ./...`; floors not lowered; deadcode and
staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Disk is tight: no large scratch files. Receipt
`work/inbox/runner-lb26-list-filter.md` (≤ 40 lines, committed, not edited
afterwards), one commit per AC, tests first.
