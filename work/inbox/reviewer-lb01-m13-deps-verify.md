# Receipt — reviewer: lb01-m13-deps-verify

Verdict: changes-needed

Gate run by me, not taken from the runner's claim. All five invariants are present, the five AC commits are in order and correctly scoped, and **no blocker**: every one of the five invariants does have a test that fails when it is broken. The gap is in the verification record, not in the tree — it proves *presence* and *green tests*, never that the tests *bite*.

## Findings

F-01 | major | The verification record proves presence, not coverage, and overclaims in its Conclusion. `work/verification/lb01-m13-deps.md` records only grep/sed presence checks for I-1..I-5 plus four passing tests; it contains no mutation evidence. Its Conclusion says "the four regression tests that cover it pass" — but the four contract-named tests cover only I-3, I-4 and I-5: they stay green when I-1 or I-2 is broken. | `internal/gui/commands.go:833` → `context.Background()`; then `go test ./internal/gui/ -run 'TestDepsTabLoads|TestDepsSelectionChangeUpdatesContent' -count=1` → `ok`, exit 0. Delete the re-kick at `internal/gui/gui.go:455-457`; same command → `ok`, exit 0. | Add per-invariant mutation evidence to the record and name the tests that actually bite — I-1 `TestFetchTabContentCmdTimesOut` (`internal/gui/gui_test.go:1276`); I-2 `TestDepsTabPopulatedAfterDataLoaded` (`:1166`) and `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply` (`:1202`); and correct the Conclusion's coverage claim.

F-02 | minor | LB-06 — "the M13 independent review itself" — is closed on a record that never reviewed the fix. | `work/backlog.md:17` = `closed-by-LB-01`; `work/verification/lb01-m13-deps.md` holds presence checks and a green test run only, no assessment of the fix's correctness. | Orchestrator decides whether LB-06 stays closed without the mutation evidence F-01 asks for.

F-03 | minor | The runner receipt was committed twice; the contract asked for one receipt commit. | `git log --oneline` → `5841cc9`, then `7cc8084`; `git show 7cc8084 --stat` → one file (`work/inbox/runner-lb01-m13-deps-verify.md`, 22+/34−), no code touched. | Content-harmless — the second commit only brought the receipt under the brief's 30-line limit — so not a blocker. Optionally squash before the slice closes.

F-04 | minor | Stale `work/STATE.md` trailing paragraph. Judged, not fixed — the orchestrator owns that edit. | `work/STATE.md:28-30` still reads "**Next slice:** `LB-01`" while `work/backlog.md:12` has LB-01 `closed`. | Outside AC-03, which named only the row; the runner flagged it correctly. Severity minor: the state row itself is correct and the paragraph names a real, now-closed item. Orchestrator rewrites it to name `LB-02`.

F-05 | minor | The review brief's invariant→test table is wrong for I-2 and I-5. The brief is the orchestrator's file, not the slice's. | I-2's named test `TestDepsSelectionChangeUpdatesContent` never sends `DataLoadedMsg`; I-5's three named tests never reach `internal/gui/gui.go:627-634` — `TestTabKey` only exercises `tabKey()`, `TestTabSwitching` uses `[`/`]` → `prevTab`/`nextTab`, `TestDataLoadedOnNonFetchTabDoesNotReissueFetch` drives `DataLoadedMsg`. All still pass under M2/M5. | Correct the mapping in `work/briefs/review-lb01-m13-deps-verify.md`. The slice itself is not at fault.

## Evidence (command → result)

- **V1 ✔** All five proving commands re-run by me (`grep -n "fetchTabContentTimeout" internal/gui/commands.go` → 18/22/833; `sed -n '455,456p' internal/gui/gui.go`; `grep -n "renderCachedTabContent" internal/gui/commands.go` → 779/788/794; `grep -n "runtime := f.Dependencies"` → 810; `grep -n "buildDeps := append" internal/brew/formulae.go` → 318; `grep -n "BuildDependencies   \[\]string"` → 64; `sed -n '627,634p' internal/gui/gui.go`): every recorded output reproduced byte-for-byte. `git show 789d506 -- work/README.md` → exactly one added row for `work/verification/`, nothing else changed in that file. The record is truthful about what it ran; only its Conclusion overreaches (F-01).
- **V2 ✔** `git show a7d5498 -- work/backlog.md` → LB-01 `active`→`closed` plus a 3-line note naming `work/verification/lb01-m13-deps.md` and `789d506`. `git diff 46e9baa..HEAD -- work/backlog.md` → only the LB-01 row and the note. `grep -n -E 'LB-0[1-6]|LB-DROP-01' work/backlog.md` → LB-06 still `closed-by-LB-01`; LB-02, LB-03, LB-04, LB-05 and LB-DROP-01 byte-identical.
- **V3 ✔** `git show 196575b -- work/STATE.md` → one row added, `| lb01-m13-deps-verify | — | implemented | awaiting reviewer |`; the `detach-master-plan` row is untouched.
- **V4 ✔** `git diff --name-only 46e9baa..196575b` → exactly `work/README.md`, `work/STATE.md`, `work/backlog.md`, `work/verification/lb01-m13-deps.md`. The same diff restricted to `internal/ cmd/ testdata/ scripts/ Makefile .github/ go.mod go.sum` printed nothing. No test file touched; `master-plan/` not read or modified.
- **V5 ✔** Gate run by me: `make build` → exit 0; `make lint` (`go vet ./...`) → exit 0; `go test ./... -race -count=1` → exit 0, every package `ok` (2 with no test files); `make cover-check` → "All coverage floors met" (brew 64.3/62, presentation 95.4/91, gui 55.9/36, modal 41.4/41), exit 0.

## Mutations (each reverted; `git status` clean after each)

| mut | inv | test named by the review brief | result | test that actually bit |
|---|---|---|---|---|
| M1 | I-1 | `TestFetchTabContentCmdTimesOut` | **FAIL** — `gui_test.go:1310: fetchTabContentCmd hung past fetchTabContentTimeout` | same |
| M2 | I-2 | `TestDepsSelectionChangeUpdatesContent` | **PASS** — no bite | `TestDepsTabPopulatedAfterDataLoaded` (`gui_test.go:1187`) and `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply` (`gui_test.go:1220`) — both FAIL |
| M3 | I-3 | `TestDepsFastPathRendersCachedDependencies`, `TestDepsTabViewLeavesLoadingAfterFastPath`, `TestDepsTabLoads` | **FAIL** — all three | same |
| M4 | I-4 | `TestParseFormulaBuildDepsFromBuildDependencies`, `TestFormulaeReaderList` | **FAIL** — both | same |
| M5 | I-5 | `TestTabKey`, `TestTabSwitching`, `TestDataLoadedOnNonFetchTabDoesNotReissueFetch` | **PASS** — no bite | `TestDepsSelectionChangeUpdatesContent` — FAIL, `gui_test.go:1403: expected zlib deps after j` |

**No blocker on coverage.** M1→`TestFetchTabContentCmdTimesOut`; M2→the two `DataLoaded` tests; M3→the three fast-path tests; M4→both brew tests; M5→`TestDepsSelectionChangeUpdatesContent` (a contract-named test). The tree does cover all five invariants; the record does not show it (F-01). Only implementation files were mutated, every mutation was reverted with `git checkout --`, and the worktree is clean.

## Questions

`work/questions.md` — the open list reads "_None. No questions are open._"; both historical Q-01 and Q-02 are marked resolved. **Nothing was answered, missed, or newly raised by this slice.** F-05 is a new observation for the orchestrator (a brief defect, not a lane question) and is not written in `work/questions.md` — worth appending there.

Verdict: changes-needed
