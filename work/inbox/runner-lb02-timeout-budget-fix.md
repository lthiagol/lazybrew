# runner receipt — lb02-timeout-budget-fix (F-01)

Two commits, one per acceptance criterion, in AC order:

1. `8647165` — fix(gui): clamp negative budget and skip reads with a spent parent (AC-01)
2. `0cfb213` — test(gui): prove the timeout budget is never negative (AC-02)

## AC-01 — the reported budget is never negative

`internal/gui/commands.go`, `boundedReadIn`: the parent's remaining time is now
computed into `remaining` before the budget is trimmed, the budget is clamped to
zero, and when the parent's deadline has already passed (`remaining <= 0`) the
read is **not started at all** — `read(ctx)` is never invoked, so the brew client
is never called with a dead context. The function returns a `*brew.TimeoutError`
naming the command. `timeoutError`, `boundedRead`'s signature, the timeout
values, the detection rule (deadline read from `ctx.Err()`) and the tab-content
path are untouched; the fan-out shape is unchanged.

Behaviour chosen for the already-spent parent: the message reports a duration
of **0** — `brew list timed out after 0s` — because no budget remained. The
duration is never negative.

## AC-02 — a test proves the message is never negative

`internal/gui/fetch_timeouts_budget_test.go` —
`TestBoundedReadInSpentParentBudgetIsNeverNegative`: parent context with a
deadline already in the past, bounded wait (2s) around the call, `errors.As` on
`*brew.TimeoutError`, `Timeout >= 0`, the message checked against a
negative-duration regexp, and an `atomic.Bool` proving the read function was
never called. The test fails against the pre-fix code (the read was invoked and
returned no error at all).

## Gate (real results)

- `make build` → ok (`go build -o bin/lazybrew ./cmd/lazybrew`)
- `make lint` → ok (`go vet ./...`)
- `go test ./... -race -count=1` → ok, all packages pass
- `make cover-check` → ok (brew 64.3% ≥ 62%, gui 57.5% ≥ 36%, gui/modal 41.4% ≥ 41%, gui/presentation 95.4% ≥ 91%)

Evidence greps: `go test ./internal/gui/ -run 'Budget|Negative|Expired|SpentParent' -count=1 -v` → PASS;
`go test ./internal/gui/ -run 'TimesOut|RespectsTotalBudget|TotalBudgetBoundsTheAggregate' -count=1` → PASS.
