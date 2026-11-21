# Slice: lb22-update-default-and-leftovers — update on start by default; LB-20, LB-21

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb21-backlog-sweep` closed. User ruling (2026-11-21): the
default of `brew.update_on_start` becomes `true`. Never launch the TUI; no brew
write commands.

## AC-01 — default `update_on_start: true`

1. `config.Default()` sets `UpdateOnStart: UpdateAlways`. `auto` and
   `update_stale_hours` stay supported, unchanged. The sample (generated from
   `Default()`) now prints `update_on_start: true` uncommented, with `auto` and
   `false` as the commented alternatives; its comment says what each does.
2. README (Configuration block and paragraph) and DESIGN: the default is `true`
   (`brew update` at every launch); say how to keep the old behaviour
   (`auto`, `false`).
3. Tests: default config → every launch enqueues one `brew update` through
   `task.Manager` and the Outdated count is loaded after it; `false` and
   `auto` behave as lb16 pinned. The many gui tests that build a Model from
   `config.Default()` must not start updating by accident: make the shared test
   helper set `UpdateNever` explicitly (one place), and name it in the
   receipt. Frames: the header shows `updating…` at launch under the default;
   name each golden that changes and why.
4. An offline launch (`brew update` fails) must leave the TUI usable: the
   failure toast shows, the panels load from the old metadata. Test it.

## AC-02 — LB-20: two Search edges

(a) If one typed read exits 1 (no results) and the other fails with a real
error, the user must see the error toast, not a silent empty list: return the
rows of the read that worked, and the error of the other **only when the rows
are empty**. Keep `Q-lb19-search-partial` otherwise (rows kept, no error when
at least one row exists). (b) Two uninstalls finishing back to back must not
let a slow first re-mark restore a stale mark: tag each re-mark with a
generation (the model's counter, bumped when a re-mark is issued) and drop a
`SearchMarksMsg` whose generation is not the latest. Failing tests first for
both.

## AC-03 — LB-21 leftovers

1. The three other singleflight tests with 20 ms sleeps use `synctest`, as
   `TestSingleflightPropagatesError` does since lb21. Prove with `-race
   -count=100`.
2. `EnableFileLogging` closes its log file on exit (return a closer or a
   shutdown hook the app calls; follow the existing app lifecycle). Test it.
3. Fix the `ResetLogger` doc in `internal/brew/logger.go` (`init()` calls it).

## Done means

`make ci` green; `GOOS=darwin go build ./...`; floors not lowered; deadcode and
staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Receipt `work/inbox/runner-lb22-update-default-and-leftovers.md`
(≤ 40 lines, committed), one commit per AC, tests first for bugs.
