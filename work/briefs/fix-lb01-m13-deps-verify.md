# Fix brief — slice `lb01-m13-deps-verify`, reviewer findings F-01 and F-05

> **Contract.** The slice contract is `work/briefs/lb01-m13-deps-verify.md`.
> This brief names the findings it must fix. Where the two differ, this brief
> wins for the fix cycle only.

## Context

The runner lane verified the M13 Deps-tab fix (commits `789d506`, `a7d5498`,
`196575b`, receipt `5841cc9` + `7cc8084`). The reviewer lane ran the gate and
five mutations and returned **`Verdict: changes-needed`** in
`work/inbox/reviewer-lb01-m13-deps-verify.md` (commit `923928c`).

Read that receipt before you start. It contains the evidence you need.

**There is no blocker.** The reviewer confirmed the tree is sound: all five
invariants are present, and **every one of them has a test that fails when it
is broken**. The defect is in the **verification record** — it proves
*presence* and *green tests*, never that the tests *bite*. That is what this
fix cycle fixes.

## The authoritative invariant → test mapping

Use this table. It is the corrected mapping (see F-05 — the original review
brief had I-2 and I-5 wrong).

| id | invariant | anchor | test that must FAIL when broken | the mutation that breaks it |
|---|---|---|---|---|
| I-1 | tab-content fetch is bounded by a context timeout | `internal/gui/commands.go:22`, `:833` | `TestFetchTabContentCmdTimesOut` (`internal/gui/gui_test.go:1276`) | replace the `context.WithTimeout(context.Background(), fetchTabContentTimeout)` at `:833` with `context.Background()` |
| I-2 | after `DataLoadedMsg` the tab content load is re-kicked | `internal/gui/gui.go:455-456` | `TestDepsTabPopulatedAfterDataLoaded` (`:1166`) **and** `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply` (`:1202`) | delete the re-kick `return m, m.loadTabContent()` at `:455-456` |
| I-3 | the Deps tab fast-paths from the cached formula's deps | `internal/gui/commands.go:779`, `:794`, `:810` | `TestDepsFastPathRendersCachedDependencies` (`:1314`), `TestDepsTabViewLeavesLoadingAfterFastPath` (`:1418`), `TestDepsTabLoads` (`:1446`) | make the fast path never apply — the `panelID == PanelFormulae && tabIdx == 2` condition at `:795` becomes `false` |
| I-4 | `parseFormula` maps build deps from `build_dependencies` | `internal/brew/formulae.go:318` (field `:64`) | `TestParseFormulaBuildDepsFromBuildDependencies` and `TestFormulaeReaderList` (`internal/brew/formulae_test.go`) | change `buildDeps := append([]string(nil), f.BuildDependencies...)` to `buildDeps := []string(f.Dependencies)` |
| I-5 | tab/panel key handlers return early so the fetch cmd is not dropped | `internal/gui/gui.go:627-634` | `TestDepsSelectionChangeUpdatesContent` (`:1370`) | remove one of the early returns in `:627-634` so the fetch command is dropped |

The reviewer's own runs confirm this mapping: M1 → `TestFetchTabContentCmdTimesOut`
failed; M2 → both `DataLoaded` tests failed; M3 → all three fast-path tests
failed; M4 → both brew tests failed; M5 → `TestDepsSelectionChangeUpdatesContent`
failed with "expected zlib deps after j".

## Findings to fix

One commit per finding, in order, staged by name — never `git add -A`, never
`git add .`.

### F-01 (major) — the record proves presence, not coverage, and overclaims

`work/verification/lb01-m13-deps.md` currently records only presence checks
(grep/sed) plus a passing test run. Its Conclusion says:

> "the four regression tests that cover it pass"

That is wrong twice over: there are **six** tests that cover the five
invariants, and the four contract-named tests cover only I-3, I-4 and I-5 —
they stay green when I-1 or I-2 is broken.

**Fix:**

1. **Add a mutation-evidence section** to the record. For each of I-1..I-5,
   record: the exact mutation you made (file, line, the edit), the test command
   you ran, and the **actual failing output** (the test name and the failure
   message from the test run). Real output, not a summary — the reviewer will
   re-run your mutations and compare.
2. **Rewrite the Conclusion** so it claims only what is now true: the fix is
   present in all five parts; each of the five invariants is covered by at
   least one test that **fails when the invariant is broken**; name those
   tests; the gate is green.
3. Keep the existing presence-check section — it is truthful and stays.

**You must perform the mutations yourself.** They are part of the evidence.
Revert every one. The worktree must be clean when you report.

Target file: `work/verification/lb01-m13-deps.md` (created in `789d506`).

### F-05 (minor) — a brief defect that was never written down

The review brief's invariant→test table was wrong for I-2 and I-5. The
reviewer caught it and correctly judged the slice itself blameless — but the
observation was never written in `work/questions.md`, and a question that is
not written there is a lost question.

**Fix:** append to `work/questions.md` — **append, do not rewrite** — an entry
recording this observation: the invariant→test mapping in the review brief was
wrong for I-2 and I-5, the correct mapping is the one above, the raiser is the
reviewer lane, the owner is the orchestrator, and the resolution is that the
mapping is corrected in `work/briefs/review-lb01-m13-deps-verify.md` and
recorded in the verification file. Because it is resolved, the
"_None. No questions are open._" line stays true — leave it in place.

Target file: `work/questions.md`.

## Findings deliberately NOT for this lane

- **F-02** (`LB-06` closed on a record that never reviewed the fix) — the
  **orchestrator** decides this one. The fix in F-01 turns the record into a
  genuine review, so `LB-01` and `LB-06` stay closed. **Do not touch
  `work/backlog.md`** — not the `LB-01` row, not the `LB-06` row, nothing.
- **F-03** (the receipt was committed twice) — accepted, content-harmless.
  History is not rewritten. **No action.**
- **F-04** (the stale "Next slice: `LB-01`" paragraph in `work/STATE.md`) —
  the **orchestrator** owns that edit. **Do not touch `work/STATE.md`.**

## Invariants

- **Exactly two commits**, one per finding, in F-01 → F-05 order.
- **No code changes.** The mutations are temporary experiments, reverted before
  you commit. Nothing under `internal/`, `cmd/`, `testdata/`, `scripts/`,
  `.github/`, `Makefile` may be changed in a commit. No test file is edited.
- The only files changed are: `work/verification/lb01-m13-deps.md` and
  `work/questions.md`.
- Do not touch `work/backlog.md`, `work/STATE.md`, `work/README.md`, the
  briefs, or any inbox receipt.
- Do not amend, rebase, or revert existing commits.
- `master-plan/` is a frozen local archive — do not read or modify it.
- One commit per finding, staged by name.

## Evidence to produce

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Then, for the mutations, run each one and record the real
failure output. Concretely:

```bash
# I-1
#   edit internal/gui/commands.go:833 -> context.Background()
go test ./internal/gui/ -run TestFetchTabContentCmdTimesOut -count=1
git checkout -- internal/gui/commands.go

# I-2
#   delete the re-kick at internal/gui/gui.go:455-456
go test ./internal/gui/ -run 'TestDepsTabPopulatedAfterDataLoaded|TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply' -count=1
git checkout -- internal/gui/gui.go

# I-3
#   make the fast path never apply at internal/gui/commands.go:795
go test ./internal/gui/ -run 'TestDepsFastPathRendersCachedDependencies|TestDepsTabViewLeavesLoadingAfterFastPath|TestDepsTabLoads' -count=1
git checkout -- internal/gui/commands.go

# I-4
#   internal/brew/formulae.go:318 -> buildDeps := []string(f.Dependencies)
go test ./internal/brew/ -run 'TestParseFormulaBuildDepsFromBuildDependencies|TestFormulaeReaderList' -count=1
git checkout -- internal/brew/formulae.go

# I-5
#   remove an early return in internal/gui/gui.go:627-634
go test ./internal/gui/ -run TestDepsSelectionChangeUpdatesContent -count=1
git checkout -- internal/gui/gui.go
```

Each mutation must make its test **fail**. If one does not, that is a blocker:
stop, write the receipt with the blocker named, and report. A mutation that
does not bite means the invariant is not covered, and `LB-01` cannot close.

## Report protocol

1. Write `work/inbox/runner-lb01-m13-deps-fix.md` — concise markdown, at most
   30 lines: the two commit ids in order, each finding answered with the
   evidence, the gate result, and per mutation the test name and its failure
   message.
2. Commit it:
   `git add work/inbox/runner-lb01-m13-deps-fix.md && git commit -m "inbox: runner reports lb01-m13-deps-fix"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `runner finished the lb01-m13-deps-verify fix cycle; receipt at work/inbox/runner-lb01-m13-deps-fix.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.
