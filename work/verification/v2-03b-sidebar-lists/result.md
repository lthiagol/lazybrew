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

## Fix round r1

Re-captured at `752583b` with the same harness, real Homebrew 7.0.15, Unicode
icons. `after/` now holds these frames; `after/140x40/*.ansi` is updated.
Compared with the round-0 frames (`33f6348`-era `after/`), the changed rows are
these and nothing else:

- **Every frame with the Status sidebar (F-r1-02):** the first row reads
  `Homebrew  7.0.15-…` (label in the same ten-column field as `Prefix`,
  `Formulae`, …). One row per frame, at both sizes.
- **06-taps and 09-search, 140x40 and 80x24:** the same Homebrew row. At 140x40
  the footer also ends `q quit  ? more` (06) and `C log  ? more` (09): the
  page-key hint (`PgUp/Dn scroll`, F-orch-03) is the lowest-priority global hint
  and is cut first. Footer rows of the other frames are unchanged, `? help`
  still shows.
- **08-help, 140x40 and 80x24:** the Navigation section gains `PgUp/PgDn  Scroll
  content by a page` and `Home/End, g/G  Top/bottom of content` (F-orch-03). The
  help's columns reflow. At 80x24 the help was already cut at 24 rows (Q-v2-help,
  v2-05); the new rows push the cut lower.
- **Not in the harness:** the batch marks (F-orch-01: no batch was made in these
  screens, so no frame changes), the Trust tab (F-orch-02: the harness does not
  visit Trust), and the page keys' effect on the content (F-orch-03: covered by
  `TestScrollKeysMoveAndClamp`).
- **Frame sizes:** every `.txt` frame is exactly W×H (checked by row count and
  width).

## Fix round r2

Re-captured at `b501d1f` with the same harness, real Homebrew 7.0.15, Unicode
icons. The only frames that changed against round r1 are the help screen
(08), at both sizes and at 140x40 in `.ansi`:

- **08-help, 140x40:** the Navigation row reads `Home/End   Top/bottom (g/G
  outside Search)` on one line. The round-r1 row wrapped mid-phrase in its
  45-cell column (F-r2-01).
- **08-help, 80x24:** the same row, wrapped at the 25-cell column; the help was
  already cut at 24 rows (Q-v2-help).

No other frame changed. The g/G change on Search (F-r2-01) and the plural
trust story (F-r2-02) are not in the harness: the harness does not type into
Search after Esc, and does not open the Trust tab. They are covered by
`TestLetterKeysOnUnfocusedSearchTypeToSearch` and `TestTrustStoryPluralises`.
