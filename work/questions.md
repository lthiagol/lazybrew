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
