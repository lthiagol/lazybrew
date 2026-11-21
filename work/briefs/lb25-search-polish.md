# Slice: lb25-search-polish — aligned Search rows, a green ✔ for installed, no duplicate `I`

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb24-copy-and-wording` closed. User requests (2026-11-21, macOS
smoke), rulings in `work/backlog.md` LB-22. Never launch the TUI against the real
brew; no brew write commands.

## AC-01 — the green ✔ is the installed mark everywhere

`style.Icons.Installed` becomes `✔` (U+2714, the glyph brew prints) in **both**
icon sets (Nerd and Unicode); `show_icons: false` keeps drawing nothing. It keeps
the green `InstalledBadge` role. It marks installed packages on Formulae and
Casks (they used a green circle-check/dot) and, in AC-02, on Search. Leave other
glyphs alone (outdated `↑`, tap `✓` trusted, service started/stopped, errors).
Name each golden that changes (every frame with a Formulae/Casks row) and check
the glyph is one cell wide in both themes (`lipgloss.Width == 1`; if a terminal
draws it wider the row alignment must still hold: say how you checked).

## AC-02 — Search rows on the cell system

Search rows are a plain string (`name  type  installed  description`) built in
`executeSearch`, so nothing aligns. Draw them like Formulae/Casks/Taps rows
(`sidebar_rows.go`, `sidebarCells`):
- marker, then the `✔` cell **only for installed rows**, a blank cell of the same
  width for the others (names stay aligned), then the name on the left;
- the type (`formula` / `cask`, the `unknown` of fixtures stays `unknown`)
  right-aligned in the subtle style, like `trusted` on Taps;
- the word `installed` disappears from the row; the description (when a row has
  one) stays out of the row as today.
Truncate long names with `…` through `textutil.Fit`, keeping the type visible.
`m.searchResults`/`Raw` stay the source (`selectedSearchRow`, `remarkSearch`,
`MarkInstalled` are unchanged). Frame tests (80x24 and 140x40, both themes) with
installed and not-installed rows and a long name; the right-aligned types line up
at one column; selection highlight and cursor unchanged. Update tests that
asserted on the old row string.

## AC-03 — drop the duplicate `I` on Search

On Search `I` does exactly what `i` does. Remove `I` from the dispatch
(`keys.go`), the key table (`sInstallCap` in `keybindings.go`), footer, help and
README; `i` stays. `I` on Search now does nothing. Keymap test passes and has a
case that `I` is not answered on Search. Name each golden/footer that changes.

## Done means

`make ci` green; `GOOS=darwin go build ./...`; floors not lowered; deadcode and
staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Receipt `work/inbox/runner-lb25-search-polish.md` (≤ 40 lines,
committed, not edited afterwards), one commit per AC, tests first.
