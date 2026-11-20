# v2-06-animations — frame result

Two sets of frames, both captured with the real Homebrew on this machine, read
only (`brew services info --all --json` and the list and info reads the harness
already makes). No brew write ran: the harness sends `S` to open the stop
confirm for moshi-hook and `esc` to cancel it, as in v2-05s.

## Resting frames (AC-06.1)

`after/` is `frame_dump_test.go.txt` run with `LAZYBREW_ANIMATIONS=0` (set in
the harness). It is the v2-05s harness with one added line.

`diff -rq work/verification/v2-05s-services/after work/verification/v2-06-animations/after`
prints nothing: all 20 frames at 140×40 (`.txt` and `.ansi`) and all 20 at 80×24
(`.txt`) are byte-identical to the v2-05s frames. Motion does not change any
resting frame.

## Motion frames (AC-06.2)

`motion/` is `motion_dump_test.go.txt` run with `LAZYBREW_ANIMATIONS=1`. The
frame ticks are delivered by hand at 60 Hz (16.7 ms apart) through `Update`,
and each returned command is discarded, so the frames are the spring's own
positions, not wall-clock samples. Every frame is checked to be exactly 140×40.

| motion | frames | what moves |
|---|---|---|
| `motion-tab` | 01–05 | the tab underline slides from *Tap Info* to *Trust*; the bar's `━` run moves right, between the two tabs, and settles on *Trust*. |
| `motion-focus` | 01–05 | Taps to Formulae: the Formulae box's bottom border moves from row 12 to row 16 by frame 5, and the Taps box's top border from row 22 to row 26. |
| `motion-modal` | 01–05 | the *Add tap* prompt starts three rows above its place (row 16) and springs down (row 18 by frame 5); the backdrop fades in over the same frames. It rests on row 19, as the harness's frame 14 does. |

## Checks

- Each motion run is driven until a frame returns no command; the settled
  frame equals the frame with animations off (unit tests:
  `TestModalSlidesDownIntoItsPlace`, `TestToastSlidesInFromTheRight`,
  `TestExistingToastsSpringWhenOneArrives`, `TestTabUnderlineRestsOnTheActiveTab`,
  `TestSlidingTabBarAtRestIsTheRestingBar`, `TestSidebarBoxesSpringWithFocus`).
- The sidebar boxes sum to the sidebar's height on every frame (`boxesFill`),
  at 140×40 and 80×24.
- A settling move of about 35 cells takes ≤ 300 ms at 60 Hz
  (`TestSettlesWithinTheShortMotionBudget`).
- No tick is scheduled when nothing moves (`TestIdleScreenSchedulesNoTick`,
  `TestAnimationsOffNeverTicks`, `TestStrayFrameAfterSettleSchedulesNothing`).

## Not verified here

- The motion as the eye sees it in a real terminal (frame rate, colour fade):
  the human smoke gate, `SMOKE-CHECKLIST.md`.
- Keys and mouse in a real terminal during a motion: covered by unit tests
  (`TestKeysWorkWhileTheUnderlineSlides`, `TestModalTakesKeysFromItsFirstFrame`),
  not by a live session.
- A batch progress bar on a real upgrade: covered by unit tests only; no brew
  write was run, so no batch ran here.
