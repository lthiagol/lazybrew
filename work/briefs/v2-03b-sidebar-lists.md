# Slice: v2-03b-sidebar-lists — lazygit-style: lists in the sidebar, details in the main area

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Fourth slice of `LB-08`. User ruling `Q-v2-sidebar` (`work/questions.md`):
**Option B, lazygit-style.** The sidebar's stacked panels hold the **lists**;
the main area shows **only the selected item's details and tabs — never the
list again.** Today the two duplicate each other: the Status dashboard is drawn
twice, and the Formulae list appears in the sidebar *and* as the main area's
`List` tab table.

`v2-03-look-foundation` gave the app its palette, titled panels, icons, header
and footer. Build on those helpers; do not restyle them.

## What each panel shows after this slice

| panel | sidebar (the list) | main area (tabs) |
|---|---|---|
| 1 Status | compact summary rows: Homebrew version, prefix, counts | `Dashboard` (the full dashboard, as today), `Config`, `Doctor` |
| 2 Formulae | one row per formula: status icon, name, version (`1.2 → 1.3` when outdated), pinned icon | **`List` tab removed.** `Info`, `Deps`, `Used By`, `Caveats`, `Files`. `Info` shows the tap (the table's old `Tap` column). |
| 3 Casks | status icon, name, version | `Info`, `Deps`, `Caveats` (as today) |
| 4 Outdated | selection mark (batch), status icon, name, `installed → latest` | `Info` (as today); `Versions` becomes the **selected item's** installed vs. latest, pinned state, and the batch summary (`N selected — u to upgrade, U all`) — not the list again |
| 5 Taps | trust icon, name, `official`/`trusted`/`untrusted` | `Tap Info` = the selected tap's details plus the one-line summary header (`3 taps · 0 official · 3 trusted · 0 untrusted`) — **not** the list; `Trust`, `Formulae` as today |
| 6 Services | running/stopped icon, name, status | `Status` = the selected service's details from the data already loaded (name, status, user, file, exit code) — not the list. (Full `brew services info` comes in the Services slice.) |
| 7 Search | the results list | input box + the selected result's info (as today) |

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes the tests; `TestEveryFrameIsTerminalSize` passes at every commit.

### AC-01 — the sidebar is wide enough to be the list

1. The sidebar width rule changes so a list row fits icon + name + version:
   `gui.sidebar_width` stays a **percentage** (default changes 30 → 35), the
   absolute cap rises from 40 to 60 columns, the floor stays 20. Document the
   default in README's config block.
2. Sidebar rows are built from the panel's typed data (formulae, casks, taps,
   services) with the icon set, not from the preformatted `items` strings
   that today end in `...`. Long names truncate with `…`; versions are
   right-aligned; `1.2 → 1.3` uses the warning role.
3. Selection: the selected row in the **focused** box is highlighted across
   the full row width (surface0 background + marker); in an unfocused box the
   marker is shown without the highlight.

### AC-02 — Formulae: the List tab goes, the table's data moves

1. Remove the `List` tab and the main-area formulae table. The tab indices of
   the remaining Formulae tabs shift; every place that hard-codes a Formulae
   tab index (`render.go`, `commands.go` `needsTabFetch`/`loadTabContent`/
   `renderCachedTabContent`, mouse hit-testing, help, tests) is updated —
   use named constants for the tab indices so this cannot drift again.
2. The default Formulae tab is `Info`. `Info` shows the tap.
3. Everything the table offered survives: sorting (outdated, pinned, name),
   the leaves filter (`L`), and mouse click-to-select now act on the sidebar
   list. The table code that becomes dead is removed (with its tests moved to
   cover the sidebar rows where the behaviour survives — named in the receipt).

### AC-03 — the other panels stop re-drawing their list

1. Status, Outdated (`Versions`), Taps (`Tap Info`) and Services (`Status`)
   main tabs render the selected item's details as in the table above, never
   the list. `renderList` is no longer called for the main area.
2. The Outdated batch selection (`Space`, `a`, `u`, `U`) works from the
   sidebar list and the `Versions` tab shows the batch summary.
3. The main-list scrolling from v2-02 (`listWindow`) moves to the sidebar
   where the lists now live; the sidebar's own `offset`/`visibleCount`
   behaviour stays coherent (one source of truth for each box's window,
   computed in the Update loop or purely at render time — never written at
   render time).

4. Status data is typed (orchestrator ruling Q-v2-03-b): the dashboard load
   carries typed counts (formulae, casks, taps, services, outdated when known)
   into the model, not only text lines. The Status sidebar summary and the
   header badges use them, so the header shows `N formulae` / `N casks` on
   first launch, before the Formulae/Casks panels are opened. Unknown values
   are still omitted, never `0`. Test: fresh model after the dashboard load →
   header contains the counts.

### AC-04 — mouse, keys and help follow the new layout

1. Click a sidebar row → focuses that panel and selects the row (as today for
   sidebar clicks); wheel over a sidebar box scrolls that list; click a tab →
   switches tab; wheel over the main area scrolls the tab content.
2. `j/k`/arrows move the selection in the focused list; `[ ]` switch tabs;
   all other keys unchanged.
3. Help text and `panelHints` reflect the removed `List` tab where they
   mention it.

### AC-05 — evidence and bookkeeping

1. Harness copied to `work/verification/v2-03b-sidebar-lists/`, same screens
   and sizes; screen `02-formulae` now shows the Info tab of the first
   formula (it used to be the List table) — note it in `result.md`. `before/`
   = `v2-03-look-foundation/after/`; `after/` = new frames, `.ansi` for
   140×40 included.
2. `result.md`: per screen, what changed; anything not explained by the table
   above / AC-01..AC-04 is a defect — fix it.
3. `work/STATE.md`: `| v2-03b-sidebar-lists | — | implemented | awaiting reviewer |`.

## Invariants

- No palette/panel-primitive/header/footer restyle (v2-03 owns those); no
  modal/toast restyle (v2-04); no animation (v2-06). No new data fetched
  (Services' `brew services info` is a later slice).
- Every frame exactly the terminal size in every state at 140×40, 80×24,
  200×60.
- No keybinding removed or remapped. `task.Manager`, refresh and lazy-load
  policy untouched. `internal/brew/**` untouched.
- No test deleted, skipped or weakened without a replacement covering the
  surviving behaviour; every changed expectation named in the receipt. Test
  count floor = the count at the slice's start commit, minus only tests of
  removed code that have a named replacement.
- Never run brew write commands on this machine.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.
  Worktree clean when you report.

## Gate

`make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.

## Report protocol

`work/inbox/runner-v2-03b-sidebar-lists.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-v2-03b-sidebar-lists.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-v2-03b-sidebar-lists-rN.md` → fix every finding, one commit
each (subject names the id), gate, receipt
`work/inbox/runner-v2-03b-sidebar-lists-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
