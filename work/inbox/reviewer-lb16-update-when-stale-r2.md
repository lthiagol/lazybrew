verdict: changes-required

# Review — lb16-update-when-stale r2 (wip @ 29712cd; fix commits fdd4f41..cd32462)

Only two minor findings remain, both cosmetic. No blocker or major. The five r1 findings are all fixed.

| r1 finding | Verdict | Evidence (revert check in a temp worktree at 29712cd) |
|---|---|---|
| F-r1-01 launch read dropped during update | fixed | gui.go:497 drops only when `updateRefreshed` (set in the brew-update completion, :361). Revert to `launch && launchStale`: `TestLaunchStatusLandingDuringStaleUpdateShowsDashboard` and `…AfterFailedEnqueueIsApplied` fail. Never drop: `TestLaunchReadLandingAfterStaleUpdateIsDropped` fails. My scratch teatest flow (Status read held 80 ms, `brew update` held open, quit mid-update, `View()` of the final model): at HEAD the Status sidebar shows `Taps 0`; with the fix reverted it shows `1 ◉ Status … not loaded`. |
| F-r1-02 quoted "true"/"false" | fixed | config.go:84. With the pre-fix config.go, `TestUpdateOnStartValues/"true"` and `/"false"` fail. |
| F-r1-03 import grouping | fixed as listed | The four named files are fixed. Three more from the same slice still have the slip (F-r2-01). |
| F-r1-04 stale comments | fixed | outdated_count_test.go:118 and :153 are reworded. One more stale comment remains (F-r2-02). |
| F-r1-05 user's bug untested | fixed | Runner returns 1 outdated before `brew update` and 2 after. With the completion changed to `RefreshMsg{}`: `TestStaleUpdateShowsEveryOutdatedPackage`, `TestLaunchCheckStaleFlow` (asserts `2 outdated` and `wget` in `View()`) and 2 more fail. With `RefreshMsg{ReloadDiagnostics:true}`: the unit test fails ("Outdated after the update = {N:0 Known:false}"). |

**Gate (V1).** `make ci` exit 0: gofmt, vet, `-race`, all floors met and unchanged (config 89.1≥85, gui 90.0≥83), build. `GOOS=darwin go build ./...` exit 0.
**Race hunt (V4).** `CGO_ENABLED=1 go test -race -count=20 -run 'TestLaunch|TestStale|TestRKey|TestRev|Update|TestRefreshKey'` over gui, flows, config and brew: all ok, no DATA RACE. It included three scratch flows I wrote for this round, all using real tea.Cmd concurrency:
- the launch read lands during a held-open update, then a resize and `R` are sent, then the update ends → `brew update` ran once and the header shows `2 outdated`;
- quit during the update → the program finishes;
- the final-view check above.
**V6/V7.** No golden changed. No `program.Send`, no new goroutine in production code. The update still goes through `tryEnqueue`. One commit per finding, in order, each with a failing run shown. No test deleted or weakened.
**Observation (not a finding).** During any brew update, the Operation overlay (`⟳ brew update · Running · Esc to cancel`) owns the main pane (AGENTS rule 4, as under `true`). So what the F-r1-01 fix changes on the real screen is the Status sidebar list, not the dashboard tab. The unit test's `renderContent` check runs without `opState`, so it never sees the overlay.

## Findings

F-r2-01 (minor) — internal/gui/brew_update_test.go:5, gui_misc_test.go:5, tick_chain_test.go:4 — the `internal/config` import added in 0ecc2ba (lb16 AC-02) sits in the stdlib group. This is the same slip as F-r1-03, which listed only four files. Evidence: an awk scan of every `.go` file the slice touched; these three are the only remaining hits. — Fix must achieve: move each into the module group; the scan finds none.

F-r2-02 (minor) — internal/gui/outdated_count_test.go:25-27 — the doc of `TestNonRRefreshEmittersStayCacheOnly` still says "only an explicit R may load the outdated count". Since lb16, the brew-update completion loads it too. The test's cases (install/uninstall) are still right. — Fix must achieve: reword it to say the install, uninstall and mutation emitters stay plain, and that R and the brew-update completion are the two that load the count.

No brew write command was run. Scratch work was under /home/agent/.cache/lb16rev2 (worktree removed).
