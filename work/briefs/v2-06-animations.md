# Slice: v2-06-animations — spring motion with harmonica, on by default, off for tests

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Last slice of `LB-08`. User rulings (`work/questions.md`, `Q-v2-run`):
animations are **on by default**, `gui.animations: false` turns them off,
**tests always run with them off**. The motion library is Charm's
`github.com/charmbracelet/harmonica` (spring physics; already an indirect
dependency through `bubbles/v2/progress`). Read its README and source in
`$(go env GOMODCACHE)` before writing code.

Motion must be **subtle and short** (settle in ≈ 150–300 ms), and must never
change what the user can do: every key works mid-animation, and the end state
of every animation is exactly the frame the app draws today.

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes the tests; `TestEveryFrameIsTerminalSize` passes at every commit.

### AC-01 — the animation engine and the switch

1. A small `internal/gui/anim` package (or file): named animated values
   (float position/size) driven by `harmonica.Spring`, a `Step(dt)` that
   advances all of them, `Settled()` when every value is within ½ cell of its
   target with ~0 velocity (then snapped exactly to the target).
2. One frame tick: a `tea.Tick` at `harmonica.FPS(60)` that runs **only while
   something is animating** and stops when all values settle (no idle ticks —
   test it). No goroutines; the tick is a `tea.Cmd` returning a message.
3. Config `gui.animations` (bool, default `true`), env
   `LAZYBREW_ANIMATIONS=0|1` overrides it (same precedence as `LAZYBREW_ICONS`).
   With animations off, targets are applied instantly and no tick is ever
   scheduled. README documents both.
4. Every test helper / `config.Default()` used by tests and the frame harness
   runs with animations **off**; a test asserts that the default config in
   production is on and that the test helpers turn it off.
5. Unit tests drive the engine with a fake clock/`Step(dt)`: convergence,
   snap at the end, retargeting mid-flight, no tick after settle.

### AC-02 — the tab underline slides

The active tab's underline (`━`) moves and resizes with a spring from the old
tab to the new one when `[`/`]` or a tab click changes tabs. At rest it is
exactly today's underline. Test: mid-animation frame has the underline
between the two tabs; settled frame equals the animations-off frame.

### AC-03 — sidebar boxes grow and shrink with focus

When focus moves to another panel, the active box's extra height (the
accordion share) springs from the old box to the new one; the sum of box
heights is always exactly the sidebar height (every mid-animation frame is
exact size — test several steps at 80×24 and 140×40). List windows/selection
stay correct while heights move (the selected row stays visible).

### AC-04 — overlays and toasts slide in

1. A modal slides in from a few rows above its resting position and the
   backdrop dim fades in (blend the muted role over 2–4 steps, e.g. with
   lipgloss v2 colour blending); at rest it is exactly today's v2-04 frame.
   Closing is instant (no exit animation — keys must feel immediate).
2. A new toast slides in from the right edge into its stack slot; existing
   toasts spring up/down when one is added or evicted.
3. Input reaches the modal from its first frame (no waiting for the slide).

### AC-05 — progress bars spring

The batch / upgrade-all progress bar (v2-05) animates its fill with the
`bubbles/v2/progress` spring (its built-in harmonica) when the count changes.
With animations off it uses `ViewAs` (static). The `n/total` text is exact at
every frame.

### AC-06 — evidence and bookkeeping

1. Harness copied to `work/verification/v2-06-animations/` — runs with
   animations **off**; all frames must equal `v2-05s-services/after/`
   exactly (`diff -r` empty): motion must not change any resting frame.
2. Mid-animation evidence: a test-side dump of 4–6 intermediate frames for
   one tab switch, one focus change and one modal open at 140×40 into
   `work/verification/v2-06-animations/motion/` (`.txt` + `.ansi`), each
   exact size.
3. `result.md`; `work/STATE.md` row
   `| v2-06-animations | — | implemented | awaiting reviewer |`;
   `DESIGN.md` config table gets `gui.animations` (Wired) and `gui.icons`.

## Invariants

- Resting frames identical to v2-05s (AC-06.1). Every intermediate frame
  exactly the terminal size.
- No tick when nothing animates; no goroutines; no `program.Send`.
- Keys and mouse work at every frame; no input is dropped or delayed by an
  animation; closing things is instant.
- No restyle; no content change; no keybinding change.
- New direct dependency allowed: `github.com/charmbracelet/harmonica`
  (promote from indirect). Nothing else.
- No test deleted, skipped or weakened. Coverage floors unchanged.
- Never run brew write commands on this machine (the Safety section of
  `work/briefs/v2-05s-services.md` still applies).
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.

## Report protocol

`work/inbox/runner-v2-06-animations.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-v2-06-animations.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-v2-06-animations-rN.md` → fix every finding, one commit each
(subject names the id), gate, receipt
`work/inbox/runner-v2-06-animations-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
