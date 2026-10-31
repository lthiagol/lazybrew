# Review brief — slice `lb07-bound-remaining-reads`

> **Contract.** The slice contract is `work/briefs/lb07-bound-remaining-reads.md`.
> This brief tells you *how to verify* it. The contract wins where the two
> differ.

## What the runner claims

The runner lane bounded the last two unbounded brew reads. Its commits:

| AC | commit | subject |
|---|---|---|
| AC-01 | `bd6dfd7` | fix(gui): bound the uninstall dependency check and surface its timeout |
| AC-02 | `b5c1cc5` | fix(gui): bound the toggleLeaves read with the brew read timeout |
| AC-03 | `ea71699` | test(gui): cover the bounded uninstall check and toggleLeaves timeout |
| AC-04 | `7b21fc0` | docs(work): close LB-07 and record the lb07 slice in the state record |

Its receipt is `work/inbox/runner-lb07-bound-remaining-reads.md`, committed
**three times** (`275f2df`, then `85b83dd` "tighten", then `2ddef85` "reflow")
while the runner brought it under the brief's 35-line limit. Note it as a
process observation; **history is not rewritten**.

The runner claims all four gate commands passed and that its own mutations
showed the three new tests bite. **A receipt is a claim, not evidence.** Run the
gate yourself and reproduce the mutations.

## What this slice must guarantee

| id | guarantee |
|---|---|
| G-1 | both reads go through the existing `boundedRead` with the existing `brewReadTimeout` (30s); **no new timeout var**, no changed value, no change to the detection rule |
| G-2 | a timed-out uninstall dependency check must **not** look like "no dependents": the dialog still opens, but its message says the check did not complete and the list may be incomplete. Any **other** error keeps today's behaviour exactly |
| G-3 | `toggleLeaves`'s timeout surfaces through the existing error toast as a `*brew.TimeoutError` — no new plumbing |
| G-4 | the tab-content "Used By" path stays untouched |

## Verify by command

### V1 — G-1: both reads are capped, nothing else changed

```bash
grep -n "Formulae.Uses\|Formulae.Leaves" internal/gui/commands.go
grep -n "brewReadTimeout = \|brewSearchTimeout = \|brewStatusTotalTimeout = " internal/gui/commands.go
grep -n -A8 "func timeoutError" internal/gui/commands.go
```

Both reads must sit inside a `boundedRead` call with the command labels
`"uses"` and `"leaves"`. The three timeout values must be unchanged (30s /
15s / 45s) and the detection rule must still be `ctx.Err() == context.DeadlineExceeded`.

### V2 — G-2: the uninstall dialog is honest about a timed-out check

This is the point of the slice. Read `confirmUninstall` in
`internal/gui/commands.go` and confirm:

1. the read's error is **no longer discarded**;
2. on a `*brew.TimeoutError`, the confirm dialog's message gains a notice that
   names the timeout duration and says the dependent list may be incomplete;
3. on **any other error**, the behaviour is exactly as before — a dialog with no
   dependents listed and **no** complaint;
4. the notice is appended **inside** the closure (the message is built before
   the command runs, the read happens inside it);
5. the dialog opens in every case — a timeout must not swallow the flow.

### V3 — G-3: `toggleLeaves` surfaces the timeout

```bash
grep -n -A12 "func (m Model) toggleLeaves" internal/gui/commands.go
```

The read must be bounded, and the error must flow through the **existing**
`MutationResultMsg` path with no new message type and no new toast. The user
should see `leaves: brew leaves timed out after 30s` through the existing
renderer at `internal/gui/gui.go:382-383`.

### V4 — AC-03: the three tests exist and bite

`internal/gui/fetch_timeouts_leaves_test.go`:

1. the uninstall check timing out — a `DepCheckMsg` is still produced, its
   `Message` contains the timeout notice and the duration, and it does **not**
   present an empty dependent list as if nothing depended on the package;
2. the uninstall check failing with a **non-timeout** error — a `DepCheckMsg`
   is produced with no dependents and **no** timeout notice;
3. `toggleLeaves` timing out — the `MutationResultMsg` error is a
   `*brew.TimeoutError` (via `errors.As`) naming `leaves` and the duration.

Every wait must be bounded, and the timeout cap must be overridden to
something small for the duration of the test and restored afterwards.

### V5 — G-4 and scope

```bash
git diff --name-only 4b90480..7b21fc0
# must be internal/gui/commands.go, the new test file, work/backlog.md, work/STATE.md

git diff --name-only 4b90480..7b21fc0 -- internal/brew/ internal/gui/gui.go internal/gui/presentation/ internal/gui/modal/ internal/gui/task/ internal/gui/gui_test.go go.mod go.sum .github/ Makefile scripts/
# must print nothing

diff <(git show 4b90480:internal/gui/commands.go | sed -n '/func fetchTabContentCmd/,/^}/p') \
     <(git show HEAD:internal/gui/commands.go | sed -n '/func fetchTabContentCmd/,/^}/p')
# must print nothing — the tab-content path is untouched

# the dashboard fan-out is unchanged
grep -n -A6 "func statusDashboard" internal/gui/commands.go | head -12
```

And the remaining bare `context.Background()` calls must be **exactly** the
four allowed: the bounded-read helper, the task-manager cancel context, the
tab-content timeout, and the status dashboard's total budget.

```bash
grep -n "context.Background()" internal/gui/commands.go
```

### V6 — the success paths are unchanged

- the uninstall dialog still lists the same dependents, with the same wording,
  when the read succeeds;
- `toggleLeaves` still filters the panel the same way and still toggles back
  the same way.

### V7 — the gate, run by you

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Report the real output.

### V8 — AC-04: the bookkeeping

```bash
git show 7b21fc0 -- work/backlog.md work/STATE.md
```

`LB-07` must be `closed` with a note; no other backlog row changed; the state
row must match the contract.

## Mutations to run and revert

Prove the three new tests bite. **Revert every mutation**; `git status` must be
clean when you report.

- **N1 (G-2's notice)** — remove the timeout notice from the uninstall
  confirmation message. Run the uninstall timeout test. It must **fail**.
  Revert.
- **N2 (G-1, uninstall read)** — revert the uninstall read to a bare
  `context.Background()`. Run the uninstall timeout test. It must **fail** (the
  bounded wait must catch the hang). Revert.
- **N3 (G-1, leaves read)** — revert `toggleLeaves` to a bare
  `context.Background()`. Run the `toggleLeaves` timeout test. It must
  **fail**. Revert.
- **N4 (G-2's discrimination)** — make the notice fire on **any** error, not
  only a timeout. Run the non-timeout-error test. It must **fail**. Revert.

A mutation that does not make its test fail is a finding: the test is vacuous.

## Open questions

Read `work/questions.md`. Report whether anything in it was answered, missed,
or newly raised by this slice.

## Scope you do not own

- You do not fix anything.
- You mutate only to prove a check bites, and you revert it.
- You do not rewrite the implementation or the tests to be nicer.
- You do not touch `work/STATE.md` beyond confirming it, `internal/brew/**`,
  `internal/gui/gui.go`, the tab-content path, or the contents of
  `master-plan/`.
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

1. Write `work/inbox/reviewer-lb07-bound-remaining-reads.md` — concise
   markdown, at most 40 lines: the verdict, the findings, the evidence
   (command → result) for V1–V8, the outcome of each mutation N1–N4, and the
   state of the questions.
2. Commit it:
   `git add work/inbox/reviewer-lb07-bound-remaining-reads.md && git commit -m "inbox: reviewer reports lb07-bound-remaining-reads"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `reviewer finished lb07-bound-remaining-reads; verdict in work/inbox/reviewer-lb07-bound-remaining-reads.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.
