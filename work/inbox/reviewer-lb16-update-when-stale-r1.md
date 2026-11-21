verdict: changes-required

# Review — lb16-update-when-stale r1 (wip @ c95d81d; slice 75ecbba..bc26bb6)

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 last-update time | met | Real runner on this machine: `LastUpdate` = 2026-11-21 06:14:31.047935326, equal to the `FETCH_HEAD` mtime from `ls -l --full-iso`; runner calls = `[[--repository]]` only, doctor.go does not import os/exec; a repo with no `FETCH_HEAD` → zero time, nil err |
| AC-02 config | met (F-r1-02) | YAML `true`/`false`/`auto`/`yes`/empty/`~` load; default `auto`, 6; hours 0 and -1 → "brew.update_stale_hours: N is below 1"; `Auto`, `1` → clear error. Quoted `"true"`/`"false"` rejected (F-r1-02) |
| AC-03 Init | **not met** (F-r1-01) | stale → 1 update via `task.Manager`, then Outdated count + panel reload with the post-update list (my scratch test: outdated 1 → 2 after the update, count 2 known, Outdated list 2, with preload on and with the panel opened mid-update); fresh → no update, header `updated 2h ago`. But the active panel is blank for the whole update (AC-03.1 "so the screen is not blank") |
| AC-04 docs | met | README block and paragraph; DESIGN row; `false` "does not update behind your back" |
| Fix r1 (R under auto) | met | keys.go gate `!= config.UpdateNever && !m.isUpdating`; 3 tests |

**Gate (V1).** `make ci` exit 0 (all floors met, unchanged); `GOOS=darwin go build ./...` exit 0.
**V2.** Reverted each fix in a temp worktree: completion `RefreshMsg{}` → `TestLaunchCheckStale…` and `TestBrewUpdateCompletionLoadsOutdated` fail; drop disabled → `TestLaunchReadLandingAfterStaleUpdateIsDropped` fails; R gate `== UpdateAlways` → `TestRKeyUnderAuto…` fails ("0 times, want 1"); err/zero → fresh → `TestLaunchCheckUnreadableTimeIsStale` and 3 more fail.
**V4 race.** `CGO_ENABLED=1 go test -race -count=20` on the LaunchCheck/LaunchRead/RKey/Update tests in gui, flows, config, brew: all ok.
**V6.** My own `View()` at 2ad3c6f and HEAD, default config, mock runner, 80x24 and 140x40: only line 1 differs (`… 1 formulae  0 casks` → `… 1 formulae  0 casks  updated 2h ago`). No golden changed.
**V7.** No `program.Send`; the only new `go func` is the test helper `runWithin`; update goes through `tryEnqueue`; one commit per AC in order; floors unchanged. The test renames hold up: `…StaysPlain` → `…LoadsOutdated` follows the brief's reversal of M9, and `preloadConfig()` → `UpdateNever` is fine (those tests drive the preload directly; `…WaitsForUpdateOnStart` still sets `UpdateAlways`). The R tests now scan `drainCmds` with the same assertion, so they are not weakened.

## Findings

F-r1-01 (major) — internal/gui/gui.go:495 (with :327) — a launch read is dropped as soon as the check finds brew stale, not only once it lands after the update's refresh. `brew --repository` + stat takes 6 ms here, and the Status read (`brew info --json=v2 --installed`, 0.53 s) is slower, so in practice the check always wins and the Init read is always dropped. The screen then shows "Status dashboard not loaded yet — press R to refresh", with every sidebar panel "not loaded", for the whole `brew update`. If the enqueue fails (`launchStale` set, `isUpdating` reset), it stays blank until R. Evidence: scratch test `Update(lastUpdateMsg{})`, then a `launch:true` Status `DataLoadedMsg` → `isUpdating=true statusLoaded=false`, View unchanged and showing the "not loaded yet" text. — Fix must achieve: a launch read that lands while the update runs (or after a failed enqueue) is applied, so the active panel shows at once. Drop only a launch read that lands after the completion's load-flagged refresh has been issued (e.g. a flag set in the brew-update completion, or a generation). Test, failing first: stale check lands first, the launch Status read lands during the update, panel loaded and `View()` shows the dashboard. `TestLaunchReadLandingAfterStaleUpdateIsDropped` must still pass.

F-r1-02 (minor) — internal/config/config.go:72 — `update_on_start: "true"` / `"false"` (quoted) fail with `unknown value "true" (want true, false or auto)`, which contradicts itself. — Fix must achieve: the string forms `"true"`/`"false"` map like the bools; add them to update_test.go.

F-r1-03 (minor) — internal/gui/header.go:5, keys.go:4, gui_test.go:6, outdated_count_test.go:4 — the `internal/config` import sits in the stdlib group; the repo puts module imports in the second group. — Fix: move each into the module group.

F-r1-04 (minor) — internal/gui/outdated_count_test.go:119, :155 — leftover comments still say the brew-update completion "must emit a plain refresh … M9 forbids" / "must stay plain", above a test that now asserts the opposite. — Fix: reword them for the lb16 rule.

F-r1-05 (minor) — internal/gui/launch_check_test.go, flows/launch_check_test.go — every mock returns the same `brew outdated` before and after the update, so no test proves the user's bug: packages the pre-update read omitted appear after it. — Fix: one test whose runner reports 1 outdated before `brew update` and 2 after; assert the count and the Outdated panel list show 2 (in `View()` for the flow).

No brew write command was run; scratch work under /home/agent/.cache/lb16rev (worktree removed).
