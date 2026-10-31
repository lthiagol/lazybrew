# The Lane Flow

Planning for `lazybrew` runs through a **lane flow**: an orchestrator writes a
brief, a runner lane implements it against the brief, a reviewer lane checks it
against the brief, and the slice's state row records the outcome.

This project does **not** use the `mp` CLI. `master-plan/` is a **frozen local
archive** — untracked by git, kept on this machine for history only, and not an
input to any decision. Do not read it, do not update it, do not run `mp` here.

## The pieces

| Piece | Where | What it is |
|---|---|---|
| **Brief** | `work/briefs/<slice>.md` | The contract for a slice. Acceptance criteria, scope, invariants, and the report protocol. It wins over any other plan, note, or prior convention. |
| **Receipt** | `work/inbox/<lane>-<slice>.md` | What a lane did, with the evidence. Written, committed, then reported. |
| **Verification** | `work/verification/<slice>.md` | The evidence record of a verification slice: the invariants checked, the commands run, and their real output. |
| **Gate** | `make build`, `make lint`, `go test ./... -race -count=1`, `make cover-check` | The commands that must pass before a lane reports. |
| **State row** | `work/STATE.md` | One row per slice: the commits, the state, the verdict. The resume record. |
| **Backlog** | `work/backlog.md` | Parked work that is not in any active slice. |
| **Questions** | `work/questions.md` | Open questions. Appended to, never rewritten. |

## Who owns what

| Role | Owns |
|---|---|
| **Orchestrator** | The brief, the slice order, the open questions, and the nudge. |
| **Runner lane** | The implementation, one commit per acceptance criterion, and the receipt. |
| **Reviewer lane** | The check against the brief, and the verdict on each acceptance criterion. |
| **Human** | The release smoke gate and anything that needs live brew writes. |

## The rules

1. **One commit per acceptance criterion**, in AC order. Stage by name — never
   `git add -A`, never `git add .`.
2. **The gate must pass before a lane reports.** All four commands, real output,
   no papering over a failure.
3. **A lane reports by writing `work/inbox/<lane>-<slice>.md`, committing it,
   then nudging the orchestrator.** The file is the record; the nudge is the bell.
4. **The slice's row in `work/STATE.md` is the resume record** and is updated in
   the same session as the slice.
5. **A question that is not written in `work/questions.md` is a lost question.**
   Append it and keep going on the criteria it does not affect.
6. **A blocked lane reports.** Write the receipt with the blocker named and
   nudge anyway. A blocked lane that reports beats a lane that goes quiet.

## The flow

```
orchestrator            runner lane                reviewer lane
     |                       |                           |
brief in work/briefs/  ->  read the brief          ->   read the brief
     |                       |                           |
     |                  implement, one commit          check each AC
     |                  per AC, in order                   |
     |                       |                           |
     |                  gate: build, lint,            verdict per AC
     |                  test -race, cover-check           |
     |                       |                           |
     |                  receipt in work/inbox/  ------>  read the receipt
     |                       |                           |
     |                  nudge orchestrator      ->   state row updated
     |
state row in work/STATE.md updated
```

## Receipt shape

A receipt is at most 40 lines of markdown and answers:

- the commit ids, one per acceptance criterion, in order;
- each AC answered with the evidence that proves it — the command and its output,
  not a restatement of the intent;
- the gate result for all four commands, pass or fail;
- what the next session must know: any command that failed, any contradiction
  found between the brief and the repo.
