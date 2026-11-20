# runner: cr4-gui-dedup

Commits, AC order: AC-01 1bf20ae, AC-02 2333f2c, AC-03 e7f87a8, AC-04 34b60d9,
AC-05 2248ac7, AC-06 20f827f, AC-07 fa00dce, AC-08 1dd15c8.
wc -l before → after: commands.go 2124 → 712; gui.go 1499 → 675.

## AC evidence
- AC-01 [R1]: 17 enqueue tails and startBrewUpdate use `m.enqueue` / `tryEnqueue`
  (startBrewUpdate needs the error to reset isUpdating). Queue-full and
  queued-behind toast tests (review_r1/r2/r3) pass unchanged.
- AC-02 [R2, R4]: task_build.go, errTask/streamTask/linesTask; task_build_test.go
  has 10 tests (success, error, ctx cancel mid-stream, nil line channel), -race ok.
  linesTask settled on runVulns's buffer of 1, every send ctx-guarded. runMissing
  and runDoctor moved from unbuffered to buffered-1: same lines, same order. Safe:
  the manager reads to close after errCh, and a send never blocks past ctx.
- AC-03 [R5, R6, Q2]: packageWriter picks the writer once in doMutation; guards
  use selectedName alone; selectedOutdatedFormula deleted; one generic selectNamed.
  Q2: 19 closures capture client (17 Run closures, doMutation, batch upgrade).
- AC-04 [R7, R8, R10, R11]: presentation.WithIcon / YesNo; textutil.Clip / Fit
  (textutil_test.go: ANSI, wide runes, widths 0 and 1); tabWidth in tab click,
  l.sideW in wheel, leftBlock in menu. New test: active tab width = tabWidth.
- AC-05 [R9, R12]: keybindings.go holds every key once (keyHint: key, helpKey,
  keys, desc, help); footer, help and Info key line list those entries. Modals
  expose Hints() (modal.Hinter). keymap_test.go: 3 tests, dispatch read from the
  msg.String() cases in source. Byte check: before/after rendering of 56 frames
  (every panel, taps x3 states, services x3 x2 tabs, help x5 widths, scroll,
  modals) with cmp: identical. Mutation: dropping "pgdown" fails the dispatch test.
- AC-06 [R13, Q3, Q4, Q5]: panelData.toggleLeaves; one toast per mutation
  result; pendingLabel replaces the label map. Q4 is the one behaviour change: a
  cask uninstall confirm makes no `uses` call (detail "removes firefox").
  Mutation: forcing the call fails with `[uses --installed firefox]`.
- AC-07 [Q6]: tabs.go holds every tab constant; caskInfo/Deps/CaveatsTab replace
  0/1/2; initPanels uses PanelID.String() and panelCount.
- AC-08 [Q1]: move by top-level declaration. Mapping: gui.go → model.go, keys.go,
  land.go, panels.go, tasks_diagnostics.go; commands.go → tasks_packages.go,
  tasks_taps.go, tasks_services.go, tasks_diagnostics.go. Line diff of old and
  new files: only headers and imports, handleKey's header and `var cmd`, the call
  site, one key assertion, and the 7-line modal pass-through (now in update and
  handleKey). Body moved one tab left; color-moved will show those lines.

## Gate
- make ci: PASS (gofmt, vet, go test ./... -race -count=1, cover floors:
  gui 89.7% >= 83%, all OK, build).
- GOOS=darwin go build ./...: PASS.
- deadcode and staticcheck: NOT RUN (not installed; fetch denied; not attempted).
- Production callers: grep for each added func finds one; dispatchKeys and the
  keymap test helpers are test-only (keymap_test.go).

## Test edits (renames, no assertion weakened)
- withIcon → presentation.WithIcon (icons_test.go, look_test.go).
- help_scroll_test.go and scroll_test.go read helpKeyText() / help.
- operation_dismiss_test.go: positional keyHint literal → keyClose.

## Questions and seen, not done
- No questions appended to work/questions.md.
- extractPackageName: test-only, already so before this slice (cr5).
- render.go Artifacts loop (~486) still builds with +=; outside commands.go scope.
- panel.go:347 and presentation/detail.go:136 keep raw ansi.Truncate (other
  ellipsis / outside gui and modal).
- Panel-specific keys stay switches; the test checks the package's msg.String()
  cases, not each panel's.
