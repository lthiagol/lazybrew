# Slice: lb19-keys-and-search — Tab wraps in Search, simpler tab keys, honest search rows

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies (bug fixes start with a
> failing test). If the repo contradicts the brief, follow the brief and
> report the contradiction. Don't stop to ask.

**Depends on:** `lb18-config-sample` closed. Found by the user in the Run 2
smoke (2026-11-21).

## Findings

1. **Tab sticks on Search.** Landing on Search with no results focuses the
   search input, and `keys.go` hands every key to the focused input, `tab` and
   `shift+tab` included. Tab from panel 7 never reaches panel 1 (and Shift+Tab
   from 1 lands on Search and stays). Fix: `tab`/`shift+tab` cycle panels even
   while the input is focused (blur it as the panel changes).
2. **Search rows say `unknown` and never `installed`.** Without a TTY, `brew
   search <q>` prints no `==> Formulae`/`==> Casks` headers, only names with a
   blank line between the two groups, so `parseSearchOutput` marks every row
   `unknown`; and nothing sets `SearchResult.Installed`. Fix: ask brew for the
   type explicitly (`brew search --formula <q>` and `brew search --cask <q>`,
   two reads, bounded like today) so each row's type is certain, and mark
   `Installed` from the installed formula and cask lists (the cached
   `Formulae.List`/`Casks.List`, matched on name/token; a tap-qualified name
   matches its last part). Keep the headered-output parser path for fixtures.
   Without a description there is nothing to show: keep the row layout, drop
   the `unknown` word. No `--desc` (it searches descriptions and returns
   unrelated rows).
3. **Keys.** Tabs switch with `left`/`right` and `h`/`l` only; `[` and `]` are
   removed. `PgUp`/`PgDn` are removed everywhere (lists and the help screen).
   `j`/`k`/`up`/`down` stay. The mouse wheel stays.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — Tab and Shift+Tab cycle past a focused search input
- [ ] AC-02 — Search rows carry the real type and installed state
- [ ] AC-03 — `h`/`l` switch tabs; `[` `]` and PgUp/PgDn removed
- [ ] AC-04 — Search can uninstall what is installed (user request, 2026-11-21)

### AC-01

Failing test first: from Search with the input focused, `tab` goes to the
first panel and `shift+tab` to the previous panel; from Status, `shift+tab`
reaches Search and the next `shift+tab` leaves it. Typing a `tab` into the
query is not wanted. Cycling both ways through all 7 panels returns to the
start (teatest flow asserting on `View()`).

### AC-02

1. Failing tests first, with the **real** headerless shape:
   `"claude-code-router\nclaude-code-templates\n\nclaude-code\nclaude-code@latest\n"`
   through the two-call path: formula rows are `formula`, cask rows are `cask`,
   none `unknown`.
2. `claude-code` (an installed cask on this machine) shows `installed`; an
   uninstalled name does not. Test with a mock runner and mock installed lists.
3. If the installed lists cannot be read, the rows still show their type, with
   no installed mark, and no error toast.
4. Search → info and install/uninstall on a row still send the right name and
   `--cask` flag (the type is now certain: check `selectedName`/the install
   arg builder).

### AC-03

1. Remove `[`, `]`, `pgup`, `pgdown` from `keybindings.go` (the single keymap),
   `keys.go`, `help.go` (the help screen scroll), footers, help pages and
   README. Add `h`/`l`. `keymap_test.go` must pass with footer and help
   matching the new keys, and have a case proving `[`/`]`/PgUp/PgDn do nothing.
2. `h` and `l` must not collide with an existing binding (grep the key table);
   while the search input or any text input is focused they type text.
3. Update `SMOKE-CHECKLIST.md` Run 2 lines that mention `[` `]` and PgUp/PgDn.
4. Name every golden/teatest expectation that changes (footers, help) and why.

### AC-04 — uninstall from Search

Search lists installed rows (after AC-02) but only offers install. Add:

1. `x` on a Search row that is **installed** opens the same uninstall confirm
   the Formulae/Casks panels use (formula, or cask with `--cask`, from the
   row's certain type), through `task.Manager`, naming the consequence. `X`
   (zap) on an installed cask row, as on the Casks panel.
2. `x` on a row that is not installed: info toast "not installed", nothing
   runs. `i` on an installed row: info toast "already installed, x to
   uninstall", nothing runs (or, if it already does something sensible, keep
   it and say so in the receipt).
3. After the uninstall completes the row's installed mark clears (invalidate
   the installed lists, then re-mark the current results), without a new
   search.
4. The Search footer and help show `x uninstall` (single keymap; keymap test
   covers it); README. Name each golden that changes.
5. Tests first: installed cask row → `x` → confirm → brew args `uninstall
   --cask claude-code`; installed formula row → `uninstall <name>`; not
   installed → toast and no task; after completion the row shows no
   `installed`.

## Done means

- `make ci` green; `GOOS=darwin go build ./...` ok; floors not lowered.
- Invariants: Bubble Tea rules; no write outside `task.Manager`; no raw
  goroutines in handlers.
