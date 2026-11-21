# Slice: lb21-backlog-sweep — LB-13, LB-14, LB-15, LB-18

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask. User
> asked for all four (2026-11-21); rulings below are the orchestrator's.

**Depends on:** `lb20-cask-followups` closed. Never launch the TUI; no brew write
commands.

## AC-01 — LB-13: footers show the tab keys and Status navigation

Ruling: yes, show them. The Status, Outdated and Taps footers add `h/l tabs`
(`keyPanelTabs`, as Formulae, Casks and Services do); Status footer also shows
its `j/k` (it lists rows). Empty `footerExempt` of every entry this removes
(keep only the Search entries that stay true: `left`/`right`, `X`) so the keymap
test pins the new footers. Check the footer still fits at **80x24** on every
panel (it truncates or wraps today? say which, and drop the lowest-priority
hint rather than overflow if it does not fit). Name every golden that changes
(footers only). Own `View()` diff base vs HEAD: only footer rows differ.

## AC-02 — LB-14: `TestSingleflightPropagatesError` without a 20 ms sleep

`go.mod` is `go 1.26`, so `testing/synctest` is stable. Use `synctest.Test` and
`synctest.Wait()` so the test knows the second caller is blocked on the first
fetch, with no production hook and no sleep. Prove it with `-race -count=200`
passing, and fail by reverting the singleflight sharing (the second caller
would run its own fetch). If synctest cannot do it here, say why and keep the
sleep with a comment, and close LB-14 as "won't fix" in the receipt.

## AC-03 — LB-15: narrow `brew.SetDebug`

`brew.SetDebug` exists for `app/debug_log_test.go`; its `true` branch logs to
stderr, which corrupts the TUI (C13). Find how production sets the debug
logger (`app.enableDebugLog` or similar). Make production the only place that
turns debug on, writing to the debug log file, and replace `SetDebug` with a
reset-only helper (`ResetLogger()` or equivalent) used by the test. After it,
no code path writes the brew debug log to stderr; `deadcode -test ./...` and
`staticcheck ./...` (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. A test proves enabling debug writes to the file, not stderr.

## AC-04 — LB-18: `TestSearchTabCyclesAllPanelsFlow` is fast

It runs one teatest program per key prefix and takes about 21 s under `-race`.
Rewrite it to a single program (send the keys one at a time and check `View()`
after each, using the final-model or a polling helper the flows package already
has) or drive the Model directly with a real `View()` assertion; target under
3 s with `-race`. It must still fail with the AC-01 (lb19) Tab fix reverted,
both directions, all 7 panels: show the mutation in the receipt.

## Done means

`make ci` green; `GOOS=darwin go build ./...`; floors not lowered. Receipt
`work/inbox/runner-lb21-backlog-sweep.md` (≤ 40 lines, committed), one commit
per AC, tests first where a bug is fixed. Strike LB-13/14/15/18 in
`work/backlog.md` only in the orchestrator's close commit.
