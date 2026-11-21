# Fix brief — lb16-update-when-stale, round r2

> Slice contract: `work/briefs/lb16-update-when-stale.md`. Findings are in
> `work/inbox/reviewer-lb16-update-when-stale-r1.md` (read it in full). One
> commit per finding that changes code, staged by name; tests first (show the
> failing run in the receipt).

## F-r1-01 (major) — the launch read is dropped while the update runs

`gui.go` drops a tagged launch read as soon as `launchStale` is set. The
check wins the race against the Status read in practice, so the active panel
is blank ("Status dashboard not loaded yet") for the whole `brew update`, and
stays blank if the enqueue fails. Fix: apply a launch read that lands while the
update runs, or after a failed enqueue. Drop one only if it lands **after the
completion's load-flagged refresh has been issued** (a flag set in the
brew-update completion, or a generation). Keep
`TestLaunchReadLandingAfterStaleUpdateIsDropped` passing.
Test, failing first: the stale check lands first, the launch Status read lands
during the update, the panel is loaded and `View()` shows the dashboard. Add
the failed-enqueue case.

## F-r1-02 (minor) — quoted `"true"`/`"false"` fail in `update_on_start`

`config.go`: map the string forms `"true"`/`"false"` like the bools. Add cases
to `update_test.go`.

## F-r1-03 (minor) — import grouping

Move the `internal/config` import into the module group in `header.go`,
`keys.go`, `gui_test.go`, `outdated_count_test.go`. `make fmt`.

## F-r1-04 (minor) — stale comments

`outdated_count_test.go` (~:119, ~:155): reword the comments that still say the
completion "must stay plain" / M9 forbids it.

## F-r1-05 (minor) — no test proves the user's bug

Add one test (unit) and one teatest flow: the runner reports 1 outdated before
`brew update` and 2 after. Assert the header count and the Outdated panel list
show 2 after the update (in `View()` for the flow).

Gate: `make test-fast PKG=...` per commit, `make ci` before the receipt.
Receipt: append a "Fix r2" section (≤ 20 lines) to
`work/inbox/runner-lb16-update-when-stale.md`, commit it
(`inbox: runner fix r2 lb16-update-when-stale`), reply with one done line.
