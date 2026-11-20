# runner: cr4-gui-dedup

Commits: AC-01 1bf20ae, AC-02 2333f2c, AC-03 e7f87a8, AC-04 34b60d9,
AC-05 2248ac7, AC-06 20f827f, AC-07 fa00dce, AC-08 1dd15c8.
wc -l: commands.go 2124 → 712; gui.go 1499 → 675.

## AC evidence
- AC-01 [R1]: 17 tails plus startBrewUpdate use enqueue / tryEnqueue. Queue-full
  and queued-behind toast tests (review_r1/r2/r3) pass unchanged.
- AC-02 [R2, R4]: task_build.go; task_build_test.go, 10 tests, -race ok.
  linesTask: runVulns's buffer of 1, every send ctx-guarded. runMissing and
  runDoctor move to buffered-1: same lines, same order.
- AC-03 [R5, R6, Q2]: packageWriter; selectedName guards; selectedOutdatedFormula
  gone; generic selectNamed. Q2: 19 closures capture client.
- AC-04 [R7, R8, R10, R11]: presentation.WithIcon / YesNo; textutil.Clip / Fit
  with tests (ANSI, wide, widths 0 and 1); tabWidth, l.sideW, leftBlock.
- AC-05 [R9, R12]: keybindings.go holds each key once; footer, help and Info key
  line are built from it; modals expose Hints(). keymap_test.go: 3 tests.
  Byte check: 56 frames (panels, taps and services states, help at 5 widths,
  scroll, modals) before and after: cmp identical. Mutation: dropping "pgdown"
  fails TestKeymapMatchesDispatch.
- AC-06 [R13, Q3, Q4, Q5]: panelData.toggleLeaves; one toast per result;
  pendingLabel. Q4 is the one behaviour change: a cask uninstall confirm makes
  no uses call. Mutation: forcing it fails with `[uses --installed firefox]`.
- AC-07 [Q6]: tabs.go holds every tab constant; caskInfo/Deps/CaveatsTab.
- AC-08 [Q1]: move by declaration. gui.go → model, keys, land, panels,
  tasks_diagnostics. commands.go → tasks_packages, tasks_taps, tasks_services,
  tasks_diagnostics. Old/new line diff: headers, imports, handleKey's header and
  `var cmd`, the call site, one key assertion, the 7-line modal pass-through.

## Gate
- make ci: PASS (gofmt, vet, race tests, cover floors, build; gui 89.7% >= 83%).
- GOOS=darwin go build ./...: PASS.
- deadcode, staticcheck: NOT RUN (not installed; fetch denied; not attempted).
- Production callers: each added func has one; dispatchKeys is test-only.

## Test edits (renames only)
- withIcon → presentation.WithIcon (icons_test, look_test); help and scroll
  tests read helpKeyText() / help; operation_dismiss uses keyClose.

## Seen, not done
- No questions filed.
- extractPackageName is test-only, as before (cr5).
- render.go Artifacts loop still uses +=; panel.go:347 and
  presentation/detail.go:136 keep raw ansi.Truncate.
- Panel keys stay switches; the test checks msg.String() cases package-wide.
