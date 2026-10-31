# Slice: LB-01 — verify the M13 Deps-tab fix

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Close `LB-01` from `work/backlog.md` with evidence.

`LB-01` is: *verify the M13 Deps-tab fix by running the Deps test suite.* The
fix itself landed in commit `ac5bac6` ("fix(gui): stop Formulae Deps infinite
loading (M13)") and its regression tests are in the tree. What never happened
is the **verification**: the milestone was executed on 2026-08-02, its findings
were fixed, and the independent review that was supposed to confirm it was
never run. This slice is that verification — and only that.

**This slice changes documentation. It changes no code.** If you find yourself
editing a `.go` file, you are out of scope.

## The five invariants of the M13 fix

The fix commit lists five changes. Each must still be present in the code, and
the reviewer will later mutate each one to prove the tests bite. These are the
exact anchors:

| id | invariant | anchor |
|---|---|---|
| I-1 | A tab-content fetch is bounded by a context timeout | `internal/gui/commands.go:22` (`fetchTabContentTimeout`), `internal/gui/commands.go:833` (`context.WithTimeout`) |
| I-2 | After `DataLoadedMsg`, the tab content load is re-kicked | `internal/gui/gui.go:455-456` |
| I-3 | The Deps tab fast-paths from the cached formula's dependencies instead of calling brew | `internal/gui/commands.go:779` (call site), `:794` (`renderCachedTabContent`), `:810` (`runtime := f.Dependencies`) |
| I-4 | `parseFormula` maps build deps from `build_dependencies`, not `dependencies` | `internal/brew/formulae.go:318` (`buildDeps := append([]string(nil), f.BuildDependencies...)`), field at `:64` |
| I-5 | The tab/panel key handlers return early so the fetch command is not dropped | `internal/gui/gui.go:627-634` |

The regression tests that must cover them:

- `internal/gui/gui_test.go` — `TestDepsTabLoads`, `TestDepsSelectionChangeUpdatesContent`
- `internal/brew/formulae_test.go` — `TestParseFormulaBuildDepsFromBuildDependencies`, and the
  dependency assertions inside `TestFormulaeReaderList`

## Acceptance criteria

One commit per criterion, in order. Stage by name — never `git add -A`, never
`git add .`.

### AC-01 — the verification record exists and is complete

Create `work/verification/lb01-m13-deps.md` (new directory `work/verification/`).
It must contain, for each of I-1..I-5: the invariant, the **command that
proves it is present**, the command's **actual output**, and the line it was
found at. Then the test-run output and the gate output.

In the same commit, add `work/verification/` to the file-role table in
`work/README.md` — one row, the same shape as the existing rows. Nothing else
in `work/README.md` changes.

### AC-02 — the backlog item is closed with its evidence

In `work/backlog.md`:

- set `LB-01`'s `status` cell to `closed`;
- add a one-line note under the table recording that the M13 review is closed
  by this verification, naming the evidence file
  (`work/verification/lb01-m13-deps.md`) and the commit that carries it;
- `LB-06` stays `closed-by-LB-01` — do not touch it.

No other row changes. `LB-02`, `LB-03`, `LB-04`, `LB-05` and `LB-DROP-01`
stay exactly as they are.

### AC-03 — the slice has a row in the resume record

In `work/STATE.md`, add a row for this slice:

```
| lb01-m13-deps-verify | — | implemented | awaiting reviewer |
```

Leave the `detach-master-plan` row untouched. The `—` in the commits column is
correct for now: the orchestrator fills it in when the slice closes.

## Invariants

- **No code changes.** Nothing under `internal/`, `cmd/`, `testdata/`,
  `scripts/`, `.github/`, `Makefile`, `go.mod`, `go.sum` is touched. No test
  is touched.
- **Nothing under `master-plan/` is read, modified, moved, or deleted.** It is
  a frozen local archive.
- The only files changed are: `work/verification/lb01-m13-deps.md` (new),
  `work/README.md`, `work/backlog.md`, `work/STATE.md`.
- One commit per acceptance criterion, in AC order, staged by name.
- Do not amend, rebase, or revert existing commits — including `ac5bac6`.
- Do not edit `work/briefs/**` or `work/inbox/**`.
- The worktree must be clean when you report.

## Scope — do NOT touch

Everything not named in an acceptance criterion.

## Evidence to produce

Run these and record the **real output** of each in
`work/verification/lb01-m13-deps.md`:

```bash
# the five invariants, one command each
grep -n "fetchTabContentTimeout" internal/gui/commands.go
sed -n '455,456p' internal/gui/gui.go
grep -n "renderCachedTabContent" internal/gui/commands.go
grep -n "buildDeps := append" internal/brew/formulae.go
sed -n '627,634p' internal/gui/gui.go

# the regression tests
go test ./internal/gui/ -run 'TestDepsTabLoads|TestDepsSelectionChangeUpdatesContent' -count=1 -v
go test ./internal/brew/ -run 'TestParseFormulaBuildDepsFromBuildDependencies|TestFormulaeReaderList' -count=1 -v

# the gate
make build
make lint
go test ./... -race -count=1
make cover-check
```

If any command fails, say so plainly in the verification file and in the
receipt. A failing check is information; a hidden failure is a defect.

## Report protocol

1. Write `work/inbox/runner-lb01-m13-deps-verify.md` — concise markdown, at
   most 30 lines: the three commit ids in order, each AC answered with the
   evidence that proves it, the gate result, and anything the next session
   must know — including any check that failed.
2. Commit it:
   `git add work/inbox/runner-lb01-m13-deps-verify.md && git commit -m "inbox: runner reports lb01-m13-deps-verify"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `runner finished lb01-m13-deps-verify; receipt at work/inbox/runner-lb01-m13-deps-verify.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.

## Questions

If something in this brief is wrong or ambiguous, append it to
`work/questions.md` (append, do not rewrite) and keep going on the criteria
that are unaffected. The orchestrator owns the open questions; do not answer
them yourself.
