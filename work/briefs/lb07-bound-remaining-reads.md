# Slice: LB-07 — bound the last two unbounded brew reads

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

`LB-07` from `work/backlog.md`: *apply context timeouts to the two brew reads
still unbounded*.

`lb02-fetch-timeouts` bounded the four panel/search/dashboard read paths. Two
brew reads were left behind, and they are the last two places in the GUI where
a wedged `brew` invocation can still hang the app forever:

| read | location | the unbounded call |
|---|---|---|
| uninstall dependency check | `internal/gui/commands.go:1053-1054` | `m.client.Formulae.Uses(ctx, name)` — the error is **discarded** (`uses, _ :=`) |
| `toggleLeaves` | `internal/gui/commands.go:1145-1148` | `m.client.Formulae.Leaves(ctx)` |

This slice bounds both, using the machinery `lb02-fetch-timeouts` already
built. It does not add a new mechanism.

## The guarantees

### G-1 — both reads are capped with the existing helper

Both must go through `boundedRead` (`internal/gui/commands.go:71`) with the
existing `brewReadTimeout` (30s). **Do not add a new timeout var** — these are
ordinary brew reads, and `brewReadTimeout` is the cap for those. The cap
values do not change: `brewReadTimeout` 30s, `brewSearchTimeout` 15s,
`brewStatusTotalTimeout` 45s.

The detection rule does not change either: a deadline overrun is detected from
the **context's own error**, inside `boundedRead`, never from the type of the
error the brew layer returned.

### G-2 — a timed-out dependency check must not look like "no dependents"

This is the point of the slice, not a detail.

Today the uninstall check discards its error, so a hung `brew uses` produces a
confirmation dialog that lists **no dependents** — indistinguishable from
"nothing depends on this package". The user then confirms an uninstall that
may break other packages, without ever being told the check did not run.

So:

- On a **timeout**, the confirmation dialog must still appear — the user must
  keep the choice to confirm or cancel — but its message must say the
  dependency check did not complete and that the list of dependents may be
  incomplete. Name the timeout duration. Something in this shape:
  *"Dependency check timed out after 30s — the list of packages that depend on
  this one may be incomplete."*
- On **any other error** (brew missing, tap unreadable, …), keep today's
  behaviour exactly: no dependents listed, no complaint. Do not change
  unrelated behaviour.
- The message is built before the command runs but the read happens inside the
  command, so the notice must be appended **inside** the closure.

### G-3 — `toggleLeaves`'s timeout surfaces through the existing toast

`toggleLeaves` already propagates its error as
`MutationResultMsg{…, Err: err}`, and `internal/gui/gui.go:382-383` renders it
as an error toast: `msg.Name+": "+msg.Err.Error()`. With `boundedRead`, a
timeout becomes a `*brew.TimeoutError`, so the user sees:

```
leaves: brew leaves timed out after 30s
```

**No new plumbing, no new toast, no new message type.** The existing path is
correct; it only needs the error to be the right type.

### G-4 — out of scope

`internal/gui/commands.go:971` (`Formulae.Uses` for the "Used By" tab) is
already bounded by the tab-content timeout. **Do not touch it.** Nor the
`context.WithCancel` at `:761` (a task-manager context, not a brew read), nor
the tab-content path.

## Acceptance criteria

One commit per criterion, in order. Stage by name — never `git add -A`, never
`git add .`.

### AC-01 — the uninstall dependency check is bounded and honest

In `confirmUninstall` (`internal/gui/commands.go`):

1. the read goes through `boundedRead(brewReadTimeout, "uses", …)` — the command
   label is `"uses"`;
2. on a timeout the confirm dialog's message gains the notice described in
   G-2 (inside the closure, after the read);
3. on any other error the behaviour is unchanged — no dependents listed, no
   complaint;
4. when the read succeeds, the dialog lists the dependents exactly as today.

The dialog must still open in every case. A timeout must not cancel the
uninstall flow, and it must not leave the UI without a modal.

### AC-02 — `toggleLeaves` is bounded

In `toggleLeaves` (`internal/gui/commands.go`):

1. the read goes through `boundedRead(brewReadTimeout, "leaves", …)` — the
   command label is `"leaves"`;
2. the existing `MutationResultMsg` error path is untouched, so a timeout
   reaches the user as an error toast naming the command and the duration;
3. a successful read behaves exactly as today.

### AC-03 — tests for both

Add tests (new file `internal/gui/fetch_timeouts_leaves_test.go`, or beside
the existing timeout tests) covering:

1. **the uninstall check timing out** — the mock runner blocks on `ctx.Done()`;
   the test asserts that a `DepCheckMsg` is still produced, that its `Message`
   contains the timeout notice (and the duration), and that it does **not**
   present an empty dependent list as if nothing depended on the package;
2. **the uninstall check failing with a non-timeout error** — the test asserts
   today's behaviour is preserved: a `DepCheckMsg` is produced, with no
   dependents and **no** timeout notice;
3. **`toggleLeaves` timing out** — the test asserts the returned
   `MutationResultMsg` carries an error that is a `*brew.TimeoutError`
   (via `errors.As`) naming `leaves` and the duration.

Use the established pattern: `brew.NewMockRunner()` with an `ExecuteFn` that
blocks on `ctx.Done()`, and **override `brewReadTimeout` to something small**
(50ms) for the duration of the test, restoring it afterwards — the same shape
as `internal/gui/fetch_timeouts_test.go`. Every wait must be bounded; a test
that can hang is worse than no test.

### AC-04 — the slice closes its own bookkeeping

1. In `work/backlog.md`: set `LB-07`'s `status` cell to `closed`, and add a
   one-line note under the table naming this slice. **Do not change any other
   row.**
2. In `work/STATE.md`: add a row for this slice —
   `| lb07-bound-remaining-reads | — | implemented | awaiting reviewer |`. The
   `—` is correct for now; the orchestrator fills it in when the slice closes.

## Invariants

- **No new timeout var**, no change to any timeout value.
- **No change to the detection rule**, to `boundedRead`, to `boundedReadIn`,
  or to `timeoutError`.
- **No change to the fan-out shape** of the status dashboard.
- **No change to the tab-content path** (G-4).
- The uninstall flow's **success path is unchanged**: same modal, same
  dependents list, same wording.
- `toggleLeaves`' success path is unchanged: same panel filtering, same
  toggle behaviour.
- No changes to `internal/brew/**`, `internal/gui/gui.go`,
  `internal/gui/presentation/**`, `internal/gui/modal/**`, `internal/gui/task/**`,
  `Makefile`, `.github/workflows/**`, `scripts/**`, `go.mod`, `go.sum`.
- The only files changed: `internal/gui/commands.go`, the test file(s) you add
  under `internal/gui/`, and the two `work/` files named in AC-04.
- One commit per acceptance criterion, in AC order, staged by name.
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

All four must pass. Plus:

```bash
# both reads now go through the helper
grep -n "Formulae.Uses\|Formulae.Leaves" internal/gui/commands.go
# every remaining hit must be inside a boundedRead call (or the already-bounded
# tab-content path at :971)

# no bare context.Background() is left in these two paths
grep -n "context.Background()" internal/gui/commands.go
# the only hits must be: the bounded-read helper, the tab-content path, the
# task-manager cancel context, and the status dashboard's total budget
```

Record the real output of each in your receipt. If a check fails, report it.

## Report protocol

1. Write `work/inbox/runner-lb07-bound-remaining-reads.md` — concise markdown,
   at most 35 lines: the four commit ids in order, each AC answered with the
   evidence that proves it, the gate result, the exact wording you used for the
   timeout notice in the uninstall dialog, and anything the next session must
   know.
2. Commit it:
   `git add work/inbox/runner-lb07-bound-remaining-reads.md && git commit -m "inbox: runner reports lb07-bound-remaining-reads"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `runner finished lb07-bound-remaining-reads; receipt at work/inbox/runner-lb07-bound-remaining-reads.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.

## Questions

If something in this brief is wrong or ambiguous, append it to
`work/questions.md` (append, do not rewrite) and keep going on the criteria
that are unaffected. The orchestrator owns the open questions; do not answer
them yourself.
