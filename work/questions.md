# Questions — open questions

Append to this file. Never rewrite it: the history of what was asked and when is
the point.

## Open questions

_None. No questions are open._

## Rules

- The orchestrator owns the open questions. A lane does not answer them itself.
- A lane that finds something ambiguous in its brief writes it here and keeps
  going on the criteria it does not affect.
- The reviewer reports each question as **answered**, **missed**, or **newly
  raised**.
- A question that is not written here is a lost question.

## Questions raised late — slice `detach-master-plan` (runner lane)

Both items below were found by the runner lane while executing
`detach-master-plan` and recorded only in its receipt. They are appended here
for the record. The "no questions are open" line above stays true: both are
resolved.

### Q-01 — the `.mp-write.lock` contradiction (raised late)

- **Question:** AC-04 said the smoke checklist was "tracked only inside
  `master-plan/`", but git also tracked `master-plan/.mp-write.lock` — so the
  checklist was not the only tracked file there. Which of the two readings
  wins?
- **Raised by:** runner lane (slice `detach-master-plan`).
- **Owner:** the orchestrator.
- **Resolution:** resolved — the checklist was copied out to
  `SMOKE-CHECKLIST.md` and the directory was untracked with no on-disk file
  modified; the extra tracked lock file was removed from the index with the
  rest of the directory.

### Q-02 — the LB-06 clause ambiguity (raised late)

- **Question:** AC-04 required LB-06 to be "closed-by-verification with
  LB-01", while also saying "every row's status starts as `active` except the
  dropped row". The two clauses conflict for LB-06 specifically. Which one
  wins?
- **Raised by:** runner lane (slice `detach-master-plan`).
- **Owner:** the orchestrator.
- **Resolution:** resolved — the more specific clause wins: LB-06's status is
  `closed-by-LB-01`.

## Questions raised late — slice `lb01-m13-deps-verify` (reviewer lane)

The item below was found by the reviewer lane while reviewing
`lb01-m13-deps-verify` and recorded only in its receipt. It is appended here
for the record. The "no questions are open" line above stays true: it is
resolved.

### Q-03 — the review brief's invariant→test mapping was wrong for I-2 and I-5 (raised late)

- **Observation:** the invariant→test table in
  `work/briefs/review-lb01-m13-deps-verify.md` mis-mapped two of the five
  invariants. For I-2 it named `TestDepsSelectionChangeUpdatesContent`, which
  never sends a `DataLoadedMsg`; for I-5 it named `TestTabKey`,
  `TestTabSwitching` and `TestDataLoadedOnNonFetchTabDoesNotReissueFetch`, none
  of which reach `internal/gui/gui.go:627-634`. The reviewer's own mutations
  confirmed it: under M2 and M5 those named tests still passed — they did not
  bite. The slice itself was not at fault.
- **Correct mapping:** the authoritative table is the one in
  `work/briefs/fix-lb01-m13-deps-verify.md`:

  | id | test that fails when the invariant is broken |
  |---|---|
  | I-1 | `TestFetchTabContentCmdTimesOut` |
  | I-2 | `TestDepsTabPopulatedAfterDataLoaded`, `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply` |
  | I-3 | `TestDepsFastPathRendersCachedDependencies`, `TestDepsTabViewLeavesLoadingAfterFastPath`, `TestDepsTabLoads` |
  | I-4 | `TestParseFormulaBuildDepsFromBuildDependencies`, `TestFormulaeReaderList` |
  | I-5 | `TestDepsSelectionChangeUpdatesContent` |

- **Raised by:** reviewer lane (slice `lb01-m13-deps-verify`).
- **Owner:** the orchestrator.
- **Resolution:** resolved — the mapping is corrected in
  `work/briefs/fix-lb01-m13-deps-verify.md` and recorded, with the mutation
  evidence that proves each of the five tests bites, in
  `work/verification/lb01-m13-deps.md`. The stale rows in
  `work/briefs/review-lb01-m13-deps-verify.md` are the orchestrator's file and
  were left untouched by the fix lane.
- **Update (appended, after the fix cycle):** the last sentence above was
  written when the fix lane reported and was already out of date by the time
  it was read. The orchestrator corrected the stale rows in
  `work/briefs/review-lb01-m13-deps-verify.md` in commit `fa32989`
  ("docs(brief): correct the invariant-test mapping flagged by the reviewer"),
  and the re-review confirmed the correction (see
  `work/inbox/reviewer-lb01-m13-deps-fix.md`, check S3). Q-03 is closed; this
  line records that the correction happened, not that it is pending.

## Questions raised late — slice `lb03-dashboard-outdated-count` (runner lane)

### Q-04 — AC-05.5's "exactly one `brew outdated` call" is unreachable on a cold cache

- **Question:** AC-05.5 asks that an explicit load produce "exactly one
  `brew outdated` call". Brew has two outdated readers with one cache key
  each (`KeyOutdatedFormulae`, `KeyOutdatedCasks`), each of which shells
  `brew outdated` with its own `--formula` / `--cask` filter, so a complete
  count on a cold cache costs two calls. One combined call would mean
  changing `internal/brew/**`, which the brief forbids. Which reading is
  the contract: one call per outdated class, or one call in total?
- **Raised by:** runner lane (slice `lb03-dashboard-outdated-count`).
- **Owner:** the orchestrator.
- **Resolution:** the runner implemented the per-class reading and covered
  both in `internal/gui/gui_test.go`:
  `TestStatusDashboardExplicitLoadShellsOutdatedOnce` (the casks class is
  already cached → the load makes exactly one `brew outdated` call and shows
  the real count) and
  `TestStatusDashboardExplicitLoadShellsOneCallPerOutdatedClass` (cold cache →
  exactly two calls, one per class). The brief's wording is met literally in
  the first case; the second documents the two-reader reality.
- **Orchestrator ruling (appended):** the **per-class** reading is the contract.
  AC-05.5's "exactly one `brew outdated` call" was imprecise wording on my
  part; the requirement it meant was that *only an explicit `R` may cause a
  `brew outdated` call at all*, and that the resulting count must be
  complete. One call per outdated class satisfies both: two calls on a cold
  cache (formulae and casks are separate readers with separate cache keys and
  separate `--formula` / `--cask` filters), one when the other class is
  already cached. Combining them into a single call would mean changing
  `internal/brew/**`, which the slice deliberately does not touch — and it
  would diverge from the Outdated panel's own path, which already makes two
  calls. The brief's AC-05.5 has been corrected to state the per-class
  wording; the contract is unchanged in every other respect. Q-04 is closed.
