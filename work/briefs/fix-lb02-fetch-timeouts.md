# Fix brief — slice `lb02-fetch-timeouts`, reviewer finding F-01

> **Contract.** The slice contract is `work/briefs/lb02-fetch-timeouts.md`. This
> brief names the one finding it must fix. Where the two differ, this brief wins
> for the fix cycle only.

## Context

The reviewer lane verified `lb02-fetch-timeouts` and returned
**`Verdict: ok`** in `work/inbox/reviewer-lb02-fetch-timeouts.md` (commit
`633dc50`). It reproduced every mutation, confirmed the gate, and judged the
concurrency decision sound.

Two findings, both minor:

- **F-01** — `boundedReadIn` can report a **negative** duration when the
  parent's budget is already spent.
- **F-02** — the fan-out puts both heavy `brew info --json=v2` loads and
  `brew doctor` in the same instant. The reviewer judged it acceptable and
  correct; it is a shape note, **not** a defect.

**F-02 is accepted and closed. Do not change the fan-out shape.** The
orchestrator records it in the state record.

**F-01 is the only thing this cycle fixes.**

## The finding

`internal/gui/commands.go`, in `boundedReadIn` (~:90):

```go
budget := cap
if parentDeadline, ok := parent.Deadline(); ok {
    if remaining := time.Until(parentDeadline); remaining < budget {
        budget = remaining
    }
}
```

When the parent's deadline has already passed, `time.Until(parentDeadline)` is
**negative**, and that negative value becomes the budget reported to the user.
The reviewer's probe produced:

```
brew list timed out after -1s
```

It is latent today — the four paths start all six reads at once, so no read
reaches an already-expired parent — but the path is reachable, and a message
that says the read timed out "after -1s" is wrong in a way a user can see.

## Acceptance criteria

One commit per criterion, in order. Stage by name — never `git add -A`, never
`git add .`.

### AC-01 — the reported budget is never negative

In `boundedReadIn` (`internal/gui/commands.go`):

1. The budget reported to the user must never be negative. Clamp it.
2. When the parent's budget is **already spent**, do not start the read at
   all: the read has no chance, so return the timeout error immediately
   instead of invoking the brew client with a dead context. The error must
   still be a `*brew.TimeoutError` naming the brew command and a
   **non-negative** duration.
3. The clamp must not change the behaviour of any path that currently works:
   a read that starts with budget remaining behaves exactly as before, and a
   read whose parent expires *during* the read still reports the trimmed
   remaining budget as it does today.

### AC-02 — a test proves the message is never negative

Add a test to `internal/gui/fetch_timeouts_test.go` (or a new test file beside
it) that:

1. builds a **parent context whose deadline has already passed**, and calls
   `boundedReadIn` with it;
2. asserts the call returns promptly;
3. asserts the error is a `*brew.TimeoutError` (via `errors.As`);
4. asserts the message contains **no negative duration** — i.e. it must not
   match a pattern like `-1s` / `-500ms`, and the duration it reports must be
   `>= 0`;
5. asserts the read function was **not** called (the parent was already
   spent — the brew client must not be invoked at all).

The test's own waits must be bounded, like the existing ones (`awaitMsg`).

## Invariants

- **Exactly two commits**, one per criterion, in AC-01 → AC-02 order.
- **No change to the fan-out shape.** The six dashboard reads stay concurrent
  under the total budget (F-02 is accepted).
- **No change to the timeout values**: `brewReadTimeout` 30s,
  `brewSearchTimeout` 15s, `brewStatusTotalTimeout` 45s.
- **No change to the detection rule**: the deadline is still detected from
  `ctx.Err()`, never from the returned error's type.
- **No change to `timeoutError`, `boundedRead`'s signature, or any message
  plumbing.**
- The tab-content path stays untouched.
- The only files changed: `internal/gui/commands.go` and the test file.
- No change to `internal/brew/**`, `go.mod`, `go.sum`, `Makefile`, `.github/`,
  `scripts/`, `work/backlog.md`, `work/STATE.md`, the briefs, or any receipt.
- Do not amend, rebase, or revert existing commits.
- `master-plan/` is a frozen local archive — do not read or modify it.
- The worktree must be clean when you report.

## Evidence to produce

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Plus, for the fix:

```bash
# the budget is clamped — no unguarded time.Until in the budget computation
grep -n -A12 "func boundedReadIn" internal/gui/commands.go
# the new test exists and passes
go test ./internal/gui/ -run 'Budget|Negative|Expired' -count=1 -v
# and the existing timeout tests still pass
go test ./internal/gui/ -run 'TimesOut|RespectsTotalBudget|TotalBudgetBoundsTheAggregate' -count=1
```

Record the real output of each in your receipt.

## Report protocol

1. Write `work/inbox/runner-lb02-timeout-budget-fix.md` — concise markdown,
   at most 25 lines: the two commit ids in order, each AC answered with the
   evidence, the gate result, and the exact behaviour you chose for the
   already-spent-parent case (what duration the message reports).
2. Commit it:
   `git add work/inbox/runner-lb02-timeout-budget-fix.md && git commit -m "inbox: runner reports lb02-timeout-budget-fix"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `runner finished the lb02 timeout-budget fix; receipt at work/inbox/runner-lb02-timeout-budget-fix.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.

## Questions

If something in this brief is wrong or ambiguous, append it to
`work/questions.md` (append, do not rewrite) and keep going on the criteria
that are unaffected. The orchestrator owns the open questions; do not answer
them yourself.
