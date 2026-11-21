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
| **Gate** | `make ci` (gofmt check, vet, `go test ./... -race -count=1`, `make cover-check`, build) | Must pass before a lane reports. Per commit, `make test-fast PKG=<touched packages>` is enough (`Q-v2-builder-speed`). |
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

## Running the flow with herdr (orchestrator runbook)

How the LB-08 run (v2-01 … v2-11) was driven; a new orchestrator session
should start here, then read `work/STATE.md` (bottom = latest) and the open
items in `work/questions.md` and `work/backlog.md`.

**Agents.** One builder and one reviewer, each in its own herdr tab in the
orchestrator's workspace. A **fresh pair per slice** (stop both, start new) so
each slice starts with clean context.

**Names.** Other projects run builders and reviewers on the same herdr server, so
agent names and tab labels carry the `lb-` slug: tabs `lb-builder` and
`lb-reviewer`, agents `lb-builder` and `lb-reviewer`. Never address a bare
`builder` or `reviewer`. Exit an agent with `herdr agent prompt <name> "/exit"`
(a double Ctrl-C can open an exit dialog); the pane returns to a shell.

| Lane | Model | Effort | Start |
|---|---|---|---|
| builder (runner) | `claude-haiku-5-5` | high | `herdr agent start lb-builder --kind claude --pane <builder pane> -- --model claude-haiku-5-5 --effort high --permission-mode auto` |
| reviewer | `claude-opus-5-5` | medium | `herdr agent start lb-reviewer --kind claude --pane <reviewer pane> -- --model claude-opus-5-5 --effort medium --permission-mode auto` |

**Loop per slice.**
1. Orchestrator writes `work/briefs/<slice>.md` and `review-<slice>.md`
   (copy the last pair as a template — check every path names the new slice),
   commits them.
2. `herdr agent prompt lb-builder "<read AGENTS.md + work/README.md, implement the brief, reply with the done line>"`,
   then `herdr agent wait lb-builder --timeout <ms>` — **the timeout is in
   milliseconds** (3600000 = 1 h).
3. Read the receipt; rule on any question it raised (append to
   `questions.md`) before the review.
4. Prompt the reviewer for round rN; read `work/inbox/reviewer-<slice>-rN.md`.
5. `changes-required` → write `work/briefs/fix-<slice>-rN.md` listing every
   finding, prompt the builder, then the next review round.
   **At most 5 review rounds**; the builder fixes every finding. If r5 is
   still not ok, stop that slice, record it, and move on (user ruling
   `Q-v2-run`).
6. `ok` → update the slice row and a short paragraph in `STATE.md`, commit,
   **push `wip`** (authorised after each closed slice), check `wip-ci`
   (`gh run list --branch wip --limit 1`), stop both agents
   (`herdr agent send-keys <name> C-c C-c`).

**Standing rules.**
- Never run brew write commands on this machine (`moshi-hook` is a live
  service); tests use the mock runner or fake brew scripts. Harness frames
  may open a confirm only to capture it, then `esc`.
- Never push to `stable`. The `wip` → `stable` PR needs the user's go-ahead,
  after the human smoke gate (`SMOKE-CHECKLIST.md`, LB-05).
- Put decisions to the user with a recommended option marked.
- If brew reads are slow (network), frames show fallbacks — re-capture when
  brew is back to normal before reviewing (`Q-v2-08-b`).
- Claude Code's greyed-out "next prompt" suggestion in an agent pane is not
  typed input; it is never sent.
