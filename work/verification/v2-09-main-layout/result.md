# v2-09-main-layout — result

The main area uses its space: one column of padding either side, full-width
tables, two-column Info when wide, and a Log box that gives its rows back when
idle. The sidebar, header, footer, modals, toasts and keys are unchanged.

## Commits

Listed in git order, oldest first (`git log --reverse`). One commit per AC was the
rule, and this slice breaks it once: the AC-03 refinement (7f6e93b) landed after
AC-04 (539a515). It was a fix to AC-03's left column found while checking the
AC-04 work, and it was committed as a new commit, not by amending or rebasing,
since the git rules forbid both. The AC-03 and AC-04 order breach is recorded here
and is not undone.

| # | AC | commit | subject |
|---|---|---|---|
| 1 | AC-01 | fd1bde3 | feat(gui): main area keeps one column of padding either side (v2-09 AC-01) |
| 2 | AC-02 | 12a3b1b | feat(gui): lists become full-width tables (v2-09 AC-02) |
| 3 | AC-03 | 95766a6 | feat(gui): wide Info views split into fields and prose columns (v2-09 AC-03) |
| 4 | AC-04 | 539a515 | feat(gui): the Log box gives its space back when idle (v2-09 AC-04) |
| 5 | AC-03 refinement (out of order) | 7f6e93b | feat(gui): the Info left column is as wide as its fields (v2-09 AC-03) |
| 6 | AC-05 | d61ea76 | docs(plan): v2-09 main layout evidence, result, STATE row, Q-v2-09-a (v2-09 AC-05) |
| 7 | receipt | 228b145 | inbox: runner receipt for v2-09-main-layout |

Fix round r1 (`work/briefs/fix-v2-09-main-layout-r1.md`), after the receipt above:

| # | Finding | commit | subject |
|---|---|---|---|
| 8 | F-r1-01 | 801054e | fix(gui): a long description is cut with … instead of dropped (v2-09 F-r1-01) |

The F-r1-02 and F-r1-03 commits are listed in `work/inbox/runner-v2-09-main-layout-fix-r1.md`.

## Gate

Run at every AC and again on the final tree:

- `make build` — ok
- `make lint` (`go vet ./...`) — ok
- `CGO_ENABLED=1 go test ./... -race -count=1` — every package ok
- `make cover-check` — "All coverage floors met"

`TestEveryFrameIsTerminalSize` passes at every commit. No test was deleted or
skipped. Expectations that moved are listed under each AC below.

## AC-01 — inner padding

- `render.go`: `mainPadX = 1` and `mainContentWidth()` (main box less its border
  and the padding either side). `renderMainPanel` lays the tab bar and the content
  out to that width and pads them one column either side. The tab row and its
  underline move with the content.
- `scroll.go`: `contentLast` and the scroll window use the padded width.
- `mouse.go`: a tab click subtracts `mainPadX`, so clicks land on the drawn tab.
- The Log box is padded too (review r3, the orchestrator's ruling): the same one
  column either side, from `mainPadX`. The first AC-01 commit left it unpadded, as a
  separate box outside the brief's "main-area views". Its text no longer touches its
  border. See "Log padding" below.
- Test: `main_padding_test.go` checks that every body row of the main box has a
  blank column at both inner edges, for every panel state at 140x40 and 80x24.

## AC-02 — lists become full-width tables

`table.go` is a small helper, no new dependency. `renderTable` draws:

- a header row in the Muted role;
- columns as wide as their widest cell, the last kept column taking the slack, so
  a table fills the main width;
- cells cut with `…`; rows never wrap.

Drop order at narrow widths: **Description first (drop 2), then Version (drop 1).**
The icon and the name are never dropped.

A column is dropped only when the kept columns, each at its minimum useful width,
do not fit. The minimum useful width is **12 cells** (`tableMinUseful`); a column
narrower than that already (a version, a short name) keeps its natural width. So a
long description is cut with `…` to the slack and is kept, and it goes only when
even 12 cells of it cannot fit beside the name and the version. (Review F-r1-01:
the first AC-02 rule dropped Description as soon as one description was wider than
the slack.)

Every kept column keeps at least its floor, `min(natural, 12)`: when the columns
do not fit, the other columns (the name) are cut down to their floors before the
last column is squeezed, so the description is never a sliver (review F-r2-01: at
46 cells a 29-character name had left the description at `G…`). A column goes only
when the floors of the kept columns cannot all be met, in the order above. As a last
resort the other columns are cut to 4 cells (`tableMinW`).

At 80x24 the Deps frames keep Description and cut it: `Library for command-line …`.
At 140x40 every description fits, so those frames do not change.

Where it is used:

- Formulae **Deps** and **Used By**: icon, Name, Version, Description. Version and
  Description come from the loaded installed-formulae list, and stay empty when the
  package is not in it. No new brew call.
- Casks **Deps**: the same columns, with version and description from the loaded
  installed-casks list.
- **Search** result dependencies: the same table, from the same loaded lists.
- Taps **Formulae**: the same package table (icon, Name, Version, Description).
- Formulae **Files**: a single `Path` column; the count line stays above it. The
  directory is Muted and the file name is Text, as in v2-05.
- Group headings (`Runtime`/`Build`, `Formulae`/`Casks`/`Requirements`) stay, as
  heading rows with a blank row between groups.

Structural change: the Deps tab was cached as a rendered string at load time, which
has no width. It is now drawn from the loaded formula at render time
(`renderContent`), and `renderCachedTabContent` stores nothing for it. Icons off:
no icon column, and the state is written after the name, as before.

Expectation changes, all in tests that checked the same facts:

- `depviews_test.go`: checks for `"✓ ncurses"` became `hasRow(out, "✓", "ncurses")`,
  because the icon and the name are now separate columns. The Files test expects
  the `Path` header on line 2 and the path on line 3, and `renderFiles` takes the
  width.
- `gui_test.go`: the Deps tests read the rendered view (`renderContent`) instead of
  the stored `tabContent` string.
- New tests: `table_test.go` (alignment, slack, cut, drop order, muted header,
  heading rows, package rows carry only loaded data, icons off).

## AC-03 — two-column Info views when wide

`presentation.Info` takes the fields. A field with `Right: true` goes to the right
column. Below the threshold it is one column, in the listed order, as today.

Threshold (Q-v2-09-a, ruled by the orchestrator, r3): the main box's **inner width
≥ 84** splits the view. In code this is one named constant, `splitWidth` in
`presentation/detail.go`, set to **82**, the content width at inner 84 (the padding
takes a cell either side). The first AC-03 threshold was inner ≥ 100; at the default
140x40 layout (`sidebar_width` 35, inner 89) that never split, which is why the
ruling lowered it.

Split per view (left = fields, right = prose and lists):

| View | Left column | Right column |
|---|---|---|
| Formulae Info, Outdated (formula) | Name, Version, Tap, Status, License, Homepage, Bottled | Description |
| Casks Info, Outdated (cask) | Name, Version, Tap, Status, Homepage | Description |
| Tap Info | Tap, Trusted, Official | Remote, Formulae and Casks counts |
| Services Info | Name … Exit code, Interval, Cron | Unit file, Command, Working dir, Log, Error log |

Services' actions line stays under both columns. The left column is as wide as its
fields need, but at least 2/5 of the width, and it leaves the right column at least
`infoMinRight` = 26 cells; a gap of 4 separates them. (The first AC-03 commit used a
fixed 2/5, which cut the Unit file path. The refinement commit 7f6e93b used up to
3/5, which cut long left values such as the Remote and Unit file.) Right-column
values wrap with the v2-05 hanging indent and are never cut.

Revision r3 moved the Unit file (Services) and the Remote (Tap Info) to the right
column: they are paths and URLs, the long values that the 140x40 split squeezed.
Paths have no spaces, so a long path breaks mid-word at the column edge; that is
the wrap, not a cut.

Expectation change: `services_view_test.go` at width 100 expects the log path to
wrap (`/home/linuxbrew/.linuxbrew/var/log/moshi-` then `hook.log`). The test used to
look for the whole path on one line. It now checks both halves.

Tests: `presentation/info_test.go` (split at the constant, one column one cell below,
wrap, order), `info_split_test.go` (Tap Info split at 82 and one column at 81; the
140x40 Formulae Info splits, `TestInfoSplitsAt140x40`). The Tap Info test now checks
the counts sit past the left column, because the Remote is on the right above them.

### Changed by r3 (Q-v2-09-a): the 140x40 screens

At 140x40 the Info views now split. The screens that changed because of the split
(the others at 140x40 are unchanged by it): 02 formulae (Info is the default tab),
03 formulae info, 06 taps (Tap Info), 17 cask info, 19 services info, and the
overlays that sit over those Info views: 07 confirm modal, 10 toast, 12 confirm
upgrade, 13 toast stack, 14 input modal, 20 services stop confirm. 80x24 is
byte-identical to before this revision (below the threshold). 200x60 was already
split; its Tap Info, Services Info and the overlays over them (06, 14, 19, 20) change
only because the Remote and the Unit file moved to the right column.


### Log padding (review r3)

The Log box has the same one column of padding left and right as the main box.
`renderCommandLogBox` gives the log text a width of `width-2-2*mainPadX` and pads it,
so `No commands executed yet` and the command lines start at column 2 of the box.
Test: `TestLogBoxHasInnerPadding` (empty and with lines, at 140x40 and 80x24). Every
frame with a Log box changes in its Log rows; the rest of each frame is unchanged.

## AC-04 — the Log box gives the space back when idle

`render.go` `logHeight` (outer rows, border included), read by `screenLayout`, so
render, the C toggle and mouse hit-testing agree:

| State | Log box rows |
|---|---|
| idle and empty | 3 (border + `No commands executed yet`) |
| idle with lines | min(bodyH/5, 7): at most 5 body rows |
| a task runs (`tasks.IsRunning`) | bodyH/5, today's size |
| collapsed (C) | 0 |

The height changes at once, with no animation. Main and Log always sum to the body
height (`main_padding_test.go` and `log_height_test.go` check this).

"A task runs" is the task manager's `IsRunning`, not the operation overlay. A
finished operation left on screen does not enlarge the Log.

Expectation changes:

- `review_r3_test.go` (`TestCommandLogToggleKeepsCursorVisible`): the log is
  expanded but idle and empty, so the main content is 15 rows, not 14, and the
  clamped offset is 15, not 16. The comments say why.
- `scroll_test.go`: the fixture grows from 60 to 90 caveat lines. The main area is
  now taller, and 60 lines no longer scroll.

Toasts: a toast sits over the bottom-right of the Log box, as before. The 3-row
box is covered by the toast's own bottom border at 80x24. The toast is unchanged.

Tests: `log_height_test.go` (every state at 80x24, 140x40 and 200x60; main + Log
fill the body) and `TestLogBoxHeightChangesAtOnce`.

## AC-05 — evidence

- Harness: `frame_dump_test.go.txt` here, copied unchanged from v2-08 (the same 20
  screens). Run as `DUMP_DIR=<after> CGO_ENABLED=1 go test ./internal/gui/dump/`
  with the file copied to `internal/gui/dump/dump_test.go`, then removed.
- `before/` = `work/verification/v2-08-preload/after/` (both sizes, copied).
- `after/140x40/` (`.txt` and `.ansi`) and `after/80x24/` (`.txt`): the harness as
  is. Its run took 48 s, close to v2-08's 28 s. Brew reads were not slow. The
  harness reports no `not loaded` box beyond Search. Taps shows its trusted data.
  No brew write ran: the harness only opens and cancels confirms.
- `after/200x60/` (`.txt`): an extra set, the same harness with its size list
  changed to `{{200, 60}}`. It shows the AC-03 split, which 140x40 does not reach.
- Sidebar check: in every before/after pair at both sizes the sidebar columns are
  identical, so no brew data changed between the two captures. Every difference
  is in the main area and the Log box.

### Per screen

Frames 08 and 16 (help) are identical: the help overlay covers the whole main area.
For the other screens, the 140x40 and 80x24 frames differ by the one-column padding
and the Log box (now padded: every screen with a Log box changes in its Log rows,
see "Log padding"), plus the table or the Info change named below.

| Screen | 140x40 | 80x24 | Why |
|---|---|---|---|
| 01 status | padded, Log 3 rows | padded, Log 3 rows | AC-01, AC-04 |
| 02 formulae | padded, Log 3 rows | padded, Log 3 rows | AC-01, AC-04 |
| 03 formulae info | padded, split: Description on the right, Log 3 rows | same, single column | AC-01, AC-03, AC-04 |
| 04 formulae deps | Deps table: Name/Version/Description, groups as heading rows | Deps table: Description kept and cut with `…`, Version kept | AC-02, AC-01 |
| 05 outdated | padded, Log 3 rows | padded, Log 3 rows | AC-01, AC-04 |
| 06 taps (Tap Info) | padded, split: counts on the right, Log 3 rows | same, single column | AC-01, AC-03, AC-04 |
| 07 confirm modal | backdrop padded, Log 3 rows | same | AC-01, AC-04 |
| 08 help | unchanged | unchanged | help covers the main area |
| 09 search | padded, Log 3 rows | padded, Log 3 rows | AC-01, AC-04 |
| 10 toast | padded; toast now covers the 3-row Log | same | AC-01, AC-04 |
| 11 deps (bash) | Deps table | Deps table, Description kept and cut with `…` | AC-02 |
| 12 confirm upgrade | padded, Log 3 rows | same | AC-01, AC-04 |
| 13 toast stack | padded; toasts over the Log | same | AC-01, AC-04 |
| 14 input modal | padded, Log 3 rows | same | AC-01, AC-04 |
| 15 menu brewfile | padded, Log 3 rows | padded | AC-01, AC-04 |
| 16 help scrolled | unchanged | unchanged | help covers the main area |
| 17 cask info | padded, split: Description on the right, Log 3 rows | same, single column | AC-01, AC-03, AC-04 |
| 18 doctor | padded, Log 3 rows | same | AC-01, AC-04 |
| 19 services info | padded, split: paths on the right, Log 3 rows | same, single column | AC-01, AC-03, AC-04 |
| 20 services stop confirm | padded, Log 3 rows | same | AC-01, AC-04 |

The 200x60 set shows the AC-03 split on the Info screens: 03 (formula Description
right), 17 (cask Description right), 06 (Formulae count right of the Tap row) and
19 (Command, Working dir and Log right, wrapped; the Unit file in full on the left).

### Not in the frames

- Taps **Formulae** and Formulae **Files** tables are not in the harness. They are
  covered by `table_test.go` and `depviews_test.go` only.
- Cut behaviour at 80x24 for Services Info is unchanged from v2-08 (single column).

## Questions

- **Q-v2-09-a** (`work/questions.md`): resolved by the orchestrator's ruling in r3:
  the split threshold is box inner width ≥ 84, so 140x40 splits. See AC-03 above.

## What the next session must know

- The Log box is padded one column either side, like the main box (`mainPadX`).
- The split threshold is one constant, `splitWidth` in `presentation/detail.go`
  (82 content cells = box inner 84).
- The Deps tab no longer uses `tabContent`. Anything that expects a stored Deps
  string needs the rendered view instead.
