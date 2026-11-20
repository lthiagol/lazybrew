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
| 07-confirm-modal | changed | changed | The uninstall dialog has the new face: thick border in the danger accent (Error red), the chip ` UNINSTALL BASH ` on top, the question in bold, the subtle detail `removes bash; 1 package depends on it`, the dependents list kept as body text, and the key line `[y]es   [n]o   esc cancel` with the Yes/No selection in reverse video. The old outer border wrap is gone. The base frame is muted in `.ansi` (53 muted runs; `before/` has none). |
| 08-help | identical | identical | none (help has no backdrop) |
| 09-search | identical | identical | none |
| 10-toast | changed | changed | The single text toast became a bordered box in the success role with its icon: ` ✓ Install bash completed `. It sits bottom-right on the last body row. At 140x40 its bottom border shares the row of the Log box's bottom border, which the toast's own border covers from its left edge; the Log border still runs left of it. |
| 11-deps-bash | identical | identical | none |
| 12-confirm-upgrade | new | new | `u` on Formulae opens `CONFIRM UPGRADE`, warning-coloured, question `Upgrade bash?`, key line. The base is muted in `.ansi` (48 muted runs). The `Install bash completed` toast from 10 is still up, because the harness does not drain ticks (as in v2-03b). |
| 13-toast-stack | new | new | Three toasts stacked bottom-right, oldest first, newest lowest: success `✓ Install bash completed`, error `✗ Upgrade bash: brew upgrade failed`, warning `⚠ Cannot untap official taps`. The warning needs an official tap, and this machine has none, so the harness injects one Taps entry (`homebrew/cask-versions`) before pressing `x`. Nothing runs: the untap is refused. At 80x24 the stack grows up over the Log box and the main panel and stays off the header and footer. |
| 14-input-modal | new | new | `a` on Taps opens the tap-add input as a form frame: thick border in Accent, the chip ` Tap repository (user/repo): `, the input, and the key line `enter submit   esc cancel`. No tap is added. |

## Checks

- Screens 01–06, 08, 09 and 11 match `before/` exactly, in `.txt` and `.ansi` at 140x40 (`cmp`).
- The muted runs in `.ansi`: 53 in 07, 48 in 12, 0 in every `before/` modal frame. Only the base is muted; the footer row is not.
- The footer in 07, 12 and 14 shows the modal's keys (`y yes  n no  esc cancel`, `enter submit  esc cancel`).
- `TestEveryFrameIsTerminalSize` and the modal and toast float tests pass at 140×40, 80×24 and 200×60.
