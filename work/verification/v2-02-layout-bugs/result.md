# v2-02-layout-bugs — frame diff against the v2-01 baseline

The harness is `frame_dump_test.go.txt` (the v2-01 harness with two new screens,
`10-toast` and `11-deps-bash`). For the run it was copied to
`internal/gui/dump/dump_test.go` and removed afterwards; it is not committed as a
test. The frames are in `after/<size>/`, ANSI-stripped, from the real Homebrew on
this machine (Homebrew 7.0.15).

```
$ cp work/verification/v2-02-layout-bugs/frame_dump_test.go.txt internal/gui/dump/dump_test.go
$ DUMP_DIR=$SCR/frames-v2-02 CGO_ENABLED=1 go test ./internal/gui/dump/ -count=1
ok  	github.com/lthiagol/lazybrew/internal/gui/dump	7.783s
$ rm -rf internal/gui/dump
```

## Frame size (AC-01)

All 22 frames (11 screens × 140×40 and 80×24) are exactly the terminal size:
40 rows × 140 cells, and 24 rows × 80 cells. The v2-01 baseline was 39 rows × 138
used cells at 140×40 and 31 rows at 80×24, so the 80×24 frames overflowed the
terminal.

## Bugs found by this diff and fixed in this slice

Two changes were not in B1–B7 as written. Both were caught by reading the diff,
and each is fixed with a test. They are part of B1 (every frame fits the terminal,
the main panel fills its height):

- **Main list capped by the sidebar.** `renderList` limited the main list to the
  sidebar box's row count (`visibleCount`). That count is 1 at 80×24 with the
  fitted layout, so the Status dashboard showed one row where the baseline showed
  nine. The main list now uses its own height. At 140×40 this also shows the
  Services and Doctor rows that the baseline had hidden. Test:
  `TestMainListFillsMainPanel`.
- **Help box overflow.** At 80×24 the help text (38 rows) grew the box past the
  screen and the frame lost the box's bottom border and the close hint. The help
  text is now cut to the box's inner area, so the border and the hint always show.
  Test: `TestHelpKeepsBorderAndHint`.

## Per-screen result

Keys: B1 frame fills the terminal (geometry); B2 modal floats; B3 modal content on
the first frame; B4 toast floats; B5 `not loaded`; B6 Log is a sibling box; B7
real dependencies. "Changed region" is the set of tokens that differ (see the
accounting below).

| # | screen | 140×40 — changed regions are explained by | 80×24 — changed regions are explained by |
|---|---|---|---|
| 01 | status | B1 (rows and columns; one more dashboard row and the Services/Doctor rows now visible), B5 (6× `(empty)` → `not loaded`), B6 (Log below the main box) | B1 (baseline overflowed to 31 rows; cramped sidebar shows titles only for inactive panels), B5, hint wrap cut (Q-v2-hints) |
| 02 | formulae list | B1 (table two columns wider, more rows), B5 (5× `not loaded`), B6 | B1 (baseline table rows were below row 24), B5, hint wrap cut (Q-v2-hints) |
| 03 | formulae info | B1, B5, B6 (info text unchanged) | B1, B5, hint wrap cut (Q-v2-hints) |
| 04 | formulae deps | B7 (`No dependencies` → `Runtime: ncurses, readline` for bash), B1, B5, B6 | B7, B1, B5, hint wrap cut (Q-v2-hints) |
| 05 | outdated | B5 (4× `not loaded`), B1, B6 (`No selection` unchanged) | B1 (baseline overflowed), B5 |
| 06 | taps | B5 (3× `not loaded`), B1, B6 | B1 (cramped sidebar, baseline overflowed), B5 |
| 07 | confirm modal | B2 (modal now floats centred; it covers the Formulae rows it overlaps, `dav1d`…`frond`), B3 (`Loading...` → the modal's title, question and buttons), B5, B1, B6 | B2, B3, B1, B5 |
| 08 | help | B1 only: the 140×40 text is token-identical (0 changed); the box closes and the hint is the last row | B1: box fits 24 rows with border and hint; 54 tokens of help text past row 22 are cut (Q-v2-help) |
| 09 | search | B5 (3× `not loaded`), B1, B6 | B5, B1, hint wrap cut (Q-v2-hints) |

New screens:

- `10-toast`: B4. The toast is on the last body row, flush right; the bottom bar is
  unchanged; the Formulae list is unchanged except B1 and B5.
- `11-deps-bash`: B7. Bash shows `ncurses` and `readline` (the cursor is on bash).

Nothing is left unexplained. The cases that look like losses (help text past row
22 at 80×24, the bottom-bar hint tail) are at the 80×24 limit and are recorded in
`work/questions.md` rather than hidden.

## Token accounting

The changed text was counted per screen (box-drawing characters removed, tokens
split on whitespace). Summary by class:

- `(empty)` → `not loaded`: B5, 5–6 per sidebar screen at 140×40.
- `No dependencies` → `Runtime: ncurses readline`: B7 (screen 04).
- `Loading...` removed and the modal's text added: B3 and B2 (screen 07).
- Formulae rows covered by the modal: B2 (screen 07).
- Extra rows and columns (table, dashboard, sidebar): B1.
- Help: no change at 140×40; 54 tokens past row 22 cut at 80×24 (B1, Q-v2-help).
- Bottom-bar hint tail (`missing`, `upgrade all p pin L leaves F fetch`) cut at
  80×24: the bar is 2 rows now; the baseline wrapped that tail onto a third row
  past row 24, which a 24-row terminal never showed (Q-v2-hints).

## Test count

```
$ grep -rho "^func Test\w*" --include=*_test.go . | wc -l
425
```

Before the slice: 410. Added: `TestEveryFrameIsTerminalSize`, `TestModalFloatsOverFrame`,
`TestMouseIgnoredWhileModalOpen`, `TestModalsShowContentOnFirstFrame`,
`TestModalSurvivesResizeWhileOpen`, `TestToastFloatsAboveBottomBar`,
`TestNeverLoadedPanelSaysNotLoaded`, `TestFailedFirstFetchIsLoadedNotNotLoaded`,
`TestLogIsSiblingBoxBelowMainPanel`, `TestClickOnLogBoxDoesNothing`,
`TestMouseHitsWhatIsDrawn`, `TestBashShowsRealRuntimeDependencies`,
`TestFlexStringsReadsDependencyObjects`, `TestMainListFillsMainPanel`,
`TestHelpKeepsBorderAndHint`. None was deleted or skipped.
