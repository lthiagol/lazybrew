# runner: cr4-gui-dedup

Commits, AC order: AC-01 1bf20ae, AC-02 2333f2c, AC-03 e7f87a8, AC-04 34b60d9,
AC-05 2248ac7, AC-06 20f827f, AC-07 fa00dce, AC-08 1dd15c8.
wc -l before → after: commands.go 2124 → 712; gui.go 1499 → 675.

## AC evidence
- AC-01 [R1]: 17 enqueue tails and startBrewUpdate use enqueue / tryEnqueue.
  Queue-full and queued-behind toast tests (review_r1/r2/r3) pass unchanged.
- AC-02 [R2, R4]: task_build.go; task_build_test.go, 10 tests, -race ok.
  linesTask uses runVulns's buffer of 1 with every send ctx-guarded. runMissing
  and runDoctor move from unbuffered to buffered-1: same lines, same order.
- AC-03 [R5, R6, Q2]: packageWriter chosen once; guards use selectedName;
  selectedOutdatedFormula deleted; generic selectNamed. Q2: 19 closures capture
  client (17 Run closures, doMutation, batch upgrade).
- AC-04 [R7, R8, R10, R11]: presentation.WithIcon / YesNo; textutil.Clip / Fit
  with textutil_test.go (ANSI, wide runes, widths 0 and 1); tabWidth, l.sideW,
  leftBlock. New test: active tab width equals tabWidth.
- AC-05 [R9, R12]: keybindings.go holds each key once; footer, help and Info key
  line are built from it; modals expose Hints(). keymap_test.go: 3 tests.
  Byte check: 56 frames (panels, taps and services states, help at 5 widths and
  scroll positions, modals) rendered before and after; cmp identical. Mutation:
  dropping "pgdown" fails TestKeymapMatchesDispatch.
- AC-06 [R13, Q3, Q4, Q5]: panelData.toggleLeaves; one toast per mutation result;
  pendingLabel replaces the label map. Q4 is the one behaviour change: a cask
  uninstall confirm makes no uses call. Mutation: forcing the call fails with
  `[uses --installed firefox]`.
- AC-07 [Q6]: tabs.go holds every tab constant; caskInfo/Deps/CaveatsTab.
- AC-08 [Q1]: move by top-level declaration. gui.go → model.go, keys.go, land.go,
  panels.go, tasks_diagnostics.go. commands.go → tasks_packages.go, tasks_taps.go,
  tasks_services.go, tasks_diagnostics.go. Line diff of old and new: only headers,
  imports, handleKey's header and `var cmd`, the call site, one key assertion,
  and the 7-line modal pass-through (now in update and handleKey).

## Gate
- make ci: PASS (gofmt, vet, go test ./... -race -count=1, cover floors, build;
  gui 89.7% >= 83%).
- GOOS=darwin go build ./...: PASS.
- deadcode and staticcheck: NOT RUN (not installed; fetch denied; not attempted).
- Production callers: each added func has one; dispatchKeys is test-only.

## Test edits (renames, no assertion weakened)
- withIcon → presentation.WithIcon (icons_test, look_test).
- help_scroll_test and scroll_test read helpKeyText() / help.
- operation_dismiss_test: positional keyHint literal → keyClose.

## Questions and seen, not done
- No questions appended to work/questions.md.
- extractPackageName is test-only, as before this slice (cr5).
- render.go Artifacts loop still uses += (outside commands.go scope).
- panel.go:347 and presentation/detail.go:136 keep raw ansi.Truncate.
- Panel keys stay switches; the test checks msg.String() cases package-wide.
