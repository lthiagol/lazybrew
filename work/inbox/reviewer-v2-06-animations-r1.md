verdict: changes-required

# Review r1 — slice v2-06-animations (wip 314dd3c..0656048)

**V1 gate (CGO_ENABLED=1):** build, lint, `go test ./... -race -count=1` (11 pkgs ok), cover-check (brew 76.5, presentation 98.2, gui 84.2, modal 90.1): all pass.
**V2/V4 frames:** harness re-run against real brew (read-only keys, same as the v2-05s harness plus `LAZYBREW_ANIMATIONS=0`): `diff -r` vs `v2-05s-services/after/` and vs `v2-06-animations/after/` both empty. 40 `.txt` exactly W×H (rune widths). Motion dump re-run: `diff -r` vs `motion/` empty; 15 frames exactly 140×40. moshi-hook MainPID=800, NRestarts=0, ActiveEnter 2026-11-18 11:37:48, unchanged. `internal/gui/dump` removed; worktree clean.
**Race / clone design:** `Update` clones the Set before any write and never writes a Set it has returned; the `*p = nm` write-back runs before bubbletea starts the returned cmd (e.g. `executeSearch`'s closure reading `m.client`), so no concurrent write. A throwaway teatest session with animations **on** (tabs, focus, modal, search, resize, refresh) under `-race -count=3`: clean. No new goroutine or `program.Send`; the only `anim.Tick()` calls (motion.go:48, :67) both check `!Settled()`.

| AC | verdict | evidence |
|---|---|---|
| AC-01 engine/switch | partly | harmonica spring, ω=30 critically damped; snap at ½ cell/<5 cells/s; idle/off/stray-frame tests; config default on, env 0/1, file. Test helpers: see F-r1-01 |
| AC-02 tab underline | ok | slides between tabs (motion-tab-01..05, ~2 cells/frame); at rest equals the resting bar; keys mid-slide. Click path is the same state-driven code (no own test; note only) |
| AC-03 sidebar | ok | `roundHeights` sum exact; tests at 140×40 and 80×24; selected row visible; resize snaps |
| AC-04 overlays/toasts | ok | 3-row drop, 3-step fade, instant close, keys on first frame, toast slide + re-stack, dropped toasts' values removed. Visual note F-r1-04 |
| AC-05 progress | ok | bubbles spring via `WithSpringOptions`; `ViewAs` when off; count exact every frame; new batch starts empty |
| AC-06 evidence | ok | after/ = v2-05s; motion/ 15 frames; result.md; STATE row; DESIGN row (`gui.icons` row already existed) |
| V8 invariants | ok | only harmonica promoted (go.sum unchanged); no test deleted/weakened (only main_test.go edits add the env); no key/content change |

Motion as seen in `motion/*.ansi`: subtle and short; a 35-cell move settles ≤ 300 ms, the 3-row drop and the fade in ~100–150 ms. Nothing jarring except F-r1-04.

## Findings

**F-r1-01 (major) — internal/gui/testutil/program.go:14,23; internal/app/app_test.go; internal/gui/flows** — AC-01.4 says every test helper runs with animations off and a test asserts the helpers turn it off. `testutil.NewTestModel`/`NewTestModelWithRunner` use `config.Default()` (animations **on**) and do nothing themselves: they are off only because the flows `TestMain` sets the env. Nothing in `flows` asserts it (`TestTestsRunWithAnimationsOff` is in package gui only and checks `animationsOn(config.Default())`, not the helpers). `internal/app` tests build a gui Model through `app.New` → `config.Load` → `gui.New` with **no** `TestMain`, so that Model animates (and reads the developer's real `~/.config/lazybrew/config.yml`). Fix must achieve: the testutil helpers produce a Model with motion off by themselves (not by the caller's env), `internal/app` tests run with animations off, and a test in `flows` (and one covering the testutil helpers) fails if a helper-built Model would animate.

**F-r1-02 (minor, efficiency) — internal/gui/gui.go:250-266, internal/gui/motion.go:178** — the clone + `motionTargets` run on **every** message, also with animations off. Measured at 140×40 with a mock runner (throwaway benchmark, removed): an unhandled message costs 187 allocs / 28 µs / 48 KB with animations off and 178 allocs on, vs **2 allocs / 2.4 µs / 19 KB** at 314dd3c; a frame tick costs 292 allocs. pprof: 84 % is `motionTargets` — `activeTabSpan` renders every tab, and `restingContentHeights` → `screenLayout` → `renderTabBar` renders the tab bar again; `sidebarName` strconv 15 %; `Set.Clone` 11 %. Every mouse-motion event (CellMotion) and every streamed task line pays it. Fix must achieve: no motion work at all when animations are off (the reads already fall back to the resting values), and with them on, targets recomputed only when their inputs change (panel, tab, size, modal, toast stack, op share) or computed without re-rendering the tab bar; resting frames unchanged.

**F-r1-03 (minor, process) — 7c79a51, 0ebc253** — the brief says every commit passes the tests and one commit per AC. Reproduced in a worktree: at 7c79a51 `go test ./internal/gui` fails `TestMovingValueTicksUntilItSettles` (`settled at 0; want exactly 10`), fixed by the extra AC-03 commit 0ebc253. The runner disclosed it. History can't be rewritten (brief forbids it), so the fix must achieve: the deviation is recorded in `work/verification/v2-06-animations/result.md` (commit range, failing test, why), not only in the receipt.

**F-r1-04 (minor, visual) — internal/gui/render.go:199-208** — at fade 0 the whole backdrop is redrawn in the **text** role: `motion-modal-01.ansi` paints 51 runs `38;2;205;214;244` (brighter than the frame's own borders/subtle text) before dimming to `139;145;169` by frame 05. For one or two frames the screen behind a new modal brightens, then dims: a small flash. The fade also settles by the ½-*cell* rule on a 0..1 value, so it snaps from ≈0.8 to 1. Fix must achieve: the first fade frame is no brighter than the frame it replaces (e.g. start the blend at its first intermediate step, or from the subtle role), and the fade's settle threshold fits a 0..1 value; resting frame unchanged.

## Notes (not findings)

- A modal closed mid-entrance keeps ticking its hidden drop/fade values until they settle (≤ ~200 ms), as toast values are removed but modal values are not. Harmless.
- `TestIdleScreenSchedulesNoTick` uses a model with no window size; the motion dump's settle loop (runs until a frame returns no cmd, with real data) covers the ready case.
- `animationsOn` ignores an unknown env value while `config.Load` rejects it; production goes through `Load` first, so it is consistent in practice.
