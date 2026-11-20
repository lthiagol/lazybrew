# v2-04-overlays-styled — frame result

`before/` is `work/verification/v2-03b-sidebar-lists/after/` (the v2-03b frames).
`after/` is this slice, dumped by `frame_dump_test.go.txt` (the v2-03b harness with
three new screens, 12–14). Both sizes are text (`.txt`); `140x40` also has `.ansi`,
which is where the dim backdrop is visible (`80x24` keeps only `.txt`, as before).

Captured with the real Homebrew on this machine (3 taps, 0 official), the Unicode
icon set, and no brew write command. Nothing was confirmed.

## Per screen

| screen | 140x40 | 80x24 | what changed |
|---|---|---|---|
| 01-status | identical (.txt and .ansi) | identical | none |
| 02-formulae | identical | identical | none |
| 03-formulae-info | identical | identical | none |
| 04-formulae-deps | identical | identical | none |
| 05-outdated | identical | identical | none |
| 06-taps | identical | identical | none |
| 07-confirm-modal | changed | changed | The uninstall dialog has the new face: thick border in the danger accent (Error red), the chip ` UNINSTALL ` (the action; the name is in the question), the question `Uninstall bash? This cannot be undone.` in bold, the subtle consequence `removes bash; 1 package depends on it`, the dependents list as body text, and the key line `[y]es   [n]o   esc cancel` with the Yes/No selection in reverse video. The old outer border wrap is gone. The base frame is muted in `.ansi` (53 muted runs; `before/` has none). |
| 08-help | identical | identical | none (help has no backdrop) |
| 09-search | identical | identical | none |
| 10-toast | changed | changed | The single text toast became a bordered box in the success role with its icon: ` ✓ Install bash completed `. It sits bottom-right on the last body row. At 140x40 its bottom border shares the row of the Log box's bottom border, which the toast's own border covers from its left edge; the Log border still runs left of it. |
| 11-deps-bash | identical | identical | none |
| 12-confirm-upgrade | new | new | `u` on Formulae opens the chip ` UPGRADE ` (warning-coloured), the question `Upgrade bash?`, the consequence `Installs the newest version of bash.`, and the key line. The base is muted in `.ansi` (49 muted runs). The `Install bash completed` toast from 10 is still up, because the harness does not drain ticks (as in v2-03b). |
| 13-toast-stack | new | new | Three toasts stacked bottom-right, oldest first, newest lowest: success `✓ Install bash completed`, error `✗ Upgrade bash: brew upgrade failed`, warning `⚠ Cannot untap official taps`. Only the three boxes cover the frame: every cell beside them is the base frame (fix r1, F-r1-01). The warning needs an official tap, and this machine has none, so the harness injects one Taps entry (`homebrew/cask-versions`) before pressing `x`. Nothing runs: the untap is refused. At 80x24 the stack grows up over the Log box and the main panel and stays off the header and footer. |
| 14-input-modal | new | new | `a` on Taps opens the tap-add input as a form frame: thick border in Accent, the chip ` ADD TAP `, the prompt `Tap repository (user/repo):` as body text, the input, and the key line `enter submit   esc cancel`. No tap is added. |

## Fix round r1

Frames changed by the r1 fixes (the other screens were re-captured and compared
with `cmp`: identical, so they are not listed again):

- **07-confirm-modal** (both sizes; `.ansi` 140x40): chip is now the action
  `UNINSTALL`, not `UNINSTALL BASH` (F-r1-03). The box is at most 20 rows at 80x24
  (F-r1-04).
- **12-confirm-upgrade** (both sizes; `.ansi`): chip `UPGRADE`, not `CONFIRM UPGRADE`
  (F-r1-03); the consequence line `Installs the newest version of bash.` is new
  (F-r1-02). Muted runs in `.ansi`: 49.
- **13-toast-stack** (both sizes; `.ansi`): cells beside the narrower boxes are the
  base frame again. Before the fix a padded layer blanked them (F-r1-01).
- **14-input-modal** (both sizes; `.ansi`): chip `ADD TAP` (the action); the prompt
  is body text, not the chip (F-r1-03, Q-v2-04-wrap).

## Fix round r2

Frames changed by the r2 fix (the other screens were re-captured and compared
with `cmp`: identical):

- **14-input-modal** (both sizes; `.ansi`): the field is a fixed 33 cells (prompt,
  30 of text, cursor), where it drew 43 and sat off-centre. The chip, prompt and
  key line are centred again (F-r2-01).
- **07-confirm-modal** and **12-confirm-upgrade** (both sizes; `.ansi`): each row
  is now centred on its own, rather than the block being centred. A row with an odd
  gap sits one cell further left or right than before; the content is unchanged.
  Muted runs in `.ansi`: 53 and 49, as before.

## Checks

- Screens 01–06, 08, 09 and 11 match `before/` exactly, in `.txt` and `.ansi` at 140x40 (`cmp`). 10 differs only by the toast box, as described.
- The muted runs in `.ansi`: 53 in 07, 49 in 12, 50 in 14 (the r2 fix leaves the counts as they were), 0 in 13 (no modal open). Only the base is muted; the box and the footer are not.
- The footer in 07, 12 and 14 shows the modal's keys (`y yes  n no  esc cancel`, `enter submit  esc cancel`).
- `TestEveryFrameIsTerminalSize` and the modal, toast and backdrop tests pass at 140×40, 80×24 and 200×60.
