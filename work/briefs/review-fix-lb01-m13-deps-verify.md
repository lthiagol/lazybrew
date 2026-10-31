# Re-review brief — fix cycle for slice `lb01-m13-deps-verify`

> **Contract.** The slice contract is `work/briefs/lb01-m13-deps-verify.md`; the
> fix contract is `work/briefs/fix-lb01-m13-deps-verify.md`. This brief tells
> you how to verify the fix. The contracts win where they differ.

## Context

The first review of `lb01-m13-deps-verify` returned **`Verdict: changes-needed`**
with five findings, recorded in
`work/inbox/reviewer-lb01-m13-deps-verify.md` (commit `923928c`).

**No blocker.** The reviewer confirmed the tree is sound: all five invariants
are present and every one has a test that fails when it is broken. The defect
was in the **verification record**, which proved presence and green tests but
never that the tests bite.

The fix cycle produced:

| Finding | commit | subject |
|---|---|---|
| F-01 | `b7dec88` | docs(verify): add mutation evidence and correct the lb01-m13-deps conclusion |
| F-05 | `544e0ed` | docs(questions): record the lb01-m13-deps-verify brief-mapping observation |

Its receipt is `work/inbox/runner-lb01-m13-deps-fix.md` (commit `8449f00`).

The orchestrator also made two edits that are yours to check:

| commit | subject |
|---|---|
| see `git log --oneline -3` | docs(brief): correct the invariant→test mapping in the review brief (F-05) |
| see `git log --oneline -3` | docs(state): the next-slice paragraph now names LB-02 (F-04) |

Your job: **re-derive the fix from the tree and prove it by command.** The fix
receipt is a claim, not evidence — and neither is the runner's mutation
evidence, which you must spot-check. Everything the first review already
verified must still hold.

## Verify by command

### S1 — F-01: the record now carries mutation evidence

```bash
cat work/verification/lb01-m13-deps.md
git show b7dec88 --stat
```

For each of I-1..I-5 the record must carry: the exact mutation (file, line,
edit), the test command, and the **actual failing output**. The presence-check
section must still be there.

**Spot-check by reproducing.** Pick **at least two** invariants — you should
pick **I-2 and I-5**, the two whose mapping the first review had to correct —
and reproduce the mutation yourself:

```bash
# I-2: delete the re-kick at internal/gui/gui.go:455-456
go test ./internal/gui/ -run 'TestDepsTabPopulatedAfterDataLoaded|TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply' -count=1
git checkout -- internal/gui/gui.go

# I-5: remove one of the early returns in internal/gui/gui.go:627-634
go test ./internal/gui/ -run TestDepsSelectionChangeUpdatesContent -count=1
git checkout -- internal/gui/gui.go
```

Both must fail, with failure messages that match what the record quotes. If the
record quotes an output the reviewer cannot reproduce, that is a finding.
Revert every mutation.

### S2 — F-01: the Conclusion no longer overclaims

The old Conclusion said "the four regression tests that cover it pass". The new
Conclusion must claim only what is true: the fix is present in all five parts;
**each of the five invariants is covered by at least one test that fails when
the invariant is broken**; those tests are named; the gate is green.

Count them: the mapping is I-1 → `TestFetchTabContentCmdTimesOut`; I-2 →
`TestDepsTabPopulatedAfterDataLoaded` + `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply`;
I-3 → `TestDepsFastPathRendersCachedDependencies` +
`TestDepsTabViewLeavesLoadingAfterFastPath` + `TestDepsTabLoads`; I-4 →
`TestParseFormulaBuildDepsFromBuildDependencies` + `TestFormulaeReaderList`;
I-5 → `TestDepsSelectionChangeUpdatesContent`. That is **nine** test names
across five invariants — not four.

### S3 — F-05: the question is recorded, and the brief is corrected

```bash
git show 544e0ed -- work/questions.md
grep -n -A8 'Q-03' work/questions.md
```

Q-03 must name the raiser (reviewer lane), the owner (the orchestrator), the
observation (the review brief's mapping was wrong for I-2 and I-5), and its
resolution. The "_None. No questions are open._" line must still be there.

And the review brief must now carry the corrected mapping:

```bash
grep -n -A6 'I-2' work/briefs/review-lb01-m13-deps-verify.md | head -20
```

The I-2 row must name the two `DataLoaded` tests, the I-5 row must name
`TestDepsSelectionChangeUpdatesContent`, and the brief must carry the
correction note.

### S4 — the orchestrator's state-record fix (F-04)

```bash
grep -n -B2 -A6 'Next slice' work/STATE.md
```

It must name `LB-02` as the next slice and must no longer claim `LB-01` is
next. The `lb01-m13-deps-verify` row must still be present and untouched by
this edit.

### S5 — scope: nothing but `work/` files

```bash
git diff --name-only 923928c..HEAD
git diff --name-only 923928c..HEAD -- internal/ cmd/ testdata/ scripts/ Makefile .github/ go.mod go.sum
# must print nothing
```

The first list may contain only files under `work/`. No test file, no Go file,
no Makefile, no workflow.

### S6 — the gate, run by you

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Report the real output.

## Mutations to run and revert

Prove the record's evidence is reproducible, not decorative:

- **P1** — repeat the I-2 mutation and the I-5 mutation from S1. Both tests
  must fail with the messages the record quotes. Revert.
- **P2** — break a test's teeth differently: in `internal/gui/commands.go:795`,
  make the fast path never apply, and confirm
  `go test ./internal/gui/ -run 'TestDepsFastPathRendersCachedDependencies|TestDepsTabViewLeavesLoadingAfterFastPath|TestDepsTabLoads' -count=1`
  fails with the message the record quotes for I-3. Revert.

A mutation whose failure does not match the record is a finding: the evidence
was written rather than observed.

`git status` must be clean when you report.

## Open questions

Read `work/questions.md`. Report whether Q-03 (and the earlier Q-01, Q-02) are
answered, missed, or newly raised, and whether this fix cycle raised anything
new.

## Scope you do not own

- You do not fix anything — not the record, not the briefs, not the state row.
- You mutate only to prove a check bites, and you revert it.
- You do not touch the contents of `master-plan/`.
- You do not rewrite the verification record to be clearer or shorter.
- You do not prompt the runner lane and you do not spawn agents.
- You commit only your own receipt file.

## Verdict format

Your receipt must end with one line:

```
Verdict: ok
```

or

```
Verdict: changes-needed
```

followed by findings, each as `F-NN | severity | what is wrong | the command
that proves it | the fix`. Severity is `blocker`, `major`, or `minor`.

## Report protocol

1. Write `work/inbox/reviewer-lb01-m13-deps-fix.md` — concise markdown, at
   most 40 lines: the verdict, any findings, the evidence (command → result)
   for S1–S6, the outcome of each mutation P1–P2, and the state of the
   questions.
2. Commit it:
   `git add work/inbox/reviewer-lb01-m13-deps-fix.md && git commit -m "inbox: reviewer reports lb01-m13-deps-fix"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `reviewer finished the lb01-m13-deps-verify fix cycle; verdict in work/inbox/reviewer-lb01-m13-deps-fix.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.
