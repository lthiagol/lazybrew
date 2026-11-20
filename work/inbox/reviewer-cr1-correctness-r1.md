verdict: ok
# Review — cr1-correctness, round r1

Scope `fe1d103..10c84f0`: 6 AC commits, receipt `5f9057e`, ruling `10c84f0`.

## V1 gate
`make ci` exit 0 (gofmt, vet, `-race` tests, floors app 70.0/64 brew 81.0/78
gui 86.4/83, build). `GOOS=darwin go build ./...` ok.

## V2 per AC — fix reverted in a temp worktree, new tests re-run
| AC | Reverted file(s) | Result without the fix |
|---|---|---|
| AC-01 [C1] | commands.go, gui.go, render.go @`02bd452^` | 7 `TestLong*` FAIL, e.g. `got [[uninstall --cask font-jetbrains-mono-ne]]`, `[[untap someuser/homebrew-a-ve]]`, `[[uses --installed some-very-long-formula]]`. Services-stop and Search-install pass (guards; disclosed in the receipt) |
| AC-02 [C2] | runner.go | `WARNING: DATA RACE` ×3; `OnExec got <nil>, want brew install failed` |
| AC-03 [C6] | commands.go, types.go, casks.go | `Version = "", want 1.8.2`; `Bottled = false`; View lacks `1.8.2 (bottled)`; cask: `cannot unmarshal array into ... .name` |
| AC-04 [C5] | gui.go | key: `nil pointer dereference`; paste: `activeModal = gui.notAModal ..., want nil` |
| AC-05 [C3,C4] | gui.go, messages.go | `6 tick commands in flight ..., want 1`; update ticker `3 ..., want 1` |
| AC-06 [C12,C13] | panel.go / app.go (each alone) | `error cut split a rune`; `a debug line reached stderr after setup` |

## V3 AC-01
`grep extractPackageName|selectedItemName` (non-test): only the definition
`gui.go:1447`; `selectedItemName` is deleted (cr5 brief allows "if not already
gone"). The receipt gives a count (16) but does not list the sites, as the brief
asks. Listed from the diff instead, 16 sites: commands.go `executeTrustMenuAction`,
`executeTrustAction`, `executeUntap`, `executeRepair`, `doMutation`,
`batchUpgrade` (`nameAt`), `openUntrustConfirm`, `startTrustMenu`, `togglePin`,
`tabItemName`, `confirmUntap`, `confirmRepair`, `confirmUninstall`,
`confirmMutation`; gui.go leaves filter (`nameAt`); render.go tab fetch.
These match the grep; the evidence is here, so this is not a finding. `nameAt`
alignment checked: Outdated = formulae then casks (same as
`selectedOutdatedCask`); the leaves filter narrows `items` and `formulae`
together; Search `p.results` is set with `p.items` in `SearchDoneMsg`.

## V4 AC-05
Arms of `autoRefreshCmd`: Init (`gui.go:238`) and `armAutoRefresh` (tick and
paused handlers only). Arms of `updateTickerCmd`: Init (`gui.go:223`) and
`armUpdateTicker` (UpdateTickMsg only). The `RefreshMsg` arm and the R-key
arm are gone. Pause: `evaluateAutoRefresh` is unchanged; the tick now arms the
next tick, and the paused handler's arm is a no-op. The cadence is the same.
The existing pause tests were changed only to read a batch (`finishedMsgs`),
not weakened. Init arming in both branches was ruled accepted (Q-cr1-init-autorefresh).

## V5 invariants
`p.items`, `presentation.Format*`, `extractPackageName` present; no golden/
footer/help changes. go.mod unchanged (`x/ansi` already direct).
No `program.Send`. The new goroutine is inside `LoggingRunner` (brew pkg, sudo
pattern), not in a handler. One commit per AC, in order, ids in subjects.

## Findings
None.
## Notes (not findings)
- `parseCask` now reads brew's string `installed`, so the Casks panel shows
  the installed version, not the latest. This matches formulae, AC-03's cask
  assertion needs it, and the commit message says so. It is visible to users.
- `TestTogglePinLooksUpPinnedStateByName`: the fixture is now index-aligned. The
  old misaligned state can't happen any more (the leaves filter narrows both
  lists), and the receipt doesn't mention this change.
- `tick_chain_test.go` counts "pending" after a 300 ms cutoff; a very slow
  `-race` run could flake. gui.go:1161 (generic modal route) can't panic and
  is outside the brief's two routes (receipt: "seen, not done").
