# Fix brief — lb31-cleanup-tabs, round r1 (before review)

Rulings: `Q-lb31-tabs-80`, `Q-lb31-cleanup-autoremove` (end of `work/questions.md`). Tests
first, one commit, subject ends `[Q-lb31-tabs-80]`.

Implement the windowed tab row: when a panel's tab names do not fit the main box width, draw
a window onto the row that always contains the active tab, with `‹` on the left and/or `›` on
the right in the muted style when tabs are cut on that side; whole tab names only (a tab
that does not fit entirely is not shown), the active underline intact. When the row fits it
is drawn **exactly as today** (every existing frame unchanged). Mouse clicks on the tab row
(`handleTabBarClick`) must map through the window (click a visible tab selects it; clicking a
marker moves one tab that way). `h/l`/arrows keep cycling all tabs. Tests: Status at 80x24,
each of the seven tabs active: the active name is visible, markers correct, row ≤ width, both
themes; the same at 60 columns on Status and Formulae; a click test; no change to frames where
the row fits (140x40 all panels). Update `TestStatusTabRowWhole140x40`-style tests and name each
golden that changes. Gate: `GOFLAGS=-p=1 make ci`. Never launch the TUI. Append "Fix r1" (≤ 8
lines) to `work/inbox/runner-lb31-cleanup-tabs.md`, commit, do not edit afterwards; reply with
one done line.
