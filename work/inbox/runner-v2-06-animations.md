# runner receipt — v2-06-animations

## Commits, one per AC, in order

- AC-01 `46736ea` anim engine; `gui.animations` + `LAZYBREW_ANIMATIONS`; docs rows.
- AC-02 `4f54fc0` tab underline slides on `[`/`]` and tab click; snaps on panel switch.
- AC-03 `7c79a51` sidebar boxes spring with focus; sum always exact.
  Follow-up `0ebc253`: `7c79a51` alone fails `TestMovingValueTicksUntilItSettles`
  (the test read the pre-update model); the fix is in `0ebc253`. Both are AC-03.
- AC-04 `d5fb165` modal drop + backdrop fade; toast slide-in and re-stack.
- AC-05 `16358cf` batch progress bar springs (bubbles progress, own frames).
- AC-06 `638fb61` evidence (`work/verification/v2-06-animations/`) + STATE row.

## Evidence per AC

- AC-01: anim tests (converge, snap exact, retarget, ≤300 ms for 35 cells,
  no overshoot, clone); idle screen and animations-off never tick; config tests
  (default on, file, env 0/1, other values rejected).
- AC-02: underline rests on the active tab; mid-slide it is strictly between
  the two tabs; keys work mid-slide; resting renderer equals the sliding one at
  the same geometry; frames exact at 140×40 and 80×24.
- AC-03: at each frame the boxes sum to the sidebar height (140×40, 80×24);
  mid-flight boxes lie between old and new; selected row stays in the window;
  resize snaps.
- AC-04: modal starts 3 rows up, settles to the motion-off frame exactly; close
  is instant; keys reach the modal on its first frame; toasts slide in and the
  rest spring; dropped toasts' values are removed.
- AC-05: bar springs to 50 %, count `5/10` on every frame, settles to the
  motion-off bar; new batch starts empty; animations off schedules no frame.
- AC-06.1: harness copy with `LAZYBREW_ANIMATIONS=0`: `diff -rq` against
  `work/verification/v2-05s-services/after` prints nothing (40 + 20 frames).
- AC-06.2: `motion/` holds 5 frames each for tab switch, focus change, modal
  open at 140×40, ticks delivered by hand at 60 Hz, each frame exactly 140×40.

## Gate (HEAD `638fb61`)

- `make build` ok; `make lint` (go vet) ok.
- `CGO_ENABLED=1 go test ./... -race -count=1`: every package ok.
- `make cover-check`: gui 84.2 ≥ 36, modal 90.1 ≥ 41, presentation 98.2 ≥ 91,
  brew 76.5 ≥ 62. All floors met.

## Decisions the next session should know

- Tests run with animations off through `LAZYBREW_ANIMATIONS=0` set in the
  gui and flows `TestMain`. `config.Default()` stays on, since production
  defaults to on.
- `Model.Update` clones the animated values each update, so a model kept from
  before an update is not moved by it.
- Spring: critically damped, 30 rad/s; a value snaps within ½ cell at ≤ 5 cells/s.
- A modal that replaces an open one slides in again. Closing is instant.
- STATE row uses `—` for commits, as the brief gives it; commit ids are above.

## Not verified here

- Motion as seen in a real terminal: human smoke gate.
- A live batch upgrade bar: unit tests only; no brew write ran.
- Safety: no brew write ran. `moshi-hook` is still pid 800, `started`.
