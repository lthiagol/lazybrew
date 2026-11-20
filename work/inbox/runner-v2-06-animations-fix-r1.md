# fix round r1 — v2-06-animations

Commits, one per finding, in order:
- F-r1-01 `ee7d782` testutil helpers build still models; app tests use a temp HOME
- F-r1-02 `789caf6` motion off does no motion work; no tab-bar render for layout
- F-r1-04 `7a66ae2` modal backdrop fades from the subtle role, never brightens
- F-r1-03 `a000af9` result.md records the AC-03 intermediate red commit

Per finding:
- F-r1-01: `NewModel` = `gui.New(...).WithoutMotion()`; `NewTestModel*` use it.
  `testutil/program_test.go` and `flows/still_test.go` fail if a helper-built
  model animates with `LAZYBREW_ANIMATIONS=1`. `app_test.go` TestMain sets HOME
  to an empty temp dir and animations off; `TestAppModelIsStill`.
- F-r1-02: benchmark `BenchmarkUpdateUnhandled` (`motion_bench_test.go`, not in
  the gate), one unhandled message at 140×40:
  - 314dd3c: 2 allocs, 2.0 µs, 19.5 KB
  - before fix (animations=0): 187 allocs, 28 µs, 48 KB
  - after fix, animations=0: 2 allocs, 2.8 µs, 21.8 KB
  - after fix, animations=1: 20 allocs, 9.7 µs, 44.5 KB
  Off is within 2× of 314dd3c. Tests hold `tabWidth` and `tabBarRows` to the
  render.
- F-r1-04: `backdropColor` runs Subtle → Muted; fade is 0..100 so the settle
  threshold is half a percent. `TestBackdropFadeNeverBrightens` (luminance
  non-increasing) and `TestModalFadeSpringIsMonotone` (steps ≤ 0.25, exact 1).
  Modal frame 01 backdrop: text-role runs 51 → 0 (subtle now). `motion/`
  modal .ansi re-dumped; .txt unchanged.
- F-r1-03: `result.md` records the `7c79a51` failure and the `0ebc253` fix.

Gate (HEAD `7a66ae2`, then docs commit `a000af9`): `make build`, `make lint`,
`CGO_ENABLED=1 go test ./... -race -count=1` all ok, `make cover-check` all floors met.

Resting frames: harness re-run after F-r1-02 and after F-r1-04;
`diff -rq` vs `work/verification/v2-05s-services/after` is empty both times.
Motion dump re-run: only `motion-modal-*.ansi` changed.

Worktree clean. No push. No brew write run.
