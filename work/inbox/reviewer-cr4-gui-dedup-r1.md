verdict: changes-required

# reviewer: cr4-gui-dedup r1 (base db7d464, HEAD 8a02d11)
## Gate and harness
- V1: `make ci` PASS (gofmt, vet, race, all floors, build); `GOOS=darwin go build ./...` PASS.
- V3 frames (own harness on exported gui.New/Update/View, worktrees db7d464 and HEAD, one replayed
  cache of real read-only brew output; writes recorded, never run): panels 1-7, every tab, help +
  paging, log toggle, every confirm/input/menu modal, the overlay, footers. 832 frames (dark+light ×
  80x24+140x40); 116 differ raw, all on header line 0 ("updated 9s/10s ago"); 0 after normalising.
- V3 brew args: 164 write calls per version, identical: update, install, uninstall, zap, reinstall,
  upgrade (formula/cask/batch/all), fetch, pin, tap, untap, (un)trust, tap --repair, services
  start/run/cleanup, cleanup -n, autoremove, bundle dump, vulns. Not reached: services stop/restart.
- V2: the one regression test (TestCaskUninstallConfirmRecordsNoUsesCall) on 2248ac7 (pre-AC-06)
  FAILS: `cask uninstall confirm asked brew for dependents: [uses --installed firefox]`.
- V7: deadcode/staticcheck not run (fetch denied). Grep: all 14 added funcs have production callers;
  clipLine, both fitLine, cutCell, withIcon, boolStr, selectedOutdatedFormula, select{Service,Tap}Named,
  helpItem are gone.

## AC table
| AC | verdict | evidence |
|---|---|---|
| 01 | ok | only `Enqueue(` left: tryEnqueue and the batch loop (tasks_packages.go:208, counts per item); isUpdating reset kept via tryEnqueue |
| 02 | **F-r1-01** | helpers + 10 tests; lines not lost/reordered (manager reads to close of out after err, manager.go:173-184); cancel path hangs |
| 03 | ok | packageWriter picks Formulae/CasksWrite once; Run closures capture `client` |
| 04 | ok | textutil Clip/Fit + tests; tabWidth, l.sideW, leftBlock in menu |
| 05 | **F-r1-02** | tables, Hints(), single 1-7 / `[`,left / `]`,right cases; frames byte-identical |
| 06 | ok | toggleLeaves on panelData, pendingLabel (no fallback), Q4 test, Q5 sites use Builder |
| 07 | ok | tabs.go; no bare tab indices in render.go/panel.go |
| 08 | ok | see judgement (1); largest file commands.go 712; 2124→712, 1499→675 |

## Judgements asked for
1. AC-08: a whitespace-insensitive line-multiset diff of 1dd15c8 leaves only imports/package lines,
   handleKey's doc+header+`var cmd`, the `return m.handleKey(msg)` call, the `msg.(tea.KeyPressMsg)`
   assertion becoming the parameter, and the 7-line modal pass-through. The extraction is what the
   brief's own `keys.go (key dispatch)` requires; the pass-through copies Update's tail that
   fall-through keys reached before. No behaviour change (frames + args identical). Met.
2. dispatchKeys is a method declared in keymap_test.go, i.e. test code, not a production function;
   the "no unreachable functions" invariant is not breached. (`keyHint.keys` is data read only by tests.)
3. panel.go:375 truncates with "..." (three dots), not "…": Clip would change frames, so it is outside
   R8's `ansi.Truncate(…, "…")` pattern. presentation/detail.go:136 is in package presentation, not
   `internal/gui` or `internal/gui/modal`. Both out of scope; not findings. 4: F-r1-01. 5: F-r1-02. 6: met (V7).
## Findings
F-r1-01 (major) — internal/gui/task_build.go:84-88 — on ctx cancel linesTask returns without sending on
errCh; readChunk then blocks forever on `rt.err = <-rt.errCh` (task/manager.go:186-188): no
TaskCompletedMsg, the overlay stays "running", the queue stalls. Probe (scratch worktree): enqueue a
100-line linesTask, RunNext, read one chunk, CancelCurrent → no TaskCompletedMsg in 2 s. runDoctor/
runMissing had the same hang at base; runVulns did not (unguarded buffered send, errCh always nil) and
now can (select may pick ctx.Done). TestLinesTaskCancelMidStream asserts "errCh delivers nothing",
encoding the hang. Fix must achieve: a cancelled linesTask always ends with an error on errCh
(e.g. ctx.Err()); a test through task.Manager shows CancelCurrent mid-stream yields TaskCompletedMsg.

F-r1-02 (major) — internal/gui/keymap_test.go:80-104,129-158 — the test does not fail when a handled key
is missing from its footer. Mutation: drop stDoctor from panelHints(PanelStatus) → `go test
./internal/gui/` passes (d is still dispatched on Status). The check is a package-wide union of
strings, footers are only checked ⊆ table, and keyTable is hand-listed in the test. Fix must achieve:
per context, each key dispatched there is in that context's footer (or help where the brief allows)
and vice versa; the stDoctor mutation fails.

Seen, not done: read cmds executeSearch/fetchSelectedSearchInfo/toggleLeaves still capture `m` (not
task Run closures, outside Q2); render.go:486-488 Artifacts `+=` loop (not a Q5 site).
