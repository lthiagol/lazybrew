# v2-03b-sidebar-lists — frame result

`before/` = copy of `work/verification/v2-03-look-foundation/after/` (the v2-03
frames, real Homebrew 7.0.15, Unicode icon set).
`after/` = this slice's frames, captured at the AC-04 HEAD (`33f6348`) with real
Homebrew 7.0.15 and the Unicode icon set (`LAZYBREW_ICONS=unicode`, set in the
harness). `after/140x40/*.ansi` is the colour evidence; `.ansi` is kept at
140×40 only, as in v2-03.

The harness is `frame_dump_test.go.txt`, a copy of v2-03's with one deviation:
the List tab is gone, so `03-formulae-info` and `11-deps-bash` each lose the
`]` that used to step List → Info. Without that the names would lie (03 would
show Deps, 11 would show Used By). Every other key is unchanged. The harness
issues only read commands against Homebrew (list, info, outdated, taps,
services, config, doctor, uses); no write was run.

**Why almost every row moved.** `gui.sidebar_width` is now 35%, not 30%: the
sidebar is 49 cells wide at 140 columns (was 42) and 28 at 80 (was 24), and the
main column moves right with it. A raw line diff therefore changes nearly every
row of every frame. The evidence is the screen-by-screen reading below, not the
row count.

| screen | 140x40 | 80x24 |
|---|---|---|
| 01-status | The Status sidebar is now the summary: `Homebrew`, `Prefix`, `Formulae 125`, `Casks 2`, `Taps 3`, `Services 3`. Outdated is left out because its count is not known (AC-03.4). The header shows `125 formulae  2 casks` on first launch, before the Formulae or Casks panels open (AC-03.4). The Dashboard tab is the same text, drawn without the `▸` cursor, because the main area no longer draws a list (AC-03.1). Footer unchanged. | The active Status box has two rows (`Homebrew`, `Prefix`), as in v2-03: a 24-row terminal gives the active box two rows (Q-v2-small). The counts are in the header, `125 formulae  2 casks`. |
| 02-formulae | The List table is gone; the main tab is Info, the default (AC-02). Info shows `Tap: homebrew/core`, the column the table used to carry. The sidebar rows are built from the typed data: a status glyph, the name, and the version right-aligned (AC-01). | Same shape. Info wraps at 48 cells. |
| 03-formulae-info | Breadcrumb `Formulae › Info` for `bash`. The Info text is unchanged apart from the width shift. | Same. |
| 04-formulae-deps | Breadcrumb `Formulae › Deps`. The dependency list is unchanged apart from the width shift. | Same. |
| 05-outdated | Breadcrumb `Outdated › Info` with `No selection`, as in v2-03: the machine has nothing outdated, so the list is empty and the Info tab has no item. The Outdated box shows `0`, the loaded empty count, as before. | Same. |
| 06-taps | Breadcrumb `Taps › Tap Info`. The main area shows the summary line, `3 taps · 0 official · 2 trusted · 1 untrusted`, and the selected tap's details. The table-like list is no longer drawn there (AC-03.1). The separator changes from `|` to `·` as the brief's table has it. The Taps sidebar rows carry a trust glyph and a status word (AC-01). | Same. |
| 07-confirm-modal | The uninstall dialog for `bash` is the same text, and the footer shows the modal's keys (`y yes  n no  esc cancel`). The dialog sits over the moved layout. | Same. |
| 08-help | No change: 0 changed rows at both sizes. The help screen has no List mention and is full-screen text. | No change: 0 changed rows. |
| 09-search | The search box and the result info are unchanged. Only the sidebar width moved the main column. | Same. |
| 10-toast | The toast is on the last body row, the Log box's bottom border at this size, as in v2-03 (the toast restyle is v2-04). | Same. |
| 11-deps-bash | Breadcrumb `Formulae › Deps` for `bash`. The harness key sequence is one `]` shorter (see above). | Same. |

## What the frames do not show

The harness does not visit these, so they are covered by unit tests instead
(named in the receipt): the Services Status tab, the Outdated Versions tab with
a batch selected, the Taps Trust tab, the mouse wheel over the main panel, and
the header counts with an unknown outdated count. The Outdated Versions batch
summary is `TestOutdatedVersionsShowsBatchSummary`; the Services detail is
`TestMainAreaShowsDetailsNotTheList`.

## Defects found and fixed while reading the frames

None in the product. The one thing that looked wrong, the 11 screens' names
after the List tab went, was the harness key sequence, not the GUI, and is fixed
in the harness copy above.
