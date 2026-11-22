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
| **Brief** | `work/briefs/<slice>.md` | The contract for a slice. Acceptance criteria, scope, invariants, and the report protocol. It wins over any other plan, note, or prior convention, but never over the hard rules in `AGENTS.md`; a brief that conflicts with one is a question for `work/questions.md`. |
| **Receipt** | `work/inbox/<lane>-<slice>.md` | What a lane did, with the evidence. Written, committed, then reported. |
| **Verification** | `work/verification/<slice>.md` | The evidence record of a verification slice: the invariants checked, the commands run, and their real output. |
| **Gate** | `make ci`, five steps (gofmt check, vet, `go test ./... -race -count=1`, `make cover-check`, build) | Must pass before a lane reports. Per commit, `make test-fast PKG=<touched packages>` is enough (`Q-v2-builder-speed`). |
| **State row** | `work/STATE.md` | The "▶ Resume here" block at the top, then one row per slice (the commits, the state, the verdict) and the slice paragraphs below. The resume record. |
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
2. **The gate must pass before a lane reports.** `make ci`, real output, no
   papering over a failure. Each commit must build and pass
   `make test-fast PKG=<touched packages>`; the full race gate runs once,
   before the receipt.
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
     |                  gate: make ci (five           verdict per AC
     |                  steps)                            |
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
- the `make ci` result (gofmt check, vet, race tests, cover-check, build), pass or fail;
- what the next session must know: any command that failed, any contradiction
  found between the brief and the repo.

## Running the flow with herdr (orchestrator runbook)

How the LB-08 run (v2-01 … v2-11) was driven; a new orchestrator session
should start here, then read the "▶ Resume here" block at the top of
`work/STATE.md` (slice rows and paragraphs follow it) and the open items in
`work/questions.md` and `work/backlog.md`.

**Agents.** One builder and one reviewer, each in its own herdr tab in the
orchestrator's workspace. Start a **fresh builder for every round** (the build,
then each fix round; owner ruling) and a fresh reviewer per slice, so no lane
carries old context. A fresh builder knows nothing of the slice: its prompt names
the slice brief, the fix brief, and that the slice's earlier commits are on `wip`.

**Names.** Other projects run builders and reviewers on the same herdr server, so
agent names and tab labels carry the `lb-` slug: tabs `lb-builder` and
`lb-reviewer`, agents `lb-builder` and `lb-reviewer`. Never address a bare
`builder` or `reviewer`. Stop an agent with `herdr agent prompt <name> "/exit"`
(not Ctrl-C: a double Ctrl-C can open an exit dialog); the pane returns to a shell.

| Lane | Model | Effort | Start |
|---|---|---|---|
| builder (runner) | `claude-haiku-5-5` | high | `herdr agent start lb-builder --kind claude --pane <builder pane> -- --model claude-haiku-5-5 --effort high --permission-mode auto` |
| reviewer | `claude-opus-5-5` | medium | `herdr agent start lb-reviewer --kind claude --pane <reviewer pane> -- --model claude-opus-5-5 --effort medium --permission-mode auto` |

**Loop per slice.**
1. Orchestrator writes `work/briefs/<slice>.md` and `review-<slice>.md`
   (copy the last pair as a template — check every path names the new slice),
   commits them. Do not carry over a frame harness built on the real runner
   (`brew.NewDefaultRunner()`, as in `work/verification/v2-*/`): a brief that asks
   for frames asks for a harness on `brewtest.NewMockRunner` (AGENTS.md rule 1).
2. `herdr agent prompt lb-builder "<read AGENTS.md + work/README.md, implement the brief, reply with the done line>"`,
   then `herdr agent wait lb-builder --timeout <ms>` — **the timeout is in
   milliseconds** (3600000 = 1 h).
3. Read the receipt; rule on any question it raised (append to
   `questions.md`) before the review.
4. Prompt the reviewer for round rN; read `work/inbox/reviewer-<slice>-rN.md`.
5. `changes-required` → write `work/briefs/fix-<slice>-rN.md` listing every
   finding, stop the builder and start a fresh one, prompt it, then the next
   review round.
   **At most 5 review rounds**; the builder fixes every finding. If r5 is
   still not ok, stop that slice, record it, and move on (user ruling
   `Q-v2-run`).
6. `ok` → update the slice row and a short paragraph in `STATE.md`, commit,
   **push `wip`** (authorised after each closed slice), check `wip-ci`
   (`gh run list --branch wip --limit 1`), stop both agents
   (`herdr agent prompt <name> "/exit"`).

**Standing rules.**
- Never run brew write commands on this machine (`moshi-hook` is a live
  service); tests and frame harnesses use the mock runner or fake brew scripts,
  never the real runner. Harness frames may open a confirm only to capture it,
  then `esc`.
- Never write to `stable` unless the user asks. Moving it needs the human smoke
  gate (`SMOKE-CHECKLIST.md`) and the steps in `docs/agents/release.md`.
- Put decisions to the user with a recommended option marked.
- If brew reads are slow (network), frames show fallbacks — re-capture when
  brew is back to normal before reviewing (`Q-v2-08-b`).
- Claude Code's greyed-out "next prompt" suggestion in an agent pane is not
  typed input; it is never sent.

## Running the flow with Claude Code subagents (claude-crew)

Outside herdr, the same loop runs with subagents (the `claude-crew` skill): a
fresh Haiku builder per round, one Opus reviewer per unit, at most 5 rounds. The
briefs, reports and reviews live in a scratch round folder instead of
`work/briefs` and `work/inbox`; the rules do not change — the gate passes, files
are staged by name, one commit per unit (or per AC), and the slice gets its row
in `work/STATE.md`.
